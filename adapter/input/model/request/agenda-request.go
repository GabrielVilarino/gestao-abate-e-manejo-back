package request

type AgendaCreateRequest struct {
	FazendaID  int    `json:"fazenda_id" binding:"required,gt=0"`
	DataHora   string `json:"data_hora" binding:"required"`
	Observacao string `json:"observacao"`
}

type AgendaUpdateRequest struct {
	FazendaID  int    `json:"fazenda_id" binding:"required,gt=0"`
	DataHora   string `json:"data_hora" binding:"required"`
	Observacao string `json:"observacao"`
}

type PushSubscriptionRequest struct {
	Endpoint       string `json:"endpoint" binding:"required,url"`
	ExpirationTime *int64 `json:"expirationTime"`
	Keys           struct {
		P256DH string `json:"p256dh" binding:"required"`
		Auth   string `json:"auth" binding:"required"`
	} `json:"keys" binding:"required"`
}
