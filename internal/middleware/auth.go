package middleware

import (
	"net/http"
	"strings"

	"go-rest-boilerplate/pkg/jwt"
	"go-rest-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID = "user_id"
	ContextRoles  = "roles"
)

// JWTAuth validates the Bearer access token and stores user_id/roles in the
// gin context for downstream handlers and authorization middleware.
func JWTAuth(manager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Error(c, http.StatusUnauthorized, "missing or malformed authorization header")
			c.Abort()
			return
		}

		rawToken := strings.TrimPrefix(header, "Bearer ")
		claims, err := manager.ValidateAccessToken(rawToken)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid or expired access token")
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextRoles, claims.Roles)
		c.Next()
	}
}
