package domain

type QtdDenticao struct {
	QtdDenticao int
	QtdAnimais  int
}

type AcabamentoCarcaca struct {
	Acabamento string
	QtdAnimais float64
}

type ClassificacaoFrigorifico struct {
	Classificacao string
	QtdAnimais    float64
}

type DistribuicaoPeso struct {
	Classificacao string
	QtdAnimais    float64
	PesoTotal     float64
}

type DadosGeraisAbate struct {
	DataAbate        string
	FazendaID        uint64
	NomeFrigorifico  string
	CategoriaAnimal  string
	PrecoFunrural    float64
	PrecoSemFunrural float64
}

type EtapaFazenda struct {
	QuantidadeAnimal []QtdDenticao
	PesoTotal        float64
	Fotos            []string
}

type EtapaFrigorifico struct {
	PesoTotal                float64
	Balancao                 float64
	AcabamentoCarcaca        []AcabamentoCarcaca
	ClassificacaoFrigorifico []ClassificacaoFrigorifico
	DistribuicaoPeso         []DistribuicaoPeso
	Fotos                    []string
}

type Abate struct {
	ID               int
	DadosGeraisAbate DadosGeraisAbate
	EtapaFazenda     EtapaFazenda
	EtapaFrigorifico EtapaFrigorifico
}
