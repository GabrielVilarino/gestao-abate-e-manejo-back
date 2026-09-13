package route

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/controller"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/gin-gonic/gin"
)

type routeTokenPortStub struct {
	validate func(string) (*domain.User, error)
}

type routeUserPortStub struct {
	getByID func(int) (*domain.User, error)
}

func (s routeUserPortStub) CreateUser(*domain.User) error               { return nil }
func (s routeUserPortStub) GetUsers() (*[]domain.User, error)           { return nil, nil }
func (s routeUserPortStub) GetUserByEmail(string) (*domain.User, error) { return nil, nil }
func (s routeUserPortStub) GetUserByID(id int) (*domain.User, error) {
	if s.getByID == nil {
		return nil, nil
	}
	return s.getByID(id)
}
func (s routeUserPortStub) UpdateUser(*domain.User) error { return nil }
func (s routeUserPortStub) ActivateUser(int) error        { return nil }
func (s routeUserPortStub) DeactivateUser(int) error      { return nil }

func (s routeTokenPortStub) Generate(domain.User) (string, error) { return "", nil }
func (s routeTokenPortStub) Validate(token string) (*domain.User, error) {
	return s.validate(token)
}

type routeUserUseCaseStub struct{}

func (routeUserUseCaseStub) Login(string, string) (*string, error) { return nil, nil }
func (routeUserUseCaseStub) CreateUser(*domain.User) error         { return nil }
func (routeUserUseCaseStub) GetUsers() (*[]domain.User, error) {
	users := []domain.User{}
	return &users, nil
}
func (routeUserUseCaseStub) UpdateUser(*domain.User) error { return nil }
func (routeUserUseCaseStub) ActivateUser(int) error        { return nil }
func (routeUserUseCaseStub) DeactivateUser(int) error      { return nil }

type routeProprietarioUseCaseStub struct{}

func (routeProprietarioUseCaseStub) CreateProprietario(*domain.Proprietario) error { return nil }
func (routeProprietarioUseCaseStub) GetProprietarios() (*[]domain.Proprietario, error) {
	proprietarios := []domain.Proprietario{}
	return &proprietarios, nil
}
func (routeProprietarioUseCaseStub) UpdateProprietario(*domain.Proprietario) error { return nil }
func (routeProprietarioUseCaseStub) ActivateProprietario(int) error                { return nil }
func (routeProprietarioUseCaseStub) DeactivateProprietario(int) error              { return nil }

func TestInitRoutesRegistersV1Routes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	InitRoutes(
		router,
		&controller.UserController{},
		&controller.ProprietarioController{},
		&controller.FazendaController{},
		&controller.AbateController{},
		&controller.AgendaController{},
		routeTokenPortStub{validate: func(string) (*domain.User, error) {
			return &domain.User{Ativo: true}, nil
		}},
		routeUserPortStub{},
	)

	expected := map[string]bool{
		"POST /api/v1/user/login":                 false,
		"POST /api/v1/user/logout":                false,
		"POST /api/v1/user":                       false,
		"GET /api/v1/users":                       false,
		"PUT /api/v1/user":                        false,
		"PUT /api/v1/user/activate/:id":           false,
		"PUT /api/v1/user/deactivate/:id":         false,
		"POST /api/v1/proprietario":               false,
		"GET /api/v1/proprietarios":               false,
		"PUT /api/v1/proprietario":                false,
		"PUT /api/v1/proprietario/activate/:id":   false,
		"PUT /api/v1/proprietario/deactivate/:id": false,
		"POST /api/v1/fazenda":                    false,
		"GET /api/v1/fazendas/:idProprietario":    false,
		"PUT /api/v1/fazenda":                     false,
		"PUT /api/v1/fazenda/activate/:id":        false,
		"PUT /api/v1/fazenda/deactivate/:id":      false,
		"POST /api/v1/abate":                      false,
		"GET /api/v1/abate/:id":                   false,
		"GET /api/v1/abates":                      false,
		"PUT /api/v1/abate/:id/dados-gerais":      false,
		"PUT /api/v1/abate/:id/etapa-fazenda":     false,
		"PUT /api/v1/abate/:id/etapa-frigorifico": false,
		"DELETE /api/v1/abate/:id":                false,
		"POST /api/v1/abate/:id/fotos/:etapa":     false,
		"GET /api/v1/abate/fotos/:fotoID":         false,
		"DELETE /api/v1/abate/fotos/:fotoID":      false,
		"POST /api/v1/agenda":                     false,
		"GET /api/v1/agenda/:id":                  false,
		"GET /api/v1/agendas":                     false,
		"PUT /api/v1/agenda/:id":                  false,
		"DELETE /api/v1/agenda/:id":               false,
		"POST /api/v1/push/subscriptions":         false,
		"DELETE /api/v1/push/subscriptions/:id":   false,
	}
	for _, route := range router.Routes() {
		key := route.Method + " " + route.Path
		if _, ok := expected[key]; ok {
			expected[key] = true
		}
	}
	for route, registered := range expected {
		if !registered {
			t.Errorf("rota não registrada: %s", route)
		}
	}
}

