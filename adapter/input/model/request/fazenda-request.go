package request

type FazendaCreateRequest struct {
	Nome                 string  `json:"nome" binding:"required"`
	Cidade               string  `json:"cidade" binding:"required"`
	InscricaoRural       string  `json:"inscricao_rural" binding:"required"`
	Observacao           string  `json:"observacao"`
	DistanciaFrigorifico float64 `json:"distancia_frigorifico" binding:"required,gt=0"`
	QuantidadeRebanho    int     `json:"quantidade_rebanho" binding:"required,gt=0"`
	IDProprietario       int     `json:"id_proprietario" binding:"required,gt=0"`
}

type FazendaUpdateRequest struct {
	ID                   int     `json:"id" binding:"required,gt=0"`
	Nome                 string  `json:"nome" binding:"required"`
	Cidade               string  `json:"cidade" binding:"required"`
	InscricaoRural       string  `json:"inscricao_rural" binding:"required"`
	Observacao           string  `json:"observacao"`
	DistanciaFrigorifico float64 `json:"distancia_frigorifico" binding:"required,gt=0"`
	QuantidadeRebanho    int     `json:"quantidade_rebanho" binding:"required,gt=0"`
}
