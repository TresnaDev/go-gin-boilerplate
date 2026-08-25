package database

import (
	"fmt"
	"log"
	"time"

	"go-rest-boilerplate/internal/config"
	"go-rest-boilerplate/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	maxConnectRetries = 10
	retryDelay        = 3 * time.Second
)

// Connect opens the database connection with a retry loop, since in Docker
// Compose the api container can start before Postgres finishes accepting
// connections even when depends_on/healthcheck is set.
func Connect(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	var db *gorm.DB
	var err error

	for attempt := 1; attempt <= maxConnectRetries; attempt++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
		if err == nil {
			if pingErr := ping(db); pingErr == nil {
				log.Println("database connected")
				return db
			} else {
				err = pingErr
			}
		}

		log.Printf("database not ready (attempt %d/%d): %v", attempt, maxConnectRetries, err)
		time.Sleep(retryDelay)
	}

	log.Fatalf("failed to connect to database after %d attempts: %v", maxConnectRetries, err)
	return nil
}

func ping(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// AutoMigrate keeps the schema in sync with the models. Order matters:
// join tables are created after their referenced tables.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.RefreshToken{},
	)
}
