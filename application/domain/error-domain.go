package domain

import "errors"

var (
	// User errors
	ErrInvalidCredentials = errors.New("Credenciais inválidas")
	ErrEmailAlreadyExists = errors.New("E-mail já cadastrado")
	ErrCreateUser         = errors.New("Erro ao criar usuário")
	ErrUpdateUser         = errors.New("Erro ao atualizar usuário")
	ErrActivateUser       = errors.New("Erro ao ativar usuário")
	ErrDeactivateUser     = errors.New("Erro ao desativar usuário")

	// Proprietario errors
	ErrCpfAlreadyExists = errors.New("CPF já cadastrado")
)
