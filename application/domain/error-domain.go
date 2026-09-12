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
	ErrCpfAlreadyExists       = errors.New("CPF já cadastrado")
	ErrCreateProprietario     = errors.New("Erro ao criar proprietário")
	ErrUpdateProprietario     = errors.New("Erro ao atualizar proprietário")
	ErrActivateProprietario   = errors.New("Erro ao ativar proprietário")
	ErrDeactivateProprietario = errors.New("Erro ao desativar proprietário")

	// Fazenda errors
	ErrCreateFazenda       = errors.New("Erro ao criar fazenda")
	ErrUpdateFazenda       = errors.New("Erro ao atualizar fazenda")
	ErrActivateFazenda     = errors.New("Erro ao ativar fazenda")
	ErrDeactivateFazenda   = errors.New("Erro ao desativar fazenda")
	ErrProprietarioInativo = errors.New("Não é possível ativar a fazenda porque o proprietário está inativo")
)
