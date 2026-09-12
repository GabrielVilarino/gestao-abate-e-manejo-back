package request

type ProprietarioCreateRequest struct {
	Nome       string `json:"nome" binding:"required"`
	CPF        string `json:"cpf" binding:"required,len=11"`
	Observacao string `json:"observacao"`
}

type ProprietarioUpdateRequest struct {
	ID         int    `json:"id" binding:"required,gt=0"`
	Nome       string `json:"nome" binding:"required"`
	CPF        string `json:"cpf" binding:"required,len=11"`
	Observacao string `json:"observacao"`
}
