package domain

import "time"

const (
	EtapaFazendaFoto     = "FAZENDA"
	EtapaFrigorificoFoto = "FRIGORIFICO"
)

type QtdDenticao struct {
	QtdDenticao int
	QtdAnimais  int
}

type AcabamentoCarcaca struct {
	Acabamento string
	QtdAnimais int
}

type ClassificacaoFrigorifico struct {
	Classificacao string
	QtdAnimais    int
}

type DistribuicaoPeso struct {
	Classificacao string
	QtdAnimais    int
	PesoTotal     float64
}

type FotoAbate struct {
	ID           int
	AbateID      int
	Etapa        string
	ObjectKey    string
	NomeOriginal string
	ContentType  string
	Tamanho      int64
	SHA256       string
}

type DadosGeraisAbate struct {
	DataAbate            time.Time
	FazendaID            int
	NumeroLote           int
	NomeFrigorifico      string
	DistanciaFrigorifico float64
	CategoriaAnimal      string
	PrecoFunrural        float64
	PrecoSemFunrural     float64
}

type EtapaFazenda struct {
	QuantidadeAnimal []QtdDenticao
	PesoTotal        float64
	Fotos            []FotoAbate
}

type EtapaFrigorifico struct {
	PesoTotal                float64
	Balancao                 float64
	AcabamentoCarcaca        []AcabamentoCarcaca
	ClassificacaoFrigorifico []ClassificacaoFrigorifico
	DistribuicaoPeso         []DistribuicaoPeso
	Fotos                    []FotoAbate
}

type Abate struct {
	ID               int
	ProprietarioID   int
	NomeProprietario string
	NomeFazenda      string
	DadosGeraisAbate DadosGeraisAbate
	EtapaFazenda     EtapaFazenda
	EtapaFrigorifico EtapaFrigorifico
}

type FiltroAbate struct {
	ProprietarioID *int
	FazendaID      *int
	NumeroLote     *int
	DataInicio     *time.Time
	DataFim        *time.Time
	Limit          int
	Offset         int
}

type ContextoFotoAbate struct {
	ProprietarioID int
	NumeroLote     int
}
