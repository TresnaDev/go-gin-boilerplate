package router

import (
	"go-rest-boilerplate/internal/config"
	"go-rest-boilerplate/internal/handler"
	"go-rest-boilerplate/internal/middleware"
	"go-rest-boilerplate/internal/repository"
	"go-rest-boilerplate/internal/service"
	"go-rest-boilerplate/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(db *gorm.DB, cfg *config.Config) *gin.Engine {
	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTAccessTTL)

	userRepo := repository.NewUserRepository(db)
	refreshRepo := repository.NewRefreshTokenRepository(db)
	rbacRepo := repository.NewRBACRepository(db)

	authService := service.NewAuthService(userRepo, refreshRepo, rbacRepo, jwtManager, cfg.JWTRefreshTTL)
	authHandler := handler.NewAuthHandler(authService, userRepo)

	r := gin.Default()

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
		}

		protected := api.Group("/")
		protected.Use(middleware.JWTAuth(jwtManager))
		{
			protected.GET("/me", authHandler.Me)

			// Example of role-based (fast, JWT-only) protection:
			protected.GET("/admin/ping", middleware.RequireRole("admin"), func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "pong"})
			})

			// Example of fine-grained, DB-backed permission protection:
			protected.DELETE("/admin/users/:id",
				middleware.RequirePermission(rbacRepo, "user:delete"),
				func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "user " + c.Param("id") + " deleted"})
				},
			)
		}
	}

	return r
}
