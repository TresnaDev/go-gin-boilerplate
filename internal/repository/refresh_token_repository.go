package repository

import (
	"errors"
	"time"

	"go-rest-boilerplate/internal/models"

	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(token *models.RefreshToken) error
	FindValidByHash(hash string) (*models.RefreshToken, error)
	Revoke(id uint) error
	RevokeAllForUser(userID uint) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

// FindValidByHash returns the token only if it exists, is not revoked, and
// has not expired.
func (r *refreshTokenRepository) FindValidByHash(hash string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := r.db.Where("token_hash = ? AND revoked = ? AND expires_at > ?", hash, false, time.Now()).
		First(&rt).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *refreshTokenRepository) Revoke(id uint) error {
	return r.db.Model(&models.RefreshToken{}).Where("id = ?", id).Update("revoked", true).Error
}

func (r *refreshTokenRepository) RevokeAllForUser(userID uint) error {
	return r.db.Model(&models.RefreshToken{}).Where("user_id = ? AND revoked = ?", userID, false).
		Update("revoked", true).Error
}
