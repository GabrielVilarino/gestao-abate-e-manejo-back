package repository

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type AgendaRepository struct {
	db *sql.DB
}

func NewAgendaRepository(db *sql.DB) AgendaRepository {
	return AgendaRepository{db: db}
}

func (a AgendaRepository) CreateAgenda(agenda *domain.Agenda) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := tx.QueryRow(`
		INSERT INTO public.agenda (user_id, fazenda_id, data_hora, observacao)
		SELECT $1, f.id, $3, $4
		FROM public.fazenda f WHERE f.id = $2
		RETURNING id`, agenda.UserID, agenda.FazendaID, agenda.DataHora.UTC(), agenda.Observacao).Scan(&agenda.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrFazendaAgendaInvalida
		}
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO public.agenda_notificacao (agenda_id, user_id, status, next_attempt_at)
		VALUES ($1, $2, 'pendente', GREATEST($3::timestamptz - INTERVAL '1 hour', CURRENT_TIMESTAMP))`,
		agenda.ID, agenda.UserID, agenda.DataHora.UTC())
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (a AgendaRepository) GetAgenda(agendaID, userID int) (*domain.Agenda, error) {
	agenda := &domain.Agenda{}
	err := a.db.QueryRow(`
		SELECT id, user_id, fazenda_id, data_hora, observacao
		FROM public.agenda WHERE id = $1 AND user_id = $2`, agendaID, userID).
		Scan(&agenda.ID, &agenda.UserID, &agenda.FazendaID, &agenda.DataHora, &agenda.Observacao)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrAgendaNaoEncontrada
	}
	if err != nil {
		return nil, err
	}
	return agenda, nil
}

func (a AgendaRepository) FindAgenda(filter domain.AgendaFilter) (*[]domain.Agenda, int, error) {
	var total int
	if err := a.db.QueryRow(`
		SELECT COUNT(*) FROM public.agenda
		WHERE user_id = $1 AND ($2::integer IS NULL OR fazenda_id = $2)
		AND ($3::timestamptz IS NULL OR data_hora >= $3)
		AND ($4::timestamptz IS NULL OR data_hora <= $4)`,
		filter.UserID, nullableInt(filter.FazendaID), nullableTime(filter.DataInicio), nullableTime(filter.DataFim)).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := a.db.Query(`
		SELECT id, user_id, fazenda_id, data_hora, observacao FROM public.agenda
		WHERE user_id = $1 AND ($2::integer IS NULL OR fazenda_id = $2)
		AND ($3::timestamptz IS NULL OR data_hora >= $3)
		AND ($4::timestamptz IS NULL OR data_hora <= $4)
		ORDER BY data_hora, id LIMIT $5 OFFSET $6`,
		filter.UserID, nullableInt(filter.FazendaID), nullableTime(filter.DataInicio), nullableTime(filter.DataFim),
		filter.Limite, (filter.Pagina-1)*filter.Limite)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	agendas := make([]domain.Agenda, 0)
	for rows.Next() {
		var agenda domain.Agenda
		if err := rows.Scan(&agenda.ID, &agenda.UserID, &agenda.FazendaID, &agenda.DataHora, &agenda.Observacao); err != nil {
			return nil, 0, err
		}
		agendas = append(agendas, agenda)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return &agendas, total, nil
}

func (a AgendaRepository) UpdateAgenda(agenda *domain.Agenda) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := lockAgendaForMutation(tx, agenda.ID, agenda.UserID); err != nil {
		return err
	}
	var fazendaExiste bool
	if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM public.fazenda WHERE id = $1)`, agenda.FazendaID).Scan(&fazendaExiste); err != nil {
		return err
	}
	if !fazendaExiste {
		return domain.ErrFazendaAgendaInvalida
	}
	result, err := tx.Exec(`
		UPDATE public.agenda SET fazenda_id = $1, data_hora = $2, observacao = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4 AND user_id = $5`, agenda.FazendaID, agenda.DataHora.UTC(), agenda.Observacao, agenda.ID, agenda.UserID)
	if err != nil {
		return err
	}
	if err := ensureDataFound(result, agenda.ID); err != nil {
		return domain.ErrAgendaNaoEncontrada
	}
	_, err = tx.Exec(`
		UPDATE public.agenda_notificacao
		SET status = CASE WHEN $2 <= CURRENT_TIMESTAMP THEN 'erro' ELSE 'pendente' END,
		    tentativas = 0,
		    next_attempt_at = GREATEST($2::timestamptz - INTERVAL '1 hour', CURRENT_TIMESTAMP),
		    locked_until = NULL, lease_token = NULL,
		    ultimo_erro = CASE WHEN $2 <= CURRENT_TIMESTAMP THEN 'Horário do agendamento já passou' ELSE NULL END,
		    enviado_em = NULL
		WHERE agenda_id = $1`, agenda.ID, agenda.DataHora.UTC())
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (a AgendaRepository) DeleteAgenda(agendaID, userID int) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := lockAgendaForMutation(tx, agendaID, userID); err != nil {
		return err
	}
	result, err := tx.Exec(`DELETE FROM public.agenda WHERE id = $1 AND user_id = $2`, agendaID, userID)
	if err != nil {
		return err
	}
	if err := ensureDataFound(result, agendaID); err != nil {
		return domain.ErrAgendaNaoEncontrada
	}
	return tx.Commit()
}

