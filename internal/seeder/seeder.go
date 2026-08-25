package seeder

import (
	"log"

	"go-rest-boilerplate/internal/models"

	"gorm.io/gorm"
)

// Seed creates the baseline roles and permissions if they don't already
// exist. Safe to run on every startup. Add new permissions here as the
// application grows — RBAC middleware will pick them up automatically.
func Seed(db *gorm.DB) error {
	permissions := []models.Permission{
		{Name: "user:read", Description: "View user accounts"},
		{Name: "user:create", Description: "Create user accounts"},
		{Name: "user:update", Description: "Update user accounts"},
		{Name: "user:delete", Description: "Delete user accounts"},
	}
	for i := range permissions {
		if err := db.Where("name = ?", permissions[i].Name).
			FirstOrCreate(&permissions[i]).Error; err != nil {
			return err
		}
	}

	adminRole := models.Role{Name: "admin", Description: "Full system access"}
	if err := db.Where("name = ?", adminRole.Name).FirstOrCreate(&adminRole).Error; err != nil {
		return err
	}
	if err := db.Model(&adminRole).Association("Permissions").Replace(permissions); err != nil {
		return err
	}

	userRole := models.Role{Name: "user", Description: "Standard authenticated user"}
	if err := db.Where("name = ?", userRole.Name).FirstOrCreate(&userRole).Error; err != nil {
		return err
	}

	log.Println("roles and permissions seeded")
	return nil
}
