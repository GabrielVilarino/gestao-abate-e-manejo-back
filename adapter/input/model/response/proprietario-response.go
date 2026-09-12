package response

type ProprietarioErrorResponse struct {
	Error string `json:"error"`
}

type ProprietarioSuccessResponse struct {
	Message string `json:"message"`
}

type ProprietarioDataResponse struct {
	ID         int    `json:"id"`
	Nome       string `json:"nome"`
	CPF        string `json:"cpf"`
	Observacao string `json:"observacao"`
	Ativo      bool   `json:"ativo"`
}

type GetProprietariosResponse struct {
	Proprietarios []ProprietarioDataResponse `json:"proprietarios"`
}