func (a AgendaRepository) CreatePushSubscription(subscription *domain.PushSubscription) error {
	err := a.db.QueryRow(`
		INSERT INTO public.push_subscription (user_id, endpoint, p256dh, auth, expires_at, ativo)
		VALUES ($1, $2, $3, $4, $5, TRUE)
		ON CONFLICT (endpoint) DO UPDATE SET
			user_id = EXCLUDED.user_id,
			p256dh = EXCLUDED.p256dh,
			auth = EXCLUDED.auth,
			expires_at = EXCLUDED.expires_at,
			ativo = TRUE,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id`, subscription.UserID, subscription.Endpoint, subscription.P256DH, subscription.Auth, subscription.ExpiresAt).
		Scan(&subscription.ID)
	return err
}

func (a AgendaRepository) DeletePushSubscription(subscriptionID, userID int) error {
	result, err := a.db.Exec(`DELETE FROM public.push_subscription WHERE id = $1 AND user_id = $2`, subscriptionID, userID)
	if err != nil {
		return err
	}
	if err := ensureDataFound(result, subscriptionID); err != nil {
		return domain.ErrAssinaturaNaoEncontrada
	}
	return nil
}

func (a AgendaRepository) ClaimDueAgendaNotifications(lease time.Duration, limit int) ([]domain.AgendaNotification, error) {
	leaseToken, err := newAgendaLeaseToken()
	if err != nil {
		return nil, err
	}
	tx, err := a.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`
		UPDATE public.agenda_notificacao n
		SET status = 'erro', locked_until = NULL, lease_token = NULL,
		    ultimo_erro = COALESCE(n.ultimo_erro, 'Horário do agendamento atingido sem entrega')
		FROM public.agenda a
		WHERE a.id = n.agenda_id AND a.data_hora <= CURRENT_TIMESTAMP
		AND (n.status = 'pendente' OR (n.status = 'processando' AND n.locked_until <= CURRENT_TIMESTAMP))`); err != nil {
		return nil, err
	}
	rows, err := tx.Query(`
		WITH due AS (
			SELECT n.id FROM public.agenda_notificacao n
			JOIN public.agenda a ON a.id = n.agenda_id
			WHERE a.data_hora > CURRENT_TIMESTAMP AND n.next_attempt_at <= CURRENT_TIMESTAMP
			AND (n.status = 'pendente' OR (n.status = 'processando' AND n.locked_until <= CURRENT_TIMESTAMP))
			ORDER BY n.next_attempt_at, n.id
			LIMIT $2 FOR UPDATE OF n SKIP LOCKED
		), claimed AS (
			UPDATE public.agenda_notificacao n
			SET status = 'processando', locked_until = CURRENT_TIMESTAMP + ($1 * INTERVAL '1 second'), lease_token = $3,
			    tentativas = tentativas + 1,
			    updated_at = CURRENT_TIMESTAMP
			FROM due WHERE n.id = due.id
			RETURNING n.id, n.agenda_id, n.user_id, n.tentativas, n.lease_token
		)
		SELECT c.id, c.agenda_id, c.user_id, c.tentativas, c.lease_token, a.fazenda_id, a.data_hora
		FROM claimed c JOIN public.agenda a ON a.id = c.agenda_id`, lease.Seconds(), limit, leaseToken)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs := make([]domain.AgendaNotification, 0)
	for rows.Next() {
		var job domain.AgendaNotification
		if err := rows.Scan(&job.ID, &job.AgendaID, &job.UserID, &job.Tentativas, &job.LeaseToken, &job.FazendaID, &job.DataHora); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (a AgendaRepository) RenewAgendaNotification(notificationID int, leaseToken string, lease time.Duration) error {
	result, err := a.db.Exec(`
		UPDATE public.agenda_notificacao
		SET locked_until = CURRENT_TIMESTAMP + ($3 * INTERVAL '1 second'), updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND lease_token = $2 AND status = 'processando' AND locked_until > CURRENT_TIMESTAMP`,
		notificationID, leaseToken, lease.Seconds())
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return domain.ErrNotificationLeaseLost
	}
	return nil
}

