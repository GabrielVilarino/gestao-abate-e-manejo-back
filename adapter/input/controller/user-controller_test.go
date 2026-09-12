package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/gin-gonic/gin"
)

type userUseCaseStub struct {
	login      func(string, string) (*string, error)
	create     func(*domain.User) error
	getAll     func() (*[]domain.User, error)
	update     func(*domain.User) error
	activate   func(int) error
	deactivate func(int) error
}

func (s userUseCaseStub) Login(email, password string) (*string, error) {
	return s.login(email, password)
}
func (s userUseCaseStub) CreateUser(user *domain.User) error { return s.create(user) }
func (s userUseCaseStub) GetUsers() (*[]domain.User, error)  { return s.getAll() }
func (s userUseCaseStub) UpdateUser(user *domain.User) error { return s.update(user) }
func (s userUseCaseStub) ActivateUser(id int) error          { return s.activate(id) }
func (s userUseCaseStub) DeactivateUser(id int) error        { return s.deactivate(id) }

func TestUserControllerLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("autentica e cria cookie seguro", func(t *testing.T) {
		token := "token-gerado"
		controller := NewUserController(userUseCaseStub{
			login: func(email, password string) (*string, error) {
				if email != "user@example.com" || password != "senha" {
					t.Fatalf("credenciais recebidas = %q, %q", email, password)
				}
				return &token, nil
			},
		})
		w, c := userControllerContext(http.MethodPost, `{"email":"user@example.com","password":"senha"}`)
		controller.Login(c)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
		cookies := w.Result().Cookies()
		if len(cookies) != 1 {
			t.Fatalf("cookies = %v", cookies)
		}
		cookie := cookies[0]
		if cookie.Name != "auth" || cookie.Value != token || cookie.MaxAge != 259200 || cookie.Path != "/" || !cookie.Secure || !cookie.HttpOnly {
			t.Fatalf("cookie inesperado: %+v", cookie)
		}
	})

	t.Run("rejeita payload invalido", func(t *testing.T) {
		controller := NewUserController(userUseCaseStub{
			login: func(string, string) (*string, error) {
				t.Fatal("use case não deveria ser chamado")
				return nil, nil
			},
		})
		w, c := userControllerContext(http.MethodPost, `{"email":"user@example.com"}`)
		controller.Login(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, esperado %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("retorna unauthorized para erro de autenticacao", func(t *testing.T) {
		expectedErr := errors.New("acesso negado")
		controller := NewUserController(userUseCaseStub{
			login: func(string, string) (*string, error) { return nil, expectedErr },
		})
		w, c := userControllerContext(http.MethodPost, `{"email":"user@example.com","password":"senha"}`)
		controller.Login(c)
		if w.Code != http.StatusUnauthorized || !strings.Contains(w.Body.String(), expectedErr.Error()) {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
	})
}

func TestUserControllerLogoutExpiresCookie(t *testing.T) {
	controller := &UserController{}
	w, c := userControllerContext(http.MethodPost, "")
	controller.Logout(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "auth" || cookies[0].Value != "" || cookies[0].MaxAge != -1 || !cookies[0].Secure || !cookies[0].HttpOnly {
		t.Fatalf("cookie inesperado: %+v", cookies)
	}
}

func TestUserControllerCreateUser(t *testing.T) {
	t.Run("cria usuario ativo", func(t *testing.T) {
		controller := NewUserController(userUseCaseStub{
			create: func(user *domain.User) error {
				if user.Nome != "Maria" || user.Email != "maria@example.com" || user.Password != "senha" || user.Role != 2 || !user.Ativo {
					t.Fatalf("usuário recebido = %+v", user)
				}
				return nil
			},
		})
		w, c := userControllerContext(http.MethodPost, `{"nome":"Maria","email":"maria@example.com","password":"senha","role":2}`)
		controller.CreateUser(c)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejeita payload incompleto", func(t *testing.T) {
		controller := NewUserController(userUseCaseStub{
			create: func(*domain.User) error {
				t.Fatal("use case não deveria ser chamado")
				return nil
			},
		})
		w, c := userControllerContext(http.MethodPost, `{"nome":"Maria"}`)
		controller.CreateUser(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, esperado %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("mascara erro do use case", func(t *testing.T) {
		controller := NewUserController(userUseCaseStub{
			create: func(*domain.User) error { return errors.New("falha interna") },
		})
		w, c := userControllerContext(http.MethodPost, `{"nome":"Maria","email":"maria@example.com","password":"senha","role":2}`)
		controller.CreateUser(c)
		if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), domain.ErrCreateUser.Error()) {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
	})
}

func TestUserControllerGetUsers(t *testing.T) {
	t.Run("mapeia usuarios sem expor senha", func(t *testing.T) {
		users := []domain.User{
			{ID: 1, Nome: "Ativo", Email: "ativo@example.com", Password: "segredo", Role: 1, Ativo: true},
			{ID: 2, Nome: "Inativo", Email: "inativo@example.com", Password: "segredo", Role: 2, Ativo: false},
		}
		controller := NewUserController(userUseCaseStub{
			getAll: func() (*[]domain.User, error) { return &users, nil },
		})
		w, c := userControllerContext(http.MethodGet, "")
		controller.GetUsers(c)

		var body struct {
			Users []struct {
				Nome  string `json:"nome"`
				Email string `json:"email"`
				Role  int    `json:"role"`
				Ativo bool   `json:"ativo"`
			} `json:"users"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if w.Code != http.StatusOK || len(body.Users) != 2 || !body.Users[0].Ativo || body.Users[1].Ativo || strings.Contains(w.Body.String(), "segredo") {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
	})

	for _, tc := range []struct {
		name  string
		users *[]domain.User
	}{
		{name: "lista nil"},
		{name: "lista vazia", users: &[]domain.User{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			controller := NewUserController(userUseCaseStub{
				getAll: func() (*[]domain.User, error) { return tc.users, nil },
			})
			w, c := userControllerContext(http.MethodGet, "")
			controller.GetUsers(c)
			if w.Code != http.StatusOK || strings.TrimSpace(w.Body.String()) != `{"users":[]}` {
				t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
			}
		})
	}

	t.Run("retorna erro do use case", func(t *testing.T) {
		expectedErr := errors.New("erro de consulta")
		controller := NewUserController(userUseCaseStub{
			getAll: func() (*[]domain.User, error) { return nil, expectedErr },
		})
		w, c := userControllerContext(http.MethodGet, "")
		controller.GetUsers(c)
		if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), expectedErr.Error()) {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
	})
}

func TestUserControllerUpdateUser(t *testing.T) {
	t.Run("atualiza campos recebidos", func(t *testing.T) {
		controller := NewUserController(userUseCaseStub{
			update: func(user *domain.User) error {
				if user.ID != 7 || user.Nome != "Novo Nome" || user.Email != "novo@example.com" || user.Role != 3 {
					t.Fatalf("usuário recebido = %+v", user)
				}
				return nil
			},
		})
		w, c := userControllerContext(http.MethodPut, `{"id":7,"nome":"Novo Nome","email":"novo@example.com","role":3}`)
		controller.UpdateUser(c)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
	})

	t.Run("rejeita payload invalido", func(t *testing.T) {
		controller := NewUserController(userUseCaseStub{
			update: func(*domain.User) error {
				t.Fatal("use case não deveria ser chamado")
				return nil
			},
		})
		w, c := userControllerContext(http.MethodPut, `{"id":7}`)
		controller.UpdateUser(c)
		if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), domain.ErrUpdateUser.Error()) {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
	})

	t.Run("mascara erro do use case", func(t *testing.T) {
		controller := NewUserController(userUseCaseStub{
			update: func(*domain.User) error { return errors.New("falha interna") },
		})
		w, c := userControllerContext(http.MethodPut, `{"id":7,"nome":"Novo Nome","email":"novo@example.com","role":3}`)
		controller.UpdateUser(c)
		if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), domain.ErrUpdateUser.Error()) {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
		}
	})
}

func TestUserControllerActivateAndDeactivate(t *testing.T) {
	tests := []struct {
		name          string
		responseError error
		successText   string
		invoke        func(*UserController, *gin.Context)
		stub          func(func(int) error) userUseCaseStub
	}{
		{
			name:          "activate",
			responseError: domain.ErrActivateUser,
			successText:   "Usuário ativado com sucesso!",
			invoke:        func(controller *UserController, c *gin.Context) { controller.ActivateUser(c) },
			stub:          func(call func(int) error) userUseCaseStub { return userUseCaseStub{activate: call} },
		},
		{
			name:          "deactivate",
			responseError: domain.ErrDeactivateUser,
			successText:   "Usuário desativado com sucesso!",
			invoke:        func(controller *UserController, c *gin.Context) { controller.DeactivateUser(c) },
			stub:          func(call func(int) error) userUseCaseStub { return userUseCaseStub{deactivate: call} },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name+" success", func(t *testing.T) {
			controller := NewUserController(tc.stub(func(id int) error {
				if id != 42 {
					t.Fatalf("id recebido = %d", id)
				}
				return nil
			}))
			w, c := userControllerContext(http.MethodPut, "")
			c.Params = gin.Params{{Key: "id", Value: "42"}}
			tc.invoke(controller, c)
			if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), tc.successText) {
				t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
			}
		})

		for _, invalidID := range []string{"", "abc"} {
			t.Run(tc.name+" invalid "+invalidID, func(t *testing.T) {
				controller := NewUserController(tc.stub(func(int) error {
					t.Fatal("use case não deveria ser chamado")
					return nil
				}))
				w, c := userControllerContext(http.MethodPut, "")
				c.Params = gin.Params{{Key: "id", Value: invalidID}}
				tc.invoke(controller, c)
				if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), tc.responseError.Error()) {
					t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
				}
			})
		}

		t.Run(tc.name+" use case error", func(t *testing.T) {
			controller := NewUserController(tc.stub(func(int) error { return errors.New("falha interna") }))
			w, c := userControllerContext(http.MethodPut, "")
			c.Params = gin.Params{{Key: "id", Value: "42"}}
			tc.invoke(controller, c)
			if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), tc.responseError.Error()) {
				t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
			}
		})
	}
}

func userControllerContext(method, body string) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/", bytes.NewBufferString(body))
	if body != "" {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	return w, c
}
