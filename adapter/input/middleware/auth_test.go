package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/gin-gonic/gin"
)

type tokenPortStub struct {
	validate func(string) (*domain.User, error)
}

type userPortStub struct {
	getByID func(int) (*domain.User, error)
}

func (s userPortStub) CreateUser(*domain.User) error               { return nil }
func (s userPortStub) GetUsers() (*[]domain.User, error)           { return nil, nil }
func (s userPortStub) GetUserByEmail(string) (*domain.User, error) { return nil, nil }
func (s userPortStub) GetUserByID(id int) (*domain.User, error)    { return s.getByID(id) }
func (s userPortStub) UpdateUser(*domain.User) error               { return nil }
func (s userPortStub) ActivateUser(int) error                      { return nil }
func (s userPortStub) DeactivateUser(int) error                    { return nil }

func (s tokenPortStub) Generate(domain.User) (string, error) { return "", nil }
func (s tokenPortStub) Validate(token string) (*domain.User, error) {
	return s.validate(token)
}

func TestAuthenticateRejectsMissingAndInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name   string
		header string
	}{
		{name: "ausente"},
		{name: "inválido", header: "Bearer inválido"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.Use(Authenticate(tokenPortStub{validate: func(string) (*domain.User, error) {
				return nil, errors.New("inválido")
			}}, userPortStub{getByID: func(int) (*domain.User, error) { return nil, nil }}))
			router.GET("/abates", func(c *gin.Context) { c.Status(http.StatusOK) })
			request := httptest.NewRequest(http.MethodGet, "/abates", nil)
			request.Header.Set("Authorization", tc.header)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

func TestAuthenticateRejectsInactiveUser(t *testing.T) {
	router := gin.New()
	router.Use(Authenticate(
		tokenPortStub{validate: func(string) (*domain.User, error) { return &domain.User{ID: 7, Ativo: true}, nil }},
		userPortStub{getByID: func(int) (*domain.User, error) { return &domain.User{ID: 7, Ativo: false}, nil }},
	))
	router.GET("/abates", func(c *gin.Context) { c.Status(http.StatusOK) })
	request := httptest.NewRequest(http.MethodGet, "/abates", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestAuthenticateUsesCurrentDatabaseStatus(t *testing.T) {
	router := gin.New()
	router.Use(Authenticate(
		tokenPortStub{validate: func(string) (*domain.User, error) {
			return &domain.User{ID: 7, Ativo: true, Role: domain.RoleAdmin}, nil
		}},
		userPortStub{getByID: func(id int) (*domain.User, error) {
			if id != 7 {
				t.Fatalf("id=%d", id)
			}
			return &domain.User{ID: 7, Ativo: true, Role: domain.RoleUser}, nil
		}},
	))
	router.Use(AuthorizeRoles(domain.RoleAdmin))
	router.GET("/admin", func(c *gin.Context) { c.Status(http.StatusOK) })
	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
