package output

import "github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"

type UserPort interface {
	CreateUser(user *domain.User) error
	GetUsers() (*[]domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	UpdateUser(user *domain.User) error
	ActivateUser(id int) error
	DeactivateUser(id int) error
}
