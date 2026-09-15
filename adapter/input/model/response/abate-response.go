package response

type AbateErrorResponse struct {
	Error string `json:"error"`
}

type AbateSuccessResponse struct {
	ID      int    `json:"id,omitempty"`
	Message string `json:"message"`
}

type FotoAbateResponse struct {
	ID           int    `json:"id"`
	Etapa        string `json:"etapa"`
	NomeOriginal string `json:"nome_original"`
	ContentType  string `json:"content_type"`
	Tamanho      int64  `json:"tamanho"`
	SHA256       string `json:"sha256"`
}

type QtdDenticaoResponse struct {
	QtdDenticao int `json:"denticao"`
	QtdAnimais  int `json:"qtd_animais"`
}

type AcabamentoCarcacaResponse struct {
	Acabamento string `json:"acabamento"`
	QtdAnimais int    `json:"qtd_animais"`
}

type ClassificacaoFrigorificoResponse struct {
	Classificacao string `json:"classificacao"`
	QtdAnimais    int    `json:"qtd_animais"`
}

type DistribuicaoPesoResponse struct {
	Classificacao string  `json:"classificacao"`
	QtdAnimais    int     `json:"qtd_animais"`
	PesoTotal     float64 `json:"peso_total"`
}

type DadosGeraisAbateResponse struct {
	DataAbate            string  `json:"data_abate"`
	FazendaID            int     `json:"fazenda_id"`
	NumeroLote           int     `json:"numero_lote"`
	NomeFrigorifico      string  `json:"nome_frigorifico"`
	DistanciaFrigorifico float64 `json:"distancia_frigorifico"`
	CategoriaAnimal      string  `json:"categoria_animal"`
	PrecoFunrural        float64 `json:"preco_funrural"`
	PrecoSemFunrural     float64 `json:"preco_sem_funrural"`
	Observacao           *string `json:"observacao"`
}

type EtapaFazendaResponse struct {
	QuantidadeAnimal []QtdDenticaoResponse `json:"quantidade_animal"`
	PesoTotal        float64               `json:"peso_total"`
	Fotos            []FotoAbateResponse   `json:"fotos"`
}

type EtapaFrigorificoResponse struct {
	PesoTotal                float64                            `json:"peso_total"`
	Balancao                 float64                            `json:"balancao"`
	AcabamentoCarcaca        []AcabamentoCarcacaResponse        `json:"acabamento_carcaca"`
	ClassificacaoFrigorifico []ClassificacaoFrigorificoResponse `json:"classificacao_frigorifico"`
	DistribuicaoPeso         []DistribuicaoPesoResponse         `json:"distribuicao_peso"`
	Fotos                    []FotoAbateResponse                `json:"fotos"`
}

type AbateDataResponse struct {
	ID               int                      `json:"id"`
	ProprietarioID   int                      `json:"proprietario_id"`
	NomeProprietario string                   `json:"nome_proprietario"`
	NomeFazenda      string                   `json:"nome_fazenda"`
	DadosGerais      DadosGeraisAbateResponse `json:"dados_gerais"`
	EtapaFazenda     EtapaFazendaResponse     `json:"etapa_fazenda"`
	EtapaFrigorifico EtapaFrigorificoResponse `json:"etapa_frigorifico"`
}

type GetAbatesResponse struct {
	Abates []AbateDataResponse `json:"abates"`
	Pagina int                 `json:"pagina"`
	Limite int                 `json:"limite"`
}
