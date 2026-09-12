package repository

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

func TestUserRepositoryCreateUser(t *testing.T) {
	user := &domain.User{Nome: "Maria", Email: "maria@example.com", Password: "hash", Role: 2, Ativo: true}
	query := regexp.QuoteMeta(`INSERT INTO public.usuario (nome, email, senha, role, ativo) VALUES ($1, $2, $3, $4, $5)`)

	t.Run("insere todos os campos", func(t *testing.T) {
		db, mock := newUserRepositoryMock(t)
		mock.ExpectExec(query).
			WithArgs(user.Nome, user.Email, user.Password, user.Role, user.Ativo).
			WillReturnResult(sqlmock.NewResult(1, 1))
		if err := NewUserRepository(db).CreateUser(user); err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		assertUserRepositoryExpectations(t, mock)
	})

	t.Run("propaga erro", func(t *testing.T) {
		db, mock := newUserRepositoryMock(t)
		expectedErr := errors.New("erro de insert")
		mock.ExpectExec(query).
			WithArgs(user.Nome, user.Email, user.Password, user.Role, user.Ativo).
			WillReturnError(expectedErr)
		err := NewUserRepository(db).CreateUser(user)
		if !errors.Is(err, expectedErr) {
			t.Fatalf("erro = %v, esperado %v", err, expectedErr)
		}
		assertUserRepositoryExpectations(t, mock)
	})
}

func TestUserRepositoryGetUsers(t *testing.T) {
	query := regexp.QuoteMeta(`SELECT id, nome, email, senha, role, ativo FROM public.usuario`)

	t.Run("lista usuarios ativos e inativos", func(t *testing.T) {
		db, mock := newUserRepositoryMock(t)
		rows := sqlmock.NewRows([]string{"id", "nome", "email", "senha", "role", "ativo"}).
			AddRow(1, "Ativo", "ativo@example.com", "hash-1", 1, true).
			AddRow(2, "Inativo", "inativo@example.com", "hash-2", 2, false)
		mock.ExpectQuery(query).WillReturnRows(rows)

		users, err := NewUserRepository(db).GetUsers()
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if users == nil || len(*users) != 2 || (*users)[0].ID != 1 || !(*users)[0].Ativo || (*users)[1].Ativo || (*users)[1].Password != "hash-2" {
			t.Fatalf("usuários = %+v", users)
		}
		assertUserRepositoryExpectations(t, mock)
	})

	t.Run("retorna lista vazia", func(t *testing.T) {
		db, mock := newUserRepositoryMock(t)
		mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows([]string{"id", "nome", "email", "senha", "role", "ativo"}))
		users, err := NewUserRepository(db).GetUsers()
		if err != nil || users == nil || len(*users) != 0 {
			t.Fatalf("usuários = %+v, erro = %v", users, err)
		}
		assertUserRepositoryExpectations(t, mock)
	})

	t.Run("propaga erro da consulta", func(t *testing.T) {
		db, mock := newUserRepositoryMock(t)
		expectedErr := errors.New("erro de query")
		mock.ExpectQuery(query).WillReturnError(expectedErr)
		users, err := NewUserRepository(db).GetUsers()
		if users != nil || !errors.Is(err, expectedErr) {
			t.Fatalf("usuários = %+v, erro = %v", users, err)
		}
		assertUserRepositoryExpectations(t, mock)
	})

	t.Run("propaga erro de scan", func(t *testing.T) {
		db, mock := newUserRepositoryMock(t)
		rows := sqlmock.NewRows([]string{"id", "nome", "email", "senha", "role", "ativo"}).
			AddRow("id-invalido", "Nome", "email@example.com", "hash", 1, true)
		mock.ExpectQuery(query).WillReturnRows(rows)
		users, err := NewUserRepository(db).GetUsers()
		if users != nil || err == nil {
			t.Fatalf("usuários = %+v, erro = %v", users, err)
		}
		assertUserRepositoryExpectations(t, mock)
	})

	t.Run("propaga erro tardio sem lista parcial", func(t *testing.T) {
		db, mock := newUserRepositoryMock(t)
		expectedErr := errors.New("erro durante iteração")
		rows := sqlmock.NewRows([]string{"id", "nome", "email", "senha", "role", "ativo"}).
			AddRow(1, "Primeiro", "primeiro@example.com", "hash-1", 1, true).
			AddRow(2, "Segundo", "segundo@example.com", "hash-2", 2, false).
			RowError(1, expectedErr)
		mock.ExpectQuery(query).WillReturnRows(rows)

		users, err := NewUserRepository(db).GetUsers()
		if users != nil || !errors.Is(err, expectedErr) {
			t.Fatalf("usuários = %+v, erro = %v, esperado %v", users, err, expectedErr)
		}
		assertUserRepositoryExpectations(t, mock)
	})
}

