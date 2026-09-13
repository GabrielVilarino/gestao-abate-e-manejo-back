package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/lib/pq"
)

type AbateRepository struct {
	db *sql.DB
}

func NewAbateRepository(db *sql.DB) AbateRepository {
	return AbateRepository{db: db}
}

func (a AbateRepository) CreateAbate(abate *domain.Abate) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	d := abate.DadosGeraisAbate
	err = tx.QueryRow(`
		INSERT INTO public.abate (
			fazenda_id, numero_lote, data_abate, nome_frigorifico,
			distancia_frigorifico, categoria_animal, preco_funrural,
			preco_sem_funrural, peso_total_fazenda,
			peso_total_frigorifico, balancao
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id`,
		d.FazendaID, d.NumeroLote, d.DataAbate, d.NomeFrigorifico,
		d.DistanciaFrigorifico, d.CategoriaAnimal, d.PrecoFunrural,
		d.PrecoSemFunrural, abate.EtapaFazenda.PesoTotal,
		abate.EtapaFrigorifico.PesoTotal, abate.EtapaFrigorifico.Balancao,
	).Scan(&abate.ID)
	if err != nil {
		return err
	}
	if err := insertEtapaFazenda(tx, abate.ID, abate.EtapaFazenda); err != nil {
		return err
	}
	if err := insertEtapaFrigorifico(tx, abate.ID, abate.EtapaFrigorifico); err != nil {
		return err
	}
	return tx.Commit()
}

