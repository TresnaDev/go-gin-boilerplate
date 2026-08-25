package repository

import (
	"go-rest-boilerplate/internal/models"

	"gorm.io/gorm"
)

type RBACRepository interface {
	UserHasPermission(userID uint, permission string) (bool, error)
	FindRoleByName(name string) (*models.Role, error)
}

type rbacRepository struct {
	db *gorm.DB
}

func NewRBACRepository(db *gorm.DB) RBACRepository {
	return &rbacRepository{db: db}
}

// UserHasPermission checks, straight from the database, whether any role
// assigned to the user grants the given permission. Doing this on every
// call (instead of trusting the JWT) means a revoked permission takes
// effect immediately instead of waiting for the access token to expire.
func (r *rbacRepository) UserHasPermission(userID uint, permission string) (bool, error) {
	var count int64
	err := r.db.Table("user_roles").
		Joins("JOIN role_permissions ON role_permissions.role_id = user_roles.role_id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("user_roles.user_id = ? AND permissions.name = ?", userID, permission).
		Count(&count).Error
	return count > 0, err
}

func (r *rbacRepository) FindRoleByName(name string) (*models.Role, error) {
	var role models.Role
	if err := r.db.Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}
