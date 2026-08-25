package middleware

import (
	"net/http"

	"go-rest-boilerplate/internal/repository"
	"go-rest-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
)

// RequireRole does a fast, DB-free check against the roles embedded in the
// access token. Use it for coarse checks (e.g. "must be admin").
func RequireRole(allowed ...string) gin.HandlerFunc {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, r := range allowed {
		allowedSet[r] = struct{}{}
	}

	return func(c *gin.Context) {
		roles, _ := c.Get(ContextRoles)
		roleList, _ := roles.([]string)

		for _, r := range roleList {
			if _, ok := allowedSet[r]; ok {
				c.Next()
				return
			}
		}
		response.Error(c, http.StatusForbidden, "insufficient role")
		c.Abort()
	}
}

// RequirePermission checks the database directly, so a revoked permission
// takes effect immediately instead of waiting for the access token to
// expire. Use it for fine-grained, security-sensitive actions.
func RequirePermission(rbacRepo repository.RBACRepository, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, ok := c.Get(ContextUserID)
		if !ok {
			response.Error(c, http.StatusUnauthorized, "unauthenticated")
			c.Abort()
			return
		}
		userID := userIDVal.(uint)

		allowed, err := rbacRepo.UserHasPermission(userID, permission)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "authorization check failed")
			c.Abort()
			return
		}
		if !allowed {
			response.Error(c, http.StatusForbidden, "missing permission: "+permission)
			c.Abort()
			return
		}
		c.Next()
	}
}
