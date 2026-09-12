package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(
	db *sql.DB,
) UserRepository {
	return UserRepository{
		db: db,
	}
}

func (u UserRepository) CreateUser(user *domain.User) error {
	query := `INSERT INTO public.usuario (nome, email, senha, role, ativo) VALUES ($1, $2, $3, $4, $5)`

	_, err := u.db.Exec(query, user.Nome, user.Email, user.Password, user.Role, user.Ativo)
	if err != nil {
		return err
	}

	return nil
}

func (u UserRepository) GetUsers() (*[]domain.User, error) {
	query := `SELECT id, nome, email, senha, role, ativo FROM public.usuario`

	rows, err := u.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []domain.User{}
	for rows.Next() {
		user := domain.User{}
		if err := rows.Scan(&user.ID, &user.Nome, &user.Email, &user.Password, &user.Role, &user.Ativo); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &users, nil
}

func (u UserRepository) GetUserByEmail(email string) (*domain.User, error) {
	query := `SELECT id, nome, email, senha, role, ativo FROM public.usuario WHERE email = $1`

	row := u.db.QueryRow(query, email)

	user := &domain.User{}

	err := row.Scan(&user.ID, &user.Nome, &user.Email, &user.Password, &user.Role, &user.Ativo)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (u UserRepository) UpdateUser(user *domain.User) error {
	query := `UPDATE public.usuario SET nome = $1, email = $2, role = $3 WHERE id = $4`

	result, err := u.db.Exec(query, user.Nome, user.Email, user.Role, user.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("usuário com id %d não encontrado", user.ID)
	}

	return nil
}

func (u UserRepository) ActivateUser(id int) error {
	query := `UPDATE public.usuario SET ativo = true WHERE id = $1`

	result, err := u.db.Exec(query, id)
	if err != nil {
		return err
	}

	return ensureDataFound(result, id)
}

func (u UserRepository) DeactivateUser(id int) error {
	query := `UPDATE public.usuario SET ativo = false WHERE id = $1`

	result, err := u.db.Exec(query, id)
	if err != nil {
		return err
	}

	return ensureDataFound(result, id)
}
