package domain

import "errors"

var (
	// User errors
	ErrInvalidCredentials = errors.New("Credenciais inválidas")
	ErrEmailAlreadyExists = errors.New("E-mail já cadastrado")

	// Proprietario errors
	ErrCpfAlreadyExists = errors.New("CPF já cadastrado")
)
