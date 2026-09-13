package response

type AgendaErrorResponse struct {
	Error string `json:"error"`
}

type AgendaSuccessResponse struct {
	Message string `json:"message"`
	ID      int    `json:"id,omitempty"`
}

type AgendaDataResponse struct {
	ID         int    `json:"id"`
	FazendaID  int    `json:"fazenda_id"`
	DataHora   string `json:"data_hora"`
	Observacao string `json:"observacao"`
}

type GetAgendasResponse struct {
	Agendas []AgendaDataResponse `json:"agendas"`
	Pagina  int                  `json:"pagina"`
	Limite  int                  `json:"limite"`
	Total   int                  `json:"total"`
}

type PushSubscriptionDataResponse struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}
