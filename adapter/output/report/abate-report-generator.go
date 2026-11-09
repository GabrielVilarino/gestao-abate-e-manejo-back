package report

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/output"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

const reportGenerationTimeout = 2 * time.Minute

//go:embed abate-report-template.html
var reportTemplateHTML string

//go:embed logo.png
var brandLogoPNG []byte

type ChromedpPDFGenerator struct {
	template     *template.Template
	brandDataURL string
}

type reportTemplateData struct {
	Reports      []output.AbateReport
	BrandDataURL string
}

type photoPage struct {
	Fazenda             []output.AbateReportPhoto
	Frigorifico         []output.AbateReportPhoto
	FazendaContinuation bool
	FrigoContinuation   bool
}

func NewChromedpPDFGenerator() (*ChromedpPDFGenerator, error) {
	functions := template.FuncMap{
		"formatDate":                 func(value time.Time) string { return value.Format("02/01/2006") },
		"formatNumber":               formatNumber,
		"formatPercent":              func(value float64) string { return formatNumber(value, 1) + "%" },
		"formatRatio":                func(value float64) string { return formatNumber(value*100, 1) + "%" },
		"barHeight":                  barHeight,
		"maxDistributionWeight":      maxDistributionWeight,
		"maxDistributionHeads":       maxDistributionHeads,
		"distributionBarHeight":      distributionBarHeight,
		"distributionHeadsBarHeight": distributionHeadsBarHeight,
		"denticaoBarColor":           denticaoBarColor,
		"acabamentoBarColor":         acabamentoBarColor,
		"upper":                      strings.ToUpper,
		"classRange":                 classificationRange,
		"classPenalty":               classificationPenalty,
		"photoPages":                 buildPhotoPages,
		"hasValue":                   func(value float64) bool { return value != 0 },
		"safeURL":                    func(value string) template.URL { return template.URL(value) },
	}
	parsed, err := template.New("abate-report").Funcs(functions).Parse(reportTemplateHTML)
	if err != nil {
		return nil, fmt.Errorf("carregar template do relatório: %w", err)
	}
	return &ChromedpPDFGenerator{
		template:     parsed,
		brandDataURL: "data:image/png;base64," + base64.StdEncoding.EncodeToString(brandLogoPNG),
	}, nil
}

func (g *ChromedpPDFGenerator) Generate(ctx context.Context, reports []output.AbateReport) ([]byte, error) {
	htmlContent, err := g.renderHTML(reports)
	if err != nil {
		return nil, err
	}
	browserCtx, cancelBrowser := chromedp.NewContext(ctx)
	defer cancelBrowser()
	browserCtx, cancelTimeout := context.WithTimeout(browserCtx, reportGenerationTimeout)
	defer cancelTimeout()

	var pdf []byte
	err = chromedp.Run(browserCtx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, htmlContent).Do(ctx)
		}),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Poll(`Array.from(document.images).every(image => image.complete)`, nil),
		chromedp.ActionFunc(func(ctx context.Context) error {
			pdf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithLandscape(true).
				WithPreferCSSPageSize(true).
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("gerar PDF do relatório: %w", err)
	}
	return pdf, nil
}

func (g *ChromedpPDFGenerator) renderHTML(reports []output.AbateReport) (string, error) {
	var rendered bytes.Buffer
	if err := g.template.Execute(&rendered, reportTemplateData{Reports: reports, BrandDataURL: g.brandDataURL}); err != nil {
		return "", fmt.Errorf("renderizar relatório: %w", err)
	}
	return rendered.String(), nil
}

func extractBrandDataURL(referenceHTML string) (string, error) {
	brandIndex := strings.Index(referenceHTML, `class="brand"`)
	if brandIndex < 0 {
		return "", errors.New("marca do template de relatório não encontrada")
	}
	source := referenceHTML[brandIndex:]
	sourceIndex := strings.Index(source, `src="`)
	if sourceIndex < 0 {
		return "", errors.New("imagem da marca do template de relatório não encontrada")
	}
	source = source[sourceIndex+len(`src="`):]
	endIndex := strings.IndexByte(source, '"')
	if endIndex < 0 || !strings.HasPrefix(source[:endIndex], "data:image/") {
		return "", errors.New("imagem da marca do template de relatório inválida")
	}
	return source[:endIndex], nil
}

