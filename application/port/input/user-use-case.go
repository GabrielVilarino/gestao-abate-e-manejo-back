package input

import "github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"

type UserUseCase interface {
	Login(email, password string) (*string, error)
	CreateUser(user *domain.User) error
	GetUsers() (*[]domain.User, error)
	UpdateUser(user *domain.User) error
	DeleteUser(id string) error
}
