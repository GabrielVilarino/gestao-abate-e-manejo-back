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
		".bar-chart { height:106px; padding:10px 18px 5px;",
		".classification-wrap { padding:5px 13px 6px;",
		".classification-wrap th { padding:3px 5px;",
		".classification-wrap td { padding:2px 5px;",
	} {
		if !strings.Contains(reportTemplateHTML, expected) {
			t.Fatalf("espaçamento inferior ausente no template: %q", expected)
		}
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
		"0 de 18 a 19.9 arrobas",
		"0 de 20 a 21.9 arrobas",
		"0 de 22 a 23.9 arrobas",
		"0 acima de 24 arrobas",
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("linha zerada não renderizada: %q", expected)
		}
	}
}
