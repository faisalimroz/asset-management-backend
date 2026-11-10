package database

import (
	"fmt"
	"log"

	"asset-management-backend/internal/models"

	"gorm.io/gorm"
)

// AutoMigrate runs automatic schema migrations for all models
func AutoMigrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&models.User{},
		&models.Asset{},
		&models.AssetAssignment{},
		&models.AssetRequest{},
		&models.AttendanceLog{},
		&models.LeaveRequest{},
		&models.AuditLog{},
	)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}
