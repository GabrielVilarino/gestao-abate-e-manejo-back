package output

import "github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"

type TokenPort interface {
	Generate(user domain.User) (string, error)
	Validate(token string) (*domain.User, error)
}
