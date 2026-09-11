package service

import (
	"errors"
	"reflect"
	"testing"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
)

type userPortStub struct {
	createUser     func(*domain.User) error
	getUsers       func() (*[]domain.User, error)
	getUserByEmail func(string) (*domain.User, error)
	updateUser     func(*domain.User) error
	deleteUser     func(int) error
}

func (s userPortStub) CreateUser(user *domain.User) error { return s.createUser(user) }
func (s userPortStub) GetUsers() (*[]domain.User, error)  { return s.getUsers() }
func (s userPortStub) GetUserByEmail(email string) (*domain.User, error) {
	return s.getUserByEmail(email)
}
func (s userPortStub) UpdateUser(user *domain.User) error { return s.updateUser(user) }
func (s userPortStub) DeleteUser(id int) error            { return s.deleteUser(id) }

type hashPortStub struct {
	hash    func(string) (string, error)
	compare func(string, string) error
}

func (s hashPortStub) Hash(password string) (string, error) { return s.hash(password) }
func (s hashPortStub) Compare(password, hashedPassword string) error {
	return s.compare(password, hashedPassword)
}

type tokenPortStub struct {
	generate func(domain.User) (string, error)
}

func (s tokenPortStub) Generate(user domain.User) (string, error) { return s.generate(user) }
func (s tokenPortStub) Validate(string) (string, error)           { panic("unexpected call") }

func TestUserServiceLogin(t *testing.T) {
	t.Run("retorna erro da consulta", func(t *testing.T) {
		expectedErr := errors.New("erro de consulta")
		service := NewUserService(userPortStub{
			getUserByEmail: func(email string) (*domain.User, error) {
				if email != "user@example.com" {
					t.Fatalf("email recebido = %q", email)
				}
				return nil, expectedErr
			},
		}, hashPortStub{}, tokenPortStub{})

		token, err := service.Login("user@example.com", "senha")
		if token != nil {
			t.Fatalf("token = %q, esperado nil", *token)
		}
		if !errors.Is(err, expectedErr) {
			t.Fatalf("erro = %v, esperado %v", err, expectedErr)
		}
	})

	t.Run("rejeita usuario inexistente", func(t *testing.T) {
		service := NewUserService(userPortStub{
			getUserByEmail: func(string) (*domain.User, error) { return nil, nil },
		}, hashPortStub{}, tokenPortStub{})

		token, err := service.Login("missing@example.com", "senha")
		if token != nil {
			t.Fatalf("token = %q, esperado nil", *token)
		}
		if !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("erro = %v, esperado %v", err, domain.ErrInvalidCredentials)
		}
	})

	t.Run("rejeita senha invalida", func(t *testing.T) {
		user := domain.User{Email: "user@example.com", Password: "hash"}
		service := NewUserService(userPortStub{
			getUserByEmail: func(string) (*domain.User, error) { return &user, nil },
		}, hashPortStub{
			compare: func(password, hash string) error {
				if password != "senha" || hash != "hash" {
					t.Fatalf("Compare(%q, %q)", password, hash)
				}
				return errors.New("senha nao confere")
			},
		}, tokenPortStub{})

		token, err := service.Login(user.Email, "senha")
		if token != nil {
			t.Fatalf("token = %q, esperado nil", *token)
		}
		if !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("erro = %v, esperado %v", err, domain.ErrInvalidCredentials)
		}
	})

	t.Run("retorna erro ao gerar token", func(t *testing.T) {
		user := domain.User{ID: 7, Email: "user@example.com", Password: "hash"}
		expectedErr := errors.New("erro ao gerar token")
		service := NewUserService(userPortStub{
			getUserByEmail: func(string) (*domain.User, error) { return &user, nil },
		}, hashPortStub{
			compare: func(string, string) error { return nil },
		}, tokenPortStub{
			generate: func(got domain.User) (string, error) {
				if !reflect.DeepEqual(got, user) {
					t.Fatalf("usuario recebido = %#v, esperado %#v", got, user)
				}
				return "", expectedErr
			},
		})

		token, err := service.Login(user.Email, "senha")
		if token != nil {
			t.Fatalf("token = %q, esperado nil", *token)
		}
		if !errors.Is(err, expectedErr) {
			t.Fatalf("erro = %v, esperado %v", err, expectedErr)
		}
	})

	t.Run("retorna token", func(t *testing.T) {
		user := domain.User{ID: 7, Email: "user@example.com", Password: "hash"}
		service := NewUserService(userPortStub{
			getUserByEmail: func(email string) (*domain.User, error) {
				if email != user.Email {
					t.Fatalf("email recebido = %q", email)
				}
				return &user, nil
			},
		}, hashPortStub{
			compare: func(password, hash string) error {
				if password != "senha" || hash != user.Password {
					t.Fatalf("Compare(%q, %q)", password, hash)
				}
				return nil
			},
		}, tokenPortStub{
			generate: func(got domain.User) (string, error) {
				if !reflect.DeepEqual(got, user) {
					t.Fatalf("usuario recebido = %#v, esperado %#v", got, user)
				}
				return "token-gerado", nil
			},
		})

		token, err := service.Login(user.Email, "senha")
		if err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if token == nil || *token != "token-gerado" {
			t.Fatalf("token = %v, esperado token-gerado", token)
		}
	})
}