func TestAbateRoutesRequireAuthenticationAndActiveUser(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		user        *domain.User
		validateErr error
		want        int
	}{
		{name: "sem token", want: http.StatusUnauthorized},
		{name: "token inválido", header: "Bearer invalido", validateErr: errors.New("inválido"), want: http.StatusUnauthorized},
		{name: "usuário inativo", header: "Bearer token", user: &domain.User{Ativo: false}, want: http.StatusForbidden},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			InitRoutes(
				router,
				&controller.UserController{},
				&controller.ProprietarioController{},
				&controller.FazendaController{},
				&controller.AbateController{},
				&controller.AgendaController{},
				routeTokenPortStub{validate: func(string) (*domain.User, error) {
					return &domain.User{ID: 7}, tc.validateErr
				}},
				routeUserPortStub{getByID: func(int) (*domain.User, error) { return tc.user, nil }},
			)
			request := httptest.NewRequest(http.MethodGet, "/api/v1/abates", nil)
			request.Header.Set("Authorization", tc.header)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != tc.want {
				t.Fatalf("status=%d esperado=%d body=%s", response.Code, tc.want, response.Body.String())
			}
		})
	}
}

func TestUserRoutesAreAdminOnlyExceptLoginAndLogout(t *testing.T) {
	adminOnly := []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/api/v1/user"},
		{method: http.MethodGet, path: "/api/v1/users"},
		{method: http.MethodPut, path: "/api/v1/user"},
		{method: http.MethodPut, path: "/api/v1/user/activate/1"},
		{method: http.MethodPut, path: "/api/v1/user/deactivate/1"},
	}
	for _, tc := range adminOnly {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			response := performAuthenticatedRequest(newRouteTestRouter(domain.RoleUser, true), tc.method, tc.path)
			if response.Code != http.StatusForbidden {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}

	response := performAuthenticatedRequest(newRouteTestRouter(domain.RoleAdmin, true), http.MethodGet, "/api/v1/users")
	if response.Code != http.StatusOK {
		t.Fatalf("admin em GET /users: status=%d body=%s", response.Code, response.Body.String())
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/user/login", nil)
	response = httptest.NewRecorder()
	newRouteTestRouter(domain.RoleUser, true).ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("login público: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestLogoutRequiresValidAdminOrUserSession(t *testing.T) {
	for _, role := range []int{domain.RoleAdmin, domain.RoleUser} {
		t.Run("perfil válido", func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/user/logout", nil)
			request.AddCookie(&http.Cookie{Name: "auth", Value: "token"})
			response := httptest.NewRecorder()
			newRouteTestRouter(role, true).ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("role=%d status=%d body=%s", role, response.Code, response.Body.String())
			}
		})
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/user/logout", nil)
	response := httptest.NewRecorder()
	newRouteTestRouter(domain.RoleUser, true).ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("sem sessão: status=%d body=%s", response.Code, response.Body.String())
	}

	response = performAuthenticatedRequest(newRouteTestRouter(3, true), http.MethodPost, "/api/v1/user/logout")
	if response.Code != http.StatusForbidden {
		t.Fatalf("perfil inválido: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestAdminAndUserCanAccessBusinessGetPostAndPutRoutes(t *testing.T) {
	businessRoutes := []struct {
		method string
		path   string
		want   int
	}{
		{method: http.MethodPost, path: "/api/v1/proprietario", want: http.StatusBadRequest},
		{method: http.MethodGet, path: "/api/v1/proprietarios", want: http.StatusOK},
		{method: http.MethodPut, path: "/api/v1/proprietario", want: http.StatusBadRequest},
		{method: http.MethodPut, path: "/api/v1/proprietario/activate/invalido", want: http.StatusBadRequest},
		{method: http.MethodPut, path: "/api/v1/proprietario/deactivate/invalido", want: http.StatusBadRequest},
		{method: http.MethodPost, path: "/api/v1/fazenda", want: http.StatusBadRequest},
		{method: http.MethodGet, path: "/api/v1/fazendas/invalido", want: http.StatusBadRequest},
		{method: http.MethodPut, path: "/api/v1/fazenda", want: http.StatusBadRequest},
		{method: http.MethodPut, path: "/api/v1/fazenda/activate/invalido", want: http.StatusBadRequest},
		{method: http.MethodPut, path: "/api/v1/fazenda/deactivate/invalido", want: http.StatusBadRequest},
		{method: http.MethodPost, path: "/api/v1/abate", want: http.StatusBadRequest},
		{method: http.MethodGet, path: "/api/v1/abate/invalido", want: http.StatusBadRequest},
		{method: http.MethodGet, path: "/api/v1/abates?limite=invalido", want: http.StatusBadRequest},
		{method: http.MethodPut, path: "/api/v1/abate/invalido/dados-gerais", want: http.StatusBadRequest},
		{method: http.MethodPut, path: "/api/v1/abate/invalido/etapa-fazenda", want: http.StatusBadRequest},
		{method: http.MethodPut, path: "/api/v1/abate/invalido/etapa-frigorifico", want: http.StatusBadRequest},
		{method: http.MethodPost, path: "/api/v1/abate/invalido/fotos/fazenda", want: http.StatusBadRequest},
		{method: http.MethodGet, path: "/api/v1/abate/fotos/invalido", want: http.StatusBadRequest},
		{method: http.MethodPost, path: "/api/v1/agenda", want: http.StatusBadRequest},
		{method: http.MethodGet, path: "/api/v1/agenda/invalido", want: http.StatusBadRequest},
		{method: http.MethodGet, path: "/api/v1/agendas?pagina=invalida", want: http.StatusBadRequest},
		{method: http.MethodPut, path: "/api/v1/agenda/invalido", want: http.StatusBadRequest},
		{method: http.MethodPost, path: "/api/v1/push/subscriptions", want: http.StatusBadRequest},
	}
	for _, role := range []int{domain.RoleAdmin, domain.RoleUser} {
		for _, tc := range businessRoutes {
			t.Run(tc.method+" "+tc.path, func(t *testing.T) {
				response := performAuthenticatedRequest(newRouteTestRouter(role, true), tc.method, tc.path)
				if response.Code != tc.want {
					t.Fatalf("role=%d status=%d esperado=%d body=%s", role, response.Code, tc.want, response.Body.String())
				}
			})
		}
	}

	response := performAuthenticatedRequest(newRouteTestRouter(3, true), http.MethodGet, "/api/v1/proprietarios")
	if response.Code != http.StatusForbidden {
		t.Fatalf("perfil inválido: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestDeleteRoutesAreAdminOnly(t *testing.T) {
	deleteRoutes := []string{
		"/api/v1/abate/invalido",
		"/api/v1/abate/fotos/invalido",
		"/api/v1/agenda/invalido",
		"/api/v1/push/subscriptions/invalido",
	}
	for _, path := range deleteRoutes {
		t.Run(path, func(t *testing.T) {
			response := performAuthenticatedRequest(newRouteTestRouter(domain.RoleUser, true), http.MethodDelete, path)
			if response.Code != http.StatusForbidden {
				t.Fatalf("usuário: status=%d body=%s", response.Code, response.Body.String())
			}

			response = performAuthenticatedRequest(newRouteTestRouter(domain.RoleAdmin, true), http.MethodDelete, path)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("admin: status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

func newRouteTestRouter(role int, active bool) *gin.Engine {
	router := gin.New()
	InitRoutes(
		router,
		controller.NewUserController(routeUserUseCaseStub{}),
		controller.NewProprietarioController(routeProprietarioUseCaseStub{}),
		&controller.FazendaController{},
		&controller.AbateController{},
		&controller.AgendaController{},
		routeTokenPortStub{validate: func(string) (*domain.User, error) {
			return &domain.User{ID: 7}, nil
		}},
		routeUserPortStub{getByID: func(int) (*domain.User, error) {
			return &domain.User{ID: 7, Ativo: active, Role: role}, nil
		}},
	)
	return router
}

func performAuthenticatedRequest(router *gin.Engine, method, path string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	request.Header.Set("Authorization", "Bearer token")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
