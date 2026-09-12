package route

import (
	"testing"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/adapter/input/controller"
	"github.com/gin-gonic/gin"
)

func TestInitRoutesRegistersV1Routes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	InitRoutes(
		router,
		&controller.UserController{},
		&controller.ProprietarioController{},
		&controller.FazendaController{},
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