func formatNumber(value float64, decimals int) string {
	if math.Abs(value) < math.Pow10(-decimals)/2 {
		value = 0
	}
	formatted := strconv.FormatFloat(value, 'f', decimals, 64)
	parts := strings.SplitN(formatted, ".", 2)
	integer := parts[0]
	sign := ""
	if strings.HasPrefix(integer, "-") {
		sign = "-"
		integer = strings.TrimPrefix(integer, "-")
	}
	for i := len(integer) - 3; i > 0; i -= 3 {
		integer = integer[:i] + "." + integer[i:]
	}
	if decimals == 0 {
		return sign + integer
	}
	return sign + integer + "," + parts[1]
}

func barHeight(value float64) string {
	value = math.Max(0, math.Min(100, value))
	return strconv.FormatFloat(value, 'f', 1, 64) + "%"
}

func maxDistributionWeight(items []output.AbateReportDistribuicaoPeso) float64 {
	var maxValue float64
	for _, item := range items {
		if item.PesoTotal > maxValue {
			maxValue = item.PesoTotal
		}
	}
	return maxValue
}

func maxDistributionHeads(items []output.AbateReportDistribuicaoPeso) float64 {
	var maxValue float64
	for _, item := range items {
		if float64(item.Quantidade) > maxValue {
			maxValue = float64(item.Quantidade)
		}
	}
	return maxValue
}

func distributionBarHeight(value, maxValue float64) string {
	if maxValue <= 0 {
		return barHeight(0)
	}
	return barHeight(value / maxValue * 100)
}

func distributionHeadsBarHeight(value int, maxValue float64) string {
	return distributionBarHeight(float64(value), maxValue)
}

func denticaoBarColor(denticao int) string {
	colors := map[int]string{
		0: "#1b5e20",
		2: "#43a047",
		4: "#9ccc65",
		6: "#ef5350",
		8: "#e53935",
	}
	if color, ok := colors[denticao]; ok {
		return color
	}
	return "#9e9e9e"
}

func acabamentoBarColor(acabamento string) string {
	colors := map[string]string{
		"AUSENTE":          "#e53935",
		"ESCASSO":          "#ef5350",
		"ESCASSO POSITIVA": "#f57c00",
		"MEDIANO":          "#fbc02d",
		"UNIFORME":         "#43a047",
		"EXCESSIVO":        "#1b5e20",
	}
	if color, ok := colors[strings.ToUpper(strings.TrimSpace(acabamento))]; ok {
		return color
	}
	return "#9e9e9e"
}

func classificationRange(classification string) string {
	normalized := strings.ToLower(classification)
	switch {
	case strings.Contains(normalized, "fraco"):
		return "Abaixo de 15 @"
	case strings.Contains(normalized, "leve"):
		return "15 a 16 @"
	case strings.Contains(normalized, "médio"), strings.Contains(normalized, "medio"), strings.Contains(normalized, "normal"):
		return "16 a 25 @"
	case strings.Contains(normalized, "pesado"):
		return "Acima de 25 @"
	default:
		return "-"
	}
}

func classificationPenalty(classification string) string {
	normalized := strings.ToLower(classification)
	switch {
	case strings.Contains(normalized, "fraco"), strings.Contains(normalized, "pesado"):
		return "10% valor @"
	case strings.Contains(normalized, "leve"):
		return "5% valor @"
	case strings.Contains(normalized, "médio"), strings.Contains(normalized, "medio"), strings.Contains(normalized, "normal"):
		return "Preço normal"
	default:
		return "-"
	}
}

func buildPhotoPages(report output.AbateReport) []photoPage {
	const photosPerSectionOnFirstPage = 4
	const photosPerContinuationPage = 8
	firstFazenda := min(len(report.FotosFazenda), photosPerSectionOnFirstPage)
	firstFrigo := min(len(report.FotosFrigorifico), photosPerSectionOnFirstPage)
	pages := []photoPage{{
		Fazenda:     report.FotosFazenda[:firstFazenda],
		Frigorifico: report.FotosFrigorifico[:firstFrigo],
	}}
	for offset := firstFazenda; offset < len(report.FotosFazenda); offset += photosPerContinuationPage {
		end := min(offset+photosPerContinuationPage, len(report.FotosFazenda))
		pages = append(pages, photoPage{Fazenda: report.FotosFazenda[offset:end], FazendaContinuation: true})
	}
	for offset := firstFrigo; offset < len(report.FotosFrigorifico); offset += photosPerContinuationPage {
		end := min(offset+photosPerContinuationPage, len(report.FotosFrigorifico))
		pages = append(pages, photoPage{Frigorifico: report.FotosFrigorifico[offset:end], FrigoContinuation: true})
	}
	return pages
}
