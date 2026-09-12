package domain

type Fazenda struct {
	ID             int
	Nome           string
	Cidade         string
	InscricaoRural string
	Observacao     string
	IDProprietario int
	Ativo          bool
}