func (a AbateRepository) FindAbateByID(id int) (*domain.Abate, error) {
	row := a.db.QueryRow(baseAbateQuery+` WHERE a.id = $1`, id)
	abate, err := scanAbate(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	abates := []domain.Abate{*abate}
	if err := loadAbatesDetails(a.db, abates); err != nil {
		return nil, err
	}
	return &abates[0], nil
}

func (a AbateRepository) FindAbates(filtro domain.FiltroAbate) ([]domain.Abate, error) {
	conditions := make([]string, 0, 5)
	args := make([]any, 0, 5)
	add := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}
	if filtro.ProprietarioID != nil {
		add("p.id = $%d", *filtro.ProprietarioID)
	}
	if filtro.FazendaID != nil {
		add("f.id = $%d", *filtro.FazendaID)
	}
	if filtro.NumeroLote != nil {
		add("a.numero_lote = $%d", *filtro.NumeroLote)
	}
	if filtro.DataInicio != nil {
		add("a.data_abate >= $%d", *filtro.DataInicio)
	}
	if filtro.DataFim != nil {
		add("a.data_abate < $%d", filtro.DataFim.AddDate(0, 0, 1))
	}

	query := baseAbateQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	limit := filtro.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if filtro.Offset < 0 {
		filtro.Offset = 0
	}
	args = append(args, limit, filtro.Offset)
	query += fmt.Sprintf(" ORDER BY a.data_abate DESC, a.id DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := a.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	abates := make([]domain.Abate, 0)
	for rows.Next() {
		abate, err := scanAbate(rows)
		if err != nil {
			return nil, err
		}
		abates = append(abates, *abate)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := loadAbatesDetails(a.db, abates); err != nil {
		return nil, err
	}
	return abates, nil
}

func (a AbateRepository) UpdateDadosGeraisAbate(abateID int, dados domain.DadosGeraisAbate) error {
	query := `
		WITH lock_identidade AS MATERIALIZED (
			SELECT pg_advisory_xact_lock(hashtextextended($10,0))
		), estado AS MATERIALIZED (
			SELECT a.id,
				(a.fazenda_id<>$1 OR a.numero_lote<>$2) AS caminho_alterado,
				EXISTS (SELECT 1 FROM public.abate_fotos af WHERE af.abate_id=a.id) AS possui_fotos
			FROM public.abate a
			CROSS JOIN lock_identidade
			WHERE a.id=$9
			FOR UPDATE OF a
		), atualizado AS (
			UPDATE public.abate SET
				fazenda_id=$1, numero_lote=$2, data_abate=$3, nome_frigorifico=$4,
				distancia_frigorifico=$5, categoria_animal=$6,
				preco_funrural=$7, preco_sem_funrural=$8
			WHERE id IN (
				SELECT id FROM estado WHERE NOT (caminho_alterado AND possui_fotos)
			)
			RETURNING id
		)
		SELECT EXISTS(SELECT 1 FROM estado),
			COALESCE((SELECT caminho_alterado AND possui_fotos FROM estado),false),
			EXISTS(SELECT 1 FROM atualizado)`
	var encontrado, bloqueado, atualizado bool
	err := a.db.QueryRow(query,
		dados.FazendaID, dados.NumeroLote, dados.DataAbate, dados.NomeFrigorifico,
		dados.DistanciaFrigorifico, dados.CategoriaAnimal,
		dados.PrecoFunrural, dados.PrecoSemFunrural, abateID,
		fmt.Sprintf("identity:abate:%d", abateID),
	).Scan(&encontrado, &bloqueado, &atualizado)
	if err != nil {
		return err
	}
	if !encontrado {
		return domain.ErrAbateNaoEncontrado
	}
	if bloqueado {
		return domain.ErrDadosGeraisComFotos
	}
	if !atualizado {
		return domain.ErrAbateNaoEncontrado
	}
	return nil
}

func (a AbateRepository) UpdateEtapaFazenda(abateID int, etapa domain.EtapaFazenda) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`UPDATE public.abate SET peso_total_fazenda=$1 WHERE id=$2`, etapa.PesoTotal, abateID)
	if err != nil {
		return err
	}
	if err := ensureAbateFound(result); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM public.abate_denticao WHERE abate_id=$1`, abateID); err != nil {
		return err
	}
	if err := insertEtapaFazenda(tx, abateID, etapa); err != nil {
		return err
	}
	return tx.Commit()
}

func (a AbateRepository) UpdateEtapaFrigorifico(abateID int, etapa domain.EtapaFrigorifico) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.Exec(`UPDATE public.abate SET peso_total_frigorifico=$1, balancao=$2 WHERE id=$3`, etapa.PesoTotal, etapa.Balancao, abateID)
	if err != nil {
		return err
	}
	if err := ensureAbateFound(result); err != nil {
		return err
	}
	for _, table := range []string{"abate_acabamento_carcaca", "abate_classificacao_frigorifico", "abate_distribuicao_peso"} {
		if _, err := tx.Exec("DELETE FROM public."+table+" WHERE abate_id=$1", abateID); err != nil {
			return err
		}
	}
	if err := insertEtapaFrigorifico(tx, abateID, etapa); err != nil {
		return err
	}
	return tx.Commit()
}

func (a AbateRepository) Delete(abateID int) ([]string, error) {
	tx, err := a.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, fmt.Sprintf("identity:abate:%d", abateID)); err != nil {
		return nil, err
	}
	fotos, err := queryFotos(tx, `SELECT id, abate_id, etapa, object_key, COALESCE(nome_original,''), COALESCE(content_type,'application/octet-stream'), COALESCE(tamanho,0), COALESCE(sha256,'') FROM public.abate_fotos WHERE abate_id=$1 FOR UPDATE`, abateID)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(fotos))
	seen := make(map[string]struct{}, len(fotos))
	for _, foto := range fotos {
		if _, ok := seen[foto.ObjectKey]; ok {
			continue
		}
		seen[foto.ObjectKey] = struct{}{}
		keys = append(keys, foto.ObjectKey)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, err := tx.Exec(`SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, key); err != nil {
			return nil, err
		}
	}
	if _, err := tx.Exec(`DELETE FROM public.abate_fotos WHERE abate_id=$1`, abateID); err != nil {
		return nil, err
	}
	for _, key := range keys {
		if _, err := tx.Exec(`INSERT INTO public.abate_storage_cleanup (object_key) VALUES ($1) ON CONFLICT (object_key) DO NOTHING`, key); err != nil {
			return nil, err
		}
	}
	result, err := tx.Exec(`DELETE FROM public.abate WHERE id=$1`, abateID)
	if err != nil {
		return nil, err
	}
	if err := ensureAbateFound(result); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return keys, nil
}

func (a AbateRepository) GetContextoFoto(abateID int) (*domain.ContextoFotoAbate, error) {
	contexto := &domain.ContextoFotoAbate{}
	err := a.db.QueryRow(`
		SELECT p.id, a.numero_lote
		FROM public.abate a
		JOIN public.fazenda f ON f.id=a.fazenda_id
		JOIN public.proprietario p ON p.id=f.id_proprietario
		WHERE a.id=$1`, abateID).Scan(&contexto.ProprietarioID, &contexto.NumeroLote)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return contexto, err
}