func TestUserServiceCreateUser(t *testing.T) {
	t.Run("retorna erro da consulta", func(t *testing.T) {
		expectedErr := errors.New("erro de consulta")
		service := NewUserService(userPortStub{
			getUserByEmail: func(string) (*domain.User, error) { return nil, expectedErr },
		}, hashPortStub{}, tokenPortStub{})

		err := service.CreateUser(&domain.User{Email: "user@example.com"})
		if !errors.Is(err, expectedErr) {
			t.Fatalf("erro = %v, esperado %v", err, expectedErr)
		}
	})

	t.Run("rejeita email existente", func(t *testing.T) {
		existing := domain.User{ID: 1}
		service := NewUserService(userPortStub{
			getUserByEmail: func(string) (*domain.User, error) { return &existing, nil },
		}, hashPortStub{}, tokenPortStub{})

		err := service.CreateUser(&domain.User{Email: "user@example.com"})
		if !errors.Is(err, domain.ErrEmailAlreadyExists) {
			t.Fatalf("erro = %v, esperado %v", err, domain.ErrEmailAlreadyExists)
		}
	})

	t.Run("retorna erro do hash", func(t *testing.T) {
		expectedErr := errors.New("erro de hash")
		user := &domain.User{Email: "user@example.com", Password: "senha"}
		service := NewUserService(userPortStub{
			getUserByEmail: func(string) (*domain.User, error) { return nil, nil },
		}, hashPortStub{
			hash: func(password string) (string, error) {
				if password != "senha" {
					t.Fatalf("senha recebida = %q", password)
				}
				return "", expectedErr
			},
		}, tokenPortStub{})

		err := service.CreateUser(user)
		if !errors.Is(err, expectedErr) {
			t.Fatalf("erro = %v, esperado %v", err, expectedErr)
		}
		if user.Password != "senha" {
			t.Fatalf("senha foi alterada para %q", user.Password)
		}
	})

	t.Run("salva usuario com senha transformada", func(t *testing.T) {
		user := &domain.User{Email: "user@example.com", Password: "senha"}
		service := NewUserService(userPortStub{
			getUserByEmail: func(email string) (*domain.User, error) {
				if email != user.Email {
					t.Fatalf("email recebido = %q", email)
				}
				return nil, nil
			},
			createUser: func(got *domain.User) error {
				if got != user {
					t.Fatal("CreateUser recebeu outro ponteiro")
				}
				if got.Password != "hash-da-senha" {
					t.Fatalf("senha salva = %q", got.Password)
				}
				return nil
			},
		}, hashPortStub{
			hash: func(password string) (string, error) {
				if password != "senha" {
					t.Fatalf("senha recebida = %q", password)
				}
				return "hash-da-senha", nil
			},
		}, tokenPortStub{})

		if err := service.CreateUser(user); err != nil {
			t.Fatalf("erro inesperado: %v", err)
		}
		if user.Password != "hash-da-senha" {
			t.Fatalf("senha final = %q", user.Password)
		}
	})

	t.Run("propaga erro ao salvar", func(t *testing.T) {
		expectedErr := errors.New("erro ao salvar")
		service := NewUserService(userPortStub{
			getUserByEmail: func(string) (*domain.User, error) { return nil, nil },
			createUser:     func(*domain.User) error { return expectedErr },
		}, hashPortStub{
			hash: func(string) (string, error) { return "hash", nil },
		}, tokenPortStub{})

		err := service.CreateUser(&domain.User{Email: "user@example.com", Password: "senha"})
		if !errors.Is(err, expectedErr) {
			t.Fatalf("erro = %v, esperado %v", err, expectedErr)
		}
	})
}

func TestUserServiceCRUD(t *testing.T) {
	users := []domain.User{{ID: 1, Nome: "Usuario"}}
	user := &domain.User{ID: 2, Nome: "Atualizado"}

	t.Run("lista usuarios", func(t *testing.T) {
		testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
			service := NewUserService(userPortStub{
				getUsers: func() (*[]domain.User, error) { return &users, expectedErr },
			}, hashPortStub{}, tokenPortStub{})

			got, err := service.GetUsers()
			if got != &users {
				t.Fatalf("resultado = %v", got)
			}
			return err
		})
	})

	t.Run("atualiza usuario", func(t *testing.T) {
		testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
			service := NewUserService(userPortStub{
				updateUser: func(got *domain.User) error {
					if got != user {
						t.Fatal("UpdateUser recebeu outro ponteiro")
					}
					return expectedErr
				},
			}, hashPortStub{}, tokenPortStub{})
			return service.UpdateUser(user)
		})
	})

	t.Run("exclui usuario", func(t *testing.T) {
		testPassthroughErrors(t, func(t *testing.T, expectedErr error) error {
			service := NewUserService(userPortStub{
				deleteUser: func(id int) error {
					if id != 42 {
						t.Fatalf("id recebido = %d", id)
					}
					return expectedErr
				},
			}, hashPortStub{}, tokenPortStub{})
			return service.DeleteUser(42)
		})
	})
}
