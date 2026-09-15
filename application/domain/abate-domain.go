package domain

import "time"

const (
	EtapaFazendaFoto     = "FAZENDA"
	EtapaFrigorificoFoto = "FRIGORIFICO"

	DenticaoZero   = 0
	DenticaoDois   = 2
	DenticaoQuatro = 4
	DenticaoSeis   = 6
	DenticaoOito   = 8

	AcabamentoCarcacaAusente   = "AUSENTE"
	AcabamentoCarcacaEscasso   = "ESCASSO"
	AcabamentoCarcacaMediano   = "MEDIANO"
	AcabamentoCarcacaUniforme  = "UNIFORME"
	AcabamentoCarcacaExcessivo = "EXCESSIVO"
	AcabamentoCarcacaMedianoUP = "MEDIANO UP"

	ClassificacaoFrigorificoBoiFraco       = "BOI FRACO"
	ClassificacaoFrigorificoBoiLeve        = "BOI LEVE"
	ClassificacaoFrigorificoBoiMedioNormal = "BOI MÉDIO / NORMAL"
	ClassificacaoFrigorificoBoiPesado      = "BOI PESADO"

	FaixaDistribuicaoPeso18A19Ponto9 = "18 a 19.9"
	FaixaDistribuicaoPeso20A21Ponto9 = "20 a 21.9"
	FaixaDistribuicaoPeso22A23Ponto9 = "22 a 23.9"
	FaixaDistribuicaoPesoAcimaDe24   = "acima de 24"
)

func DenticoesAbate() []int {
	return []int{DenticaoZero, DenticaoDois, DenticaoQuatro, DenticaoSeis, DenticaoOito}
}

func AcabamentosCarcacaAbate() []string {
	return []string{
		AcabamentoCarcacaAusente,
		AcabamentoCarcacaEscasso,
		AcabamentoCarcacaMediano,
		AcabamentoCarcacaUniforme,
		AcabamentoCarcacaExcessivo,
		AcabamentoCarcacaMedianoUP,
	}
}

func ClassificacoesFrigorificoAbate() []string {
	return []string{
		ClassificacaoFrigorificoBoiFraco,
		ClassificacaoFrigorificoBoiLeve,
		ClassificacaoFrigorificoBoiMedioNormal,
		ClassificacaoFrigorificoBoiPesado,
	}
}

func FaixasDistribuicaoPesoAbate() []string {
	return []string{
		FaixaDistribuicaoPeso18A19Ponto9,
		FaixaDistribuicaoPeso20A21Ponto9,
		FaixaDistribuicaoPeso22A23Ponto9,
		FaixaDistribuicaoPesoAcimaDe24,
	}
}

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
	Observacao           *string
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