func (a AbateRepository) FindFotoByHash(abateID int, etapa, hash string) (*domain.FotoAbate, error) {
	return a.findFoto(`WHERE abate_id=$1 AND etapa=$2 AND sha256=$3`, abateID, etapa, hash)
}

func (a AbateRepository) CreateFoto(foto *domain.FotoAbate) error {
	return a.db.QueryRow(`
		INSERT INTO public.abate_fotos
			(abate_id, etapa, object_key, nome_original, content_type, tamanho, sha256)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (abate_id, etapa, sha256) DO UPDATE
		SET nome_original=EXCLUDED.nome_original
		RETURNING id`,
		foto.AbateID, foto.Etapa, foto.ObjectKey, foto.NomeOriginal,
		foto.ContentType, foto.Tamanho, foto.SHA256,
	).Scan(&foto.ID)
}

func (a AbateRepository) FindFotoByID(fotoID int) (*domain.FotoAbate, error) {
	return a.findFoto(`WHERE id=$1`, fotoID)
}

func (a AbateRepository) DeleteFoto(fotoID int) (*domain.FotoAbate, error) {
	tx, err := a.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	foto := &domain.FotoAbate{}
	err = tx.QueryRow(`
		SELECT id, abate_id, etapa, object_key, COALESCE(nome_original,''), COALESCE(content_type,'application/octet-stream'), COALESCE(tamanho,0), COALESCE(sha256,'')
		FROM public.abate_fotos WHERE id=$1 FOR UPDATE`, fotoID,
	).Scan(&foto.ID, &foto.AbateID, &foto.Etapa, &foto.ObjectKey, &foto.NomeOriginal, &foto.ContentType, &foto.Tamanho, &foto.SHA256)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, foto.ObjectKey); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`DELETE FROM public.abate_fotos WHERE id=$1`, fotoID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`INSERT INTO public.abate_storage_cleanup (object_key) VALUES ($1) ON CONFLICT (object_key) DO NOTHING`, foto.ObjectKey); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return foto, nil
}

func (a AbateRepository) CountFotosByObjectKey(objectKey string) (int, error) {
	var count int
	err := a.db.QueryRow(`SELECT COUNT(*) FROM public.abate_fotos WHERE object_key=$1`, objectKey).Scan(&count)
	return count, err
}

func (a AbateRepository) EnqueueObjectCleanup(objectKey string) error {
	_, err := a.db.Exec(`INSERT INTO public.abate_storage_cleanup (object_key) VALUES ($1) ON CONFLICT (object_key) DO NOTHING`, objectKey)
	return err
}

func (a AbateRepository) FindPendingObjectCleanup(limit int) ([]string, error) {
	rows, err := a.db.Query(`SELECT object_key FROM public.abate_storage_cleanup WHERE next_attempt_at<=CURRENT_TIMESTAMP ORDER BY next_attempt_at,created_at LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	keys := make([]string, 0)
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func (a AbateRepository) MarkObjectCleanupFailed(objectKey, message string) error {
	_, err := a.db.Exec(`
		UPDATE public.abate_storage_cleanup
		SET attempts=attempts+1,
			last_error=$2,
			next_attempt_at=CURRENT_TIMESTAMP +
				(LEAST(3600, power(2, LEAST(attempts+1,10))::int) * interval '1 second')
		WHERE object_key=$1`, objectKey, message)
	return err
}

func (a AbateRepository) DeletePendingObjectCleanup(objectKey string) error {
	_, err := a.db.Exec(`DELETE FROM public.abate_storage_cleanup WHERE object_key=$1`, objectKey)
	return err
}

func (a AbateRepository) AcquireObjectLock(ctx context.Context, objectKey string) (func() error, error) {
	conn, err := a.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock(hashtextextended($1, 0))`, objectKey); err != nil {
		conn.Close()
		return nil, err
	}
	return func() error {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		defer conn.Close()
		var unlocked bool
		return conn.QueryRowContext(unlockCtx, `SELECT pg_advisory_unlock(hashtextextended($1, 0))`, objectKey).Scan(&unlocked)
	}, nil
}

