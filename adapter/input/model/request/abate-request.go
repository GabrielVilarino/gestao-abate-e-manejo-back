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
	Observacao           *string `json:"observacao"`
}

type QtdDenticaoRequest struct {
	QtdDenticao *int `json:"denticao" binding:"required,oneof=0 2 4 6 8"`
	QtdAnimais  int  `json:"qtd_animais" binding:"gte=0"`
}

type AcabamentoCarcacaRequest struct {
	Acabamento string `json:"acabamento" binding:"required,oneof='AUSENTE' 'ESCASSO' 'MEDIANO' 'UNIFORME' 'EXCESSIVO' 'MEDIANO UP'"`
	QtdAnimais int    `json:"qtd_animais" binding:"gte=0"`
}

type ClassificacaoFrigorificoRequest struct {
	Classificacao string `json:"classificacao" binding:"required,oneof='BOI FRACO' 'BOI LEVE' 'BOI MÉDIO / NORMAL' 'BOI PESADO'"`
	QtdAnimais    int    `json:"qtd_animais" binding:"gte=0"`
}

type DistribuicaoPesoRequest struct {
	Classificacao string  `json:"classificacao" binding:"required,oneof='18 a 19.9' '20 a 21.9' '22 a 23.9' 'acima de 24'"`
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

type AbateReportRequest struct {
	AbateIDs []int `json:"abate_ids" binding:"required,min=1,dive,gt=0"`
}
