package service

import (
	"errors"
	"time"

	"go-rest-boilerplate/internal/dto"
	"go-rest-boilerplate/internal/models"
	"go-rest-boilerplate/internal/repository"
	"go-rest-boilerplate/pkg/hash"
	"go-rest-boilerplate/pkg/jwt"
	"go-rest-boilerplate/pkg/token"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidRefresh     = errors.New("invalid or expired refresh token")
	ErrUserInactive       = errors.New("user account is inactive")
)

type AuthService interface {
	Register(req dto.RegisterRequest) (*models.User, error)
	Login(req dto.LoginRequest) (*dto.AuthResponse, error)
	Refresh(rawRefreshToken string) (*dto.AuthResponse, error)
	Logout(rawRefreshToken string) error
}

type authService struct {
	userRepo    repository.UserRepository
	refreshRepo repository.RefreshTokenRepository
	rbacRepo    repository.RBACRepository
	jwtManager  *jwt.Manager
	refreshTTL  time.Duration
	defaultRole string
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshRepo repository.RefreshTokenRepository,
	rbacRepo repository.RBACRepository,
	jwtManager *jwt.Manager,
	refreshTTL time.Duration,
) AuthService {
	return &authService{
		userRepo:    userRepo,
		refreshRepo: refreshRepo,
		rbacRepo:    rbacRepo,
		jwtManager:  jwtManager,
		refreshTTL:  refreshTTL,
		defaultRole: "user",
	}
}

func (s *authService) Register(req dto.RegisterRequest) (*models.User, error) {
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	hashed, err := hash.Password(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashed,
		IsActive:     true,
	}

	if role, err := s.rbacRepo.FindRoleByName(s.defaultRole); err == nil {
		user.Roles = []models.Role{*role}
	}
	// If the default role isn't seeded yet, the user is simply created
	// without roles rather than failing registration.

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *authService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !hash.ComparePassword(user.PasswordHash, req.Password) {
		return nil, ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	return s.issueTokenPair(user)
}

func (s *authService) Refresh(rawRefreshToken string) (*dto.AuthResponse, error) {
	tokenHash := token.Hash(rawRefreshToken)

	stored, err := s.refreshRepo.FindValidByHash(tokenHash)
	if err != nil {
		return nil, ErrInvalidRefresh
	}

	user, err := s.userRepo.FindByID(stored.UserID)
	if err != nil {
		return nil, ErrInvalidRefresh
	}
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Rotate: revoke the used refresh token before issuing a new pair so a
	// stolen-and-replayed token can only be used once.
	if err := s.refreshRepo.Revoke(stored.ID); err != nil {
		return nil, err
	}

	return s.issueTokenPair(user)
}

func (s *authService) Logout(rawRefreshToken string) error {
	tokenHash := token.Hash(rawRefreshToken)
	stored, err := s.refreshRepo.FindValidByHash(tokenHash)
	if err != nil {
		// Already invalid/expired/revoked — logout is a no-op, not an error.
		return nil
	}
	return s.refreshRepo.Revoke(stored.ID)
}

func (s *authService) issueTokenPair(user *models.User) (*dto.AuthResponse, error) {
	roleNames := make([]string, 0, len(user.Roles))
	for _, r := range user.Roles {
		roleNames = append(roleNames, r.Name)
	}

	accessToken, expiresAt, err := s.jwtManager.GenerateAccessToken(user.ID, roleNames)
	if err != nil {
		return nil, err
	}

	rawRefresh, refreshHash, err := token.Generate()
	if err != nil {
		return nil, err
	}

	if err := s.refreshRepo.Create(&models.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: time.Now().Add(s.refreshTTL),
	}); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresAt:    expiresAt.Format(time.RFC3339),
		TokenType:    "Bearer",
	}, nil
}
