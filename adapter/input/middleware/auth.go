package middleware

import (
	"net/http"
	"strings"

	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/domain"
	"github.com/GabrielVilarino/gestao-abate-e-manejo-back/application/port/output"
	"github.com/gin-gonic/gin"
)

func Authenticate(tokenPort output.TokenPort, userPort output.UserPort) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" {
			if cookie, err := c.Cookie("auth"); err == nil {
				token = cookie
			}
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Autenticação obrigatória"})
			return
		}
		claimsUser, err := tokenPort.Validate(token)
		if err != nil || claimsUser == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
			return
		}
		user, err := userPort.GetUserByID(claimsUser.ID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Não foi possível validar a sessão"})
			return
		}
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Sessão revogada"})
			return
		}
		if !user.Ativo {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Usuário inativo"})
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

func AuthorizeRoles(roles ...int) gin.HandlerFunc {
	allowed := make(map[int]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(c *gin.Context) {
		value, exists := c.Get("user")
		user, ok := value.(*domain.User)
		if !exists || !ok || user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Autenticação obrigatória"})
			return
		}
		if _, ok := allowed[user.Role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Usuário sem permissão"})
			return
		}
		c.Next()
	}
}

func bearerToken(header string) string {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