func (a AbateRepository) findFoto(where string, args ...any) (*domain.FotoAbate, error) {
	foto := &domain.FotoAbate{}
	err := a.db.QueryRow(`
		SELECT id, abate_id, etapa, object_key, COALESCE(nome_original,''), COALESCE(content_type,'application/octet-stream'), COALESCE(tamanho,0), COALESCE(sha256,'')
		FROM public.abate_fotos `+where, args...).Scan(
		&foto.ID, &foto.AbateID, &foto.Etapa, &foto.ObjectKey,
		&foto.NomeOriginal, &foto.ContentType, &foto.Tamanho, &foto.SHA256,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return foto, err
}

const baseAbateQuery = `
	SELECT a.id, p.id, p.nome, f.nome, a.fazenda_id, a.numero_lote,
		a.data_abate, a.nome_frigorifico, a.distancia_frigorifico,
		a.categoria_animal, COALESCE(a.preco_funrural,0),
		COALESCE(a.preco_sem_funrural,0), COALESCE(a.peso_total_fazenda,0),
		COALESCE(a.peso_total_frigorifico,0), COALESCE(a.balancao,0)
	FROM public.abate a
	JOIN public.fazenda f ON f.id=a.fazenda_id
	JOIN public.proprietario p ON p.id=f.id_proprietario`

type scanner interface {
	Scan(dest ...any) error
}

func scanAbate(row scanner) (*domain.Abate, error) {
	abate := &domain.Abate{}
	err := row.Scan(
		&abate.ID, &abate.ProprietarioID, &abate.NomeProprietario, &abate.NomeFazenda,
		&abate.DadosGeraisAbate.FazendaID, &abate.DadosGeraisAbate.NumeroLote,
		&abate.DadosGeraisAbate.DataAbate, &abate.DadosGeraisAbate.NomeFrigorifico,
		&abate.DadosGeraisAbate.DistanciaFrigorifico, &abate.DadosGeraisAbate.CategoriaAnimal,
		&abate.DadosGeraisAbate.PrecoFunrural, &abate.DadosGeraisAbate.PrecoSemFunrural,
		&abate.EtapaFazenda.PesoTotal, &abate.EtapaFrigorifico.PesoTotal,
		&abate.EtapaFrigorifico.Balancao,
	)
	return abate, err
}

type queryer interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

func loadAbatesDetails(db queryer, abates []domain.Abate) error {
	if len(abates) == 0 {
		return nil
	}
	byID := make(map[int]*domain.Abate, len(abates))
	ids := make([]int, len(abates))
	for i := range abates {
		ids[i] = abates[i].ID
		byID[abates[i].ID] = &abates[i]
	}

	rows, err := db.Query(`SELECT abate_id, denticao, qtd_animais FROM public.abate_denticao WHERE abate_id=ANY($1) ORDER BY abate_id,id`, pq.Array(ids))
	if err != nil {
		return err
	}
	for rows.Next() {
		var abateID int
		item := domain.QtdDenticao{}
		if err := rows.Scan(&abateID, &item.QtdDenticao, &item.QtdAnimais); err != nil {
			rows.Close()
			return err
		}
		byID[abateID].EtapaFazenda.QuantidadeAnimal = append(byID[abateID].EtapaFazenda.QuantidadeAnimal, item)
	}
	if err := closeRows(rows); err != nil {
		return err
	}

	rows, err = db.Query(`SELECT abate_id, acabamento, qtd_animais FROM public.abate_acabamento_carcaca WHERE abate_id=ANY($1) ORDER BY abate_id,id`, pq.Array(ids))
	if err != nil {
		return err
	}
	for rows.Next() {
		var abateID int
		item := domain.AcabamentoCarcaca{}
		if err := rows.Scan(&abateID, &item.Acabamento, &item.QtdAnimais); err != nil {
			rows.Close()
			return err
		}
		byID[abateID].EtapaFrigorifico.AcabamentoCarcaca = append(byID[abateID].EtapaFrigorifico.AcabamentoCarcaca, item)
	}
	if err := closeRows(rows); err != nil {
		return err
	}

	rows, err = db.Query(`SELECT abate_id, classificacao, qtd_animais FROM public.abate_classificacao_frigorifico WHERE abate_id=ANY($1) ORDER BY abate_id,id`, pq.Array(ids))
	if err != nil {
		return err
	}
	for rows.Next() {
		var abateID int
		item := domain.ClassificacaoFrigorifico{}
		if err := rows.Scan(&abateID, &item.Classificacao, &item.QtdAnimais); err != nil {
			rows.Close()
			return err
		}
		byID[abateID].EtapaFrigorifico.ClassificacaoFrigorifico = append(byID[abateID].EtapaFrigorifico.ClassificacaoFrigorifico, item)
	}
	if err := closeRows(rows); err != nil {
		return err
	}

	rows, err = db.Query(`SELECT abate_id, classificacao, qtd_animais, peso_total FROM public.abate_distribuicao_peso WHERE abate_id=ANY($1) ORDER BY abate_id,id`, pq.Array(ids))
	if err != nil {
		return err
	}
	for rows.Next() {
		var abateID int
		item := domain.DistribuicaoPeso{}
		if err := rows.Scan(&abateID, &item.Classificacao, &item.QtdAnimais, &item.PesoTotal); err != nil {
			rows.Close()
			return err
		}
		byID[abateID].EtapaFrigorifico.DistribuicaoPeso = append(byID[abateID].EtapaFrigorifico.DistribuicaoPeso, item)
	}
	if err := closeRows(rows); err != nil {
		return err
	}

	fotos, err := queryFotos(db, `SELECT id, abate_id, etapa, object_key, COALESCE(nome_original,''), COALESCE(content_type,'application/octet-stream'), COALESCE(tamanho,0), COALESCE(sha256,'') FROM public.abate_fotos WHERE abate_id=ANY($1) ORDER BY abate_id,id`, pq.Array(ids))
	if err != nil {
		return err
	}
	for _, foto := range fotos {
		if foto.Etapa == domain.EtapaFazendaFoto {
			byID[foto.AbateID].EtapaFazenda.Fotos = append(byID[foto.AbateID].EtapaFazenda.Fotos, foto)
		} else {
			byID[foto.AbateID].EtapaFrigorifico.Fotos = append(byID[foto.AbateID].EtapaFrigorifico.Fotos, foto)
		}
	}
	return nil
}

func queryFotos(db queryer, query string, args ...any) ([]domain.FotoAbate, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fotos := make([]domain.FotoAbate, 0)
	for rows.Next() {
		foto := domain.FotoAbate{}
		if err := rows.Scan(&foto.ID, &foto.AbateID, &foto.Etapa, &foto.ObjectKey, &foto.NomeOriginal, &foto.ContentType, &foto.Tamanho, &foto.SHA256); err != nil {
			return nil, err
		}
		fotos = append(fotos, foto)
	}
	return fotos, rows.Err()
}

func closeRows(rows *sql.Rows) error {
	err := rows.Err()
	rows.Close()
	return err
}

func insertEtapaFazenda(tx *sql.Tx, abateID int, etapa domain.EtapaFazenda) error {
	for _, item := range etapa.QuantidadeAnimal {
		if _, err := tx.Exec(`INSERT INTO public.abate_denticao (abate_id,denticao,qtd_animais) VALUES ($1,$2,$3)`, abateID, item.QtdDenticao, item.QtdAnimais); err != nil {
			return err
		}
	}
	return nil
}

func insertEtapaFrigorifico(tx *sql.Tx, abateID int, etapa domain.EtapaFrigorifico) error {
	for _, item := range etapa.AcabamentoCarcaca {
		if _, err := tx.Exec(`INSERT INTO public.abate_acabamento_carcaca (abate_id,acabamento,qtd_animais) VALUES ($1,$2,$3)`, abateID, item.Acabamento, item.QtdAnimais); err != nil {
			return err
		}
	}
	for _, item := range etapa.ClassificacaoFrigorifico {
		if _, err := tx.Exec(`INSERT INTO public.abate_classificacao_frigorifico (abate_id,classificacao,qtd_animais) VALUES ($1,$2,$3)`, abateID, item.Classificacao, item.QtdAnimais); err != nil {
			return err
		}
	}
	for _, item := range etapa.DistribuicaoPeso {
		if _, err := tx.Exec(`INSERT INTO public.abate_distribuicao_peso (abate_id,classificacao,qtd_animais,peso_total) VALUES ($1,$2,$3,$4)`, abateID, item.Classificacao, item.QtdAnimais, item.PesoTotal); err != nil {
			return err
		}
	}
	return nil
}

func ensureAbateFound(result sql.Result) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrAbateNaoEncontrado
	}
	return nil
}
