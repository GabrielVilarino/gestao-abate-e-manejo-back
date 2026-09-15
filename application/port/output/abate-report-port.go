package output

import (
	"context"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type AbateReportPhoto struct {
	NomeOriginal string
	DataURL      string
}

type AbateReportDenticao struct {
	Denticao   int
	Quantidade int
	Percentual float64
}

type AbateReportAcabamento struct {
	Classificacao string
	Quantidade    int
	Percentual    float64
}

type AbateReportClassificacao struct {
	Classificacao string
	Quantidade    int
	Percentual    float64
}

type AbateReportDistribuicaoPeso struct {
	Classificacao string
	Quantidade    int
	PesoTotal     float64
	MediaKG       float64
	MediaArroba   float64
	Percentual    float64
}

type AbateReport struct {
	Abate                  domain.Abate
	QuantidadeAnimais      int
	MediaKGFrigorifico     float64
	MediaArrobaFrigorifico float64
	MediaKGFazenda         float64
	RendimentoCarcaca      float64
	PesoMedioBalancao      float64
	Esvaziamento           float64
	RendimentoBalancao     float64
	Observacao             string
	Denticoes              []AbateReportDenticao
	Acabamentos            []AbateReportAcabamento
	Classificacoes         []AbateReportClassificacao
	DistribuicoesPeso      []AbateReportDistribuicaoPeso
	FotosFazenda           []AbateReportPhoto
	FotosFrigorifico       []AbateReportPhoto
}

type AbateReportGenerator interface {
	Generate(ctx context.Context, reports []AbateReport) ([]byte, error)
}
