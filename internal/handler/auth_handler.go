package handler

import (
	"errors"
	"net/http"

	"go-rest-boilerplate/internal/dto"
	"go-rest-boilerplate/internal/middleware"
	"go-rest-boilerplate/internal/repository"
	"go-rest-boilerplate/internal/service"
	"go-rest-boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
	userRepo    repository.UserRepository
}

func NewAuthHandler(authService service.AuthService, userRepo repository.UserRepository) *AuthHandler {
	return &AuthHandler{authService: authService, userRepo: userRepo}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.authService.Register(req)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to register user")
		return
	}

	response.OK(c, http.StatusCreated, "registration successful", dto.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	tokens, err := h.authService.Login(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			response.Error(c, http.StatusUnauthorized, err.Error())
		case errors.Is(err, service.ErrUserInactive):
			response.Error(c, http.StatusForbidden, err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "failed to login")
		}
		return
	}

	response.OK(c, http.StatusOK, "login successful", tokens)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	tokens, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	response.OK(c, http.StatusOK, "token refreshed", tokens)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to logout")
		return
	}

	response.OK(c, http.StatusOK, "logout successful", nil)
}

// Me returns the authenticated user's profile. Requires middleware.JWTAuth.
func (h *AuthHandler) Me(c *gin.Context) {
	userIDVal, _ := c.Get(middleware.ContextUserID)
	userID := userIDVal.(uint)

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}

	roleNames := make([]string, 0, len(user.Roles))
	for _, r := range user.Roles {
		roleNames = append(roleNames, r.Name)
	}

	response.OK(c, http.StatusOK, "ok", dto.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Roles: roleNames,
	})
}
