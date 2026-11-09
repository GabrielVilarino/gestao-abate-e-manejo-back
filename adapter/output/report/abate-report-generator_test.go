package report

import (
	"strings"
	"testing"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/output"
)

func TestReportTemplateKeepsLayoutPageBreaksAndEscapesText(t *testing.T) {
	generator, err := NewChromedpPDFGenerator()
	if err != nil {
		t.Fatal(err)
	}
	report := output.AbateReport{
		Observacao:       "<script>alert('x')</script>",
		FotosFazenda:     []output.AbateReportPhoto{{NomeOriginal: "fazenda.jpg", DataURL: "data:image/jpeg;base64,Zm90bw=="}},
		FotosFrigorifico: []output.AbateReportPhoto{{NomeOriginal: "frigo.jpg", DataURL: "data:image/jpeg;base64,Zm90bw=="}},
	}
	html, err := generator.renderHTML([]output.AbateReport{report})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "data:image/png;base64,") {
		t.Fatal("template did not embed the report logo")
	}
	for _, expected := range []string{
		`class="page summary-page"`, `class="page photos-page"`, `class="watermark"`,
		"Fotos tiradas na etapa da fazenda", "Fotos tiradas na etapa do frigorífico",
		"data:image/jpeg;base64,Zm90bw==", "&lt;script&gt;alert",
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("HTML não contém %q", expected)
		}
	}
	if strings.Contains(html, "#ZgotmplZ") || strings.Contains(html, "<script>alert") {
		t.Fatal("template gerou URL bloqueada ou texto não escapado")
	}
}

func TestBuildPhotoPagesIncludesEveryPhoto(t *testing.T) {
	report := output.AbateReport{
		FotosFazenda:     make([]output.AbateReportPhoto, 13),
		FotosFrigorifico: make([]output.AbateReportPhoto, 10),
	}
	pages := buildPhotoPages(report)
	var farm, fridge int
	for _, page := range pages {
		farm += len(page.Fazenda)
		fridge += len(page.Frigorifico)
	}
	if farm != 13 || fridge != 10 {
		t.Fatalf("fazenda=%d frigorífico=%d", farm, fridge)
	}
}

func TestFormatNumberUsesBrazilianSeparators(t *testing.T) {
	if got := formatNumber(17475.6, 2); got != "17.475,60" {
		t.Fatalf("got=%q", got)
	}
}

func TestReportBarColorsFollowQualityPalettes(t *testing.T) {
	if got := denticaoBarColor(0); got != "#1b5e20" || denticaoBarColor(4) != "#9ccc65" || denticaoBarColor(8) != "#e53935" {
		t.Fatalf("paleta de dentição inesperada: 0=%s 4=%s 8=%s", got, denticaoBarColor(4), denticaoBarColor(8))
	}
	acabamentos := []struct {
		name  string
		color string
	}{
		{"AUSENTE", "#e53935"},
		{"ESCASSO", "#ef5350"},
		{"ESCASSO POSITIVA", "#f57c00"},
		{"MEDIANO", "#fbc02d"},
		{"UNIFORME", "#43a047"},
		{"EXCESSIVO", "#1b5e20"},
	}
	for _, test := range acabamentos {
		if got := acabamentoBarColor(test.name); got != test.color {
			t.Fatalf("cor de %s = %s, esperada %s", test.name, got, test.color)
		}
	}
}

func TestReportTemplateDoesNotClipWeightDistributionRows(t *testing.T) {
	if strings.Contains(reportTemplateHTML, ".weight-card { height:86px") {
		t.Fatal("card de distribuição de peso ainda possui altura fixa insuficiente")
	}
	if !strings.Contains(reportTemplateHTML, ".weight-card { min-height:104px") || !strings.Contains(reportTemplateHTML, "overflow:visible") {
		t.Fatal("card de distribuição de peso não permite exibir todas as faixas e o total")
	}
}

func TestReportTemplateKeepsBottomSpacingInSummaryCards(t *testing.T) {
	for _, expected := range []string{
		"grid-template-rows:175px 175px 160px;",
		".bar-chart { height:150px; padding:10px 52px 7px;",
		".classification-wrap { padding:9px 14px 8px;",
		".classification-wrap th { padding:3px 5px;",
		".classification-wrap td { padding:2.5px 5px;",
	} {
		if !strings.Contains(reportTemplateHTML, expected) {
			t.Fatalf("espaçamento inferior ausente no template: %q", expected)
		}
	}
}

func TestReportTemplateRendersWeightAndHeadsDistributionChart(t *testing.T) {
	generator, err := NewChromedpPDFGenerator()
	if err != nil {
		t.Fatal(err)
	}
	report := output.AbateReport{
		DistribuicoesPeso: []output.AbateReportDistribuicaoPeso{
			{Classificacao: "18 a 19.9", Quantidade: 2, PesoTotal: 900},
			{Classificacao: "20 a 21.9", Quantidade: 4, PesoTotal: 1900},
			{Classificacao: "22 a 23.9", Quantidade: 3, PesoTotal: 1500},
			{Classificacao: "acima de 24", Quantidade: 1, PesoTotal: 500},
		},
	}
	html, err := generator.renderHTML([]output.AbateReport{report})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(html, `class="distribution-group"`) != 4 {
		t.Fatalf("grupos de distribuiÃ§Ã£o=%d, esperado=4", strings.Count(html, `class="distribution-group"`))
	}
	if !strings.Contains(html, `class="distribution-bar weight"`) || !strings.Contains(html, `class="distribution-bar heads"`) {
		t.Fatal("grÃ¡fico agrupado de peso e cabeÃ§as nÃ£o foi renderizado")
	}
	if !strings.Contains(html, `--distribution-height:100.0%`) {
		t.Fatal("grÃ¡fico nÃ£o normalizou as maiores barras")
	}
}

func TestReportTemplateRendersBarsForZeroValues(t *testing.T) {
	generator, err := NewChromedpPDFGenerator()
	if err != nil {
		t.Fatal(err)
	}
	report := output.AbateReport{
		Denticoes: []output.AbateReportDenticao{
			{Denticao: 0}, {Denticao: 2}, {Denticao: 4}, {Denticao: 6}, {Denticao: 8},
		},
		Acabamentos: make([]output.AbateReportAcabamento, 6),
		DistribuicoesPeso: []output.AbateReportDistribuicaoPeso{
			{Classificacao: "18 a 19.9"},
			{Classificacao: "20 a 21.9"},
			{Classificacao: "22 a 23.9"},
			{Classificacao: "acima de 24"},
		},
	}
	html, err := generator.renderHTML([]output.AbateReport{report})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(html, `class="bar-item"`); got != 11 {
		t.Fatalf("barras renderizadas=%d, esperado=11", got)
	}
	if got := strings.Count(html, "--height:0.0%"); got != 11 {
		t.Fatalf("barras zeradas=%d, esperado=11", got)
	}
	for _, expected := range []string{
		"0 dentes</h3><small>até 18 meses</small><b>0 animais</b>",
		"2 dentes</h3><small>18 a 24 meses</small><b>0 animais</b>",
		"4 dentes</h3><small>24 a 36 meses</small><b>0 animais</b>",
		"6 dentes</h3><small>36 a 48 meses</small><b>0 animais</b>",
		"8 dentes</h3><small>acima de 48 meses</small><b>0 animais</b>",
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("linha zerada não renderizada: %q", expected)
		}
	}
	if strings.Contains(html, "Detalhamento da distribuição de peso") {
		t.Fatal("detalhamento da distribuição de peso não deveria ser renderizado")
	}
}
