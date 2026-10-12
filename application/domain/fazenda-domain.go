package domain

type Fazenda struct {
	ID                   int
	Nome                 string
	Cidade               string
	InscricaoRural       string
	Observacao           string
	DistanciaFrigorifico float64
	QuantidadeRebanho    int
	IDProprietario       int
	Ativo                bool
}
