package request

type DadosGeraisAbateRequest struct {
	DataAbate            string  `json:"data_abate" binding:"required"`
	FazendaID            int     `json:"fazenda_id" binding:"required,gt=0"`
	NumeroLote           int     `json:"numero_lote" binding:"required,gt=0"`
	NomeFrigorifico      string  `json:"nome_frigorifico" binding:"required"`
	DistanciaFrigorifico float64 `json:"distancia_frigorifico" binding:"required,gt=0"`
	CategoriaAnimal      string  `json:"categoria_animal" binding:"required"`
	PrecoFunrural        float64 `json:"preco_funrural" binding:"gte=0"`
	PrecoSemFunrural     float64 `json:"preco_sem_funrural" binding:"gte=0"`
}

type QtdDenticaoRequest struct {
	QtdDenticao int `json:"denticao" binding:"gte=0"`
	QtdAnimais  int `json:"qtd_animais" binding:"gte=0"`
}

type AcabamentoCarcacaRequest struct {
	Acabamento string `json:"acabamento" binding:"required"`
	QtdAnimais int    `json:"qtd_animais" binding:"gte=0"`
}

type ClassificacaoFrigorificoRequest struct {
	Classificacao string `json:"classificacao" binding:"required"`
	QtdAnimais    int    `json:"qtd_animais" binding:"gte=0"`
}

type DistribuicaoPesoRequest struct {
	Classificacao string  `json:"classificacao" binding:"required"`
	QtdAnimais    int     `json:"qtd_animais" binding:"gte=0"`
	PesoTotal     float64 `json:"peso_total" binding:"gte=0"`
}

type EtapaFazendaRequest struct {
	QuantidadeAnimal []QtdDenticaoRequest `json:"quantidade_animal" binding:"dive"`
	PesoTotal        float64              `json:"peso_total" binding:"gte=0"`
}

type EtapaFrigorificoRequest struct {
	PesoTotal                float64                           `json:"peso_total" binding:"gte=0"`
	Balancao                 float64                           `json:"balancao" binding:"gte=0"`
	AcabamentoCarcaca        []AcabamentoCarcacaRequest        `json:"acabamento_carcaca" binding:"dive"`
	ClassificacaoFrigorifico []ClassificacaoFrigorificoRequest `json:"classificacao_frigorifico" binding:"dive"`
	DistribuicaoPeso         []DistribuicaoPesoRequest         `json:"distribuicao_peso" binding:"dive"`
}

type AbateCreateRequest struct {
	DadosGeraisAbate DadosGeraisAbateRequest `json:"dados_gerais" binding:"required"`
	EtapaFazenda     EtapaFazendaRequest     `json:"etapa_fazenda"`
	EtapaFrigorifico EtapaFrigorificoRequest `json:"etapa_frigorifico"`
}