func (a AgendaRepository) GetPushSubscriptions(userID int) ([]domain.PushSubscription, error) {
	rows, err := a.db.Query(`
		SELECT s.id, s.user_id, s.endpoint, s.p256dh, s.auth, s.expires_at
		FROM public.push_subscription s
		JOIN public.usuario u ON u.id = s.user_id
		WHERE s.user_id = $1 AND s.ativo = TRUE AND u.ativo = TRUE
		AND (s.expires_at IS NULL OR s.expires_at > CURRENT_TIMESTAMP)
		ORDER BY s.id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	subscriptions := make([]domain.PushSubscription, 0)
	for rows.Next() {
		var subscription domain.PushSubscription
		if err := rows.Scan(&subscription.ID, &subscription.UserID, &subscription.Endpoint, &subscription.P256DH, &subscription.Auth, &subscription.ExpiresAt); err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions, rows.Err()
}

func (a AgendaRepository) FinishAgendaNotification(notificationID int, leaseToken string, sendError string) error {
	var result sql.Result
	var err error
	if sendError == "" {
		result, err = a.db.Exec(`
			UPDATE public.agenda_notificacao
			SET status = 'enviada', enviado_em = CURRENT_TIMESTAMP, locked_until = NULL, lease_token = NULL, ultimo_erro = NULL,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND lease_token = $2 AND status = 'processando' AND locked_until > CURRENT_TIMESTAMP`, notificationID, leaseToken)
	} else {
		result, err = a.db.Exec(`
			UPDATE public.agenda_notificacao n
			SET status = CASE WHEN a.data_hora <= CURRENT_TIMESTAMP THEN 'erro' ELSE 'pendente' END,
			    next_attempt_at = CURRENT_TIMESTAMP + INTERVAL '1 minute', locked_until = NULL, lease_token = NULL, ultimo_erro = $2,
			    updated_at = CURRENT_TIMESTAMP
			FROM public.agenda a
			WHERE n.id = $1 AND n.agenda_id = a.id AND n.lease_token = $3
			AND n.status = 'processando' AND n.locked_until > CURRENT_TIMESTAMP`, notificationID, sendError, leaseToken)
	}
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return domain.ErrNotificationLeaseLost
	}
	return nil
}

func lockAgendaForMutation(tx *sql.Tx, agendaID, userID int) error {
	var status string
	var leaseAtiva bool
	err := tx.QueryRow(`
		SELECT n.status, COALESCE(n.locked_until > CURRENT_TIMESTAMP, TRUE)
		FROM public.agenda_notificacao n
		JOIN public.agenda a ON a.id = n.agenda_id
		WHERE n.agenda_id = $1 AND a.user_id = $2
		FOR UPDATE OF n`, agendaID, userID).Scan(&status, &leaseAtiva)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrAgendaNaoEncontrada
	}
	if err != nil {
		return err
	}
	if status == "processando" && leaseAtiva {
		return domain.ErrAgendaEmProcessamento
	}
	return nil
}

func newAgendaLeaseToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (a AgendaRepository) DeletePushSubscriptionByEndpoint(endpoint string) error {
	_, err := a.db.Exec(`DELETE FROM public.push_subscription WHERE endpoint = $1`, endpoint)
	return err
}

func nullableInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC()
}
