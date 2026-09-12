package response

type FazendaErrorResponse struct {
	Error string `json:"error"`
}

type FazendaSuccessResponse struct {
	Message string `json:"message"`
}

type FazendaDataResponse struct {
	ID             int    `json:"id"`
	Nome           string `json:"nome"`
	Cidade         string `json:"cidade"`
	InscricaoRural string `json:"inscricao_rural"`
	Observacao     string `json:"observacao"`
	IDProprietario int    `json:"id_proprietario"`
	Ativo          bool   `json:"ativo"`
}

type GetFazendasResponse struct {
	Fazendas []FazendaDataResponse `json:"fazendas"`
}
