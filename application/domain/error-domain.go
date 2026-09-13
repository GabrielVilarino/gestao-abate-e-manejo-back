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

	// Abate errors
	ErrAbateNaoEncontrado  = errors.New("Abate não encontrado")
	ErrFotoNaoEncontrada   = errors.New("Foto não encontrada")
	ErrPeriodoInvalido     = errors.New("Data inicial não pode ser maior que a data final")
	ErrEtapaFotoInvalida   = errors.New("Etapa da foto inválida")
	ErrFormatoFotoInvalido = errors.New("Formato de foto inválido; envie PNG ou JPG")
	ErrTamanhoFotoInvalido = errors.New("A foto deve ter no máximo 10 MB")
	ErrDadosGeraisComFotos = errors.New("Não é possível alterar fazenda ou número do lote após o envio de fotos")

	// Agenda and push errors
	ErrAgendaNaoEncontrada     = errors.New("Agendamento não encontrado")
	ErrPeriodoAgendaInvalido   = errors.New("Data inicial não pode ser maior que a data final")
	ErrHorarioAgendaInvalido   = errors.New("Data e hora do agendamento inválida")
	ErrFazendaAgendaInvalida   = errors.New("Fazenda do agendamento inválida")
	ErrAssinaturaPushInvalida  = errors.New("Assinatura push inválida")
	ErrAssinaturaNaoEncontrada = errors.New("Assinatura push não encontrada")
	ErrPushSubscriptionGone    = errors.New("assinatura push expirada ou inválida")
	ErrAgendaEmProcessamento   = errors.New("Agendamento está sendo processado; tente novamente em instantes")
	ErrNotificationLeaseLost   = errors.New("posse da notificação expirada")
)
