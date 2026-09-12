package service

import (
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/output"
)

type UserService struct {
	UserPort  output.UserPort
	HashPort  output.HashPort
	TokenPort output.TokenPort
}

func NewUserService(
	userPort output.UserPort,
	hashPort output.HashPort,
	tokenPort output.TokenPort,
) *UserService {
	return &UserService{
		UserPort:  userPort,
		HashPort:  hashPort,
		TokenPort: tokenPort,
	}
}

func (u *UserService) Login(email, password string) (*string, error) {
	userData, err := u.UserPort.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if userData == nil || !userData.Ativo {
		return nil, domain.ErrInvalidCredentials
	}

	if err := u.HashPort.Compare(password, userData.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	token, err := u.TokenPort.Generate(*userData)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (u *UserService) CreateUser(user *domain.User) error {
	userData, err := u.UserPort.GetUserByEmail(user.Email)
	if err != nil {
		return err
	}
	if userData != nil {
		return domain.ErrEmailAlreadyExists
	}

	hashedPassword, err := u.HashPort.Hash(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	return u.UserPort.CreateUser(user)
}

func (u *UserService) GetUsers() (*[]domain.User, error) {
	return u.UserPort.GetUsers()
}

func (u *UserService) UpdateUser(user *domain.User) error {
	return u.UserPort.UpdateUser(user)
}

func (u *UserService) ActivateUser(id int) error {
	return u.UserPort.ActivateUser(id)
}

func (u *UserService) DeactivateUser(id int) error {
	return u.UserPort.DeactivateUser(id)
}