func TestUserRepositoryGetUserByEmail(t *testing.T) {
	query := regexp.QuoteMeta(`SELECT id, nome, email, senha, role, ativo FROM public.usuario WHERE email = $1`)
	email := "user@example.com"

	t.Run("retorna usuario", func(t *testing.T) {
		db, mock := newUserRepositoryMock(t)
		mock.ExpectQuery(query).
			WithArgs(email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "nome", "email", "senha", "role", "ativo"}).
				AddRow(7, "Usuário", email, "hash", 3, true))
		user, err := NewUserRepository(db).GetUserByEmail(email)
		if err != nil || user == nil || user.ID != 7 || user.Email != email || user.Password != "hash" || !user.Ativo {
			t.Fatalf("usuário = %+v, erro = %v", user, err)
		}
		assertUserRepositoryExpectations(t, mock)
	})

	t.Run("nenhum registro", func(t *testing.T) {
		db, mock := newUserRepositoryMock(t)
		mock.ExpectQuery(query).WithArgs(email).WillReturnError(sql.ErrNoRows)
		user, err := NewUserRepository(db).GetUserByEmail(email)
		if user != nil || err != nil {
			t.Fatalf("usuário = %+v, erro = %v", user, err)
		}
		assertUserRepositoryExpectations(t, mock)
	})

	t.Run("propaga erro de consulta", func(t *testing.T) {
		db, mock := newUserRepositoryMock(t)
		expectedErr := errors.New("erro do banco")
		mock.ExpectQuery(query).WithArgs(email).WillReturnError(expectedErr)
		user, err := NewUserRepository(db).GetUserByEmail(email)
		if user != nil || !errors.Is(err, expectedErr) {
			t.Fatalf("usuário = %+v, erro = %v, esperado %v", user, err, expectedErr)
		}
		assertUserRepositoryExpectations(t, mock)
	})
}

func TestUserRepositoryMutations(t *testing.T) {
	user := &domain.User{ID: 7, Nome: "Novo Nome", Email: "novo@example.com", Role: 3}
	tests := []struct {
		name   string
		query  string
		args   []driver.Value
		invoke func(UserRepository) error
	}{
		{
			name:  "update",
			query: `UPDATE public.usuario SET nome = $1, email = $2, role = $3 WHERE id = $4`,
			args:  []driver.Value{user.Nome, user.Email, user.Role, user.ID},
			invoke: func(repository UserRepository) error {
				return repository.UpdateUser(user)
			},
		},
		{
			name:  "activate",
			query: `UPDATE public.usuario SET ativo = true WHERE id = $1`,
			args:  []driver.Value{7},
			invoke: func(repository UserRepository) error {
				return repository.ActivateUser(7)
			},
		},
		{
			name:  "deactivate",
			query: `UPDATE public.usuario SET ativo = false WHERE id = $1`,
			args:  []driver.Value{7},
			invoke: func(repository UserRepository) error {
				return repository.DeactivateUser(7)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name+" success", func(t *testing.T) {
			db, mock := newUserRepositoryMock(t)
			mock.ExpectExec(regexp.QuoteMeta(tc.query)).WithArgs(tc.args...).WillReturnResult(sqlmock.NewResult(0, 1))
			if err := tc.invoke(NewUserRepository(db)); err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			assertUserRepositoryExpectations(t, mock)
		})

		t.Run(tc.name+" exec error", func(t *testing.T) {
			db, mock := newUserRepositoryMock(t)
			expectedErr := errors.New("erro de exec")
			mock.ExpectExec(regexp.QuoteMeta(tc.query)).WithArgs(tc.args...).WillReturnError(expectedErr)
			err := tc.invoke(NewUserRepository(db))
			if !errors.Is(err, expectedErr) {
				t.Fatalf("erro = %v, esperado %v", err, expectedErr)
			}
			assertUserRepositoryExpectations(t, mock)
		})

		t.Run(tc.name+" rows affected error", func(t *testing.T) {
			db, mock := newUserRepositoryMock(t)
			expectedErr := errors.New("erro no resultado")
			mock.ExpectExec(regexp.QuoteMeta(tc.query)).WithArgs(tc.args...).WillReturnResult(sqlmock.NewErrorResult(expectedErr))
			err := tc.invoke(NewUserRepository(db))
			if !errors.Is(err, expectedErr) {
				t.Fatalf("erro = %v, esperado %v", err, expectedErr)
			}
			assertUserRepositoryExpectations(t, mock)
		})

		t.Run(tc.name+" zero rows", func(t *testing.T) {
			db, mock := newUserRepositoryMock(t)
			mock.ExpectExec(regexp.QuoteMeta(tc.query)).WithArgs(tc.args...).WillReturnResult(sqlmock.NewResult(0, 0))
			err := tc.invoke(NewUserRepository(db))
			if err == nil || !strings.Contains(err.Error(), "id 7") {
				t.Fatalf("erro = %v", err)
			}
			assertUserRepositoryExpectations(t, mock)
		})
	}
}

func newUserRepositoryMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func assertUserRepositoryExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
