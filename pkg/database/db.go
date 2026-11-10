package database

import (
	"fmt"
	"log"
	"time"

	"asset-management-backend/internal/config"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes the database connection based on config
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	logLevel := logger.Warn
	if cfg.Environment == "development" {
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	switch cfg.DBDriver {
	case "postgres":
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
			cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode,
		)
		dialector = postgres.Open(dsn)
		log.Printf("Connecting to PostgreSQL at %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)

	case "sqlite":
		fallthrough
	default:
		// Using pure-Go SQLite driver (no CGO required)
		dbPath := cfg.DBPath
		if dbPath == "" {
			dbPath = "./asset_management.db"
		}
		dialector = sqlite.Open(dbPath)
		log.Printf("Connecting to SQLite database at %s", dbPath)
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm: %w", err)
	}

	// Configure connection pool
	if cfg.DBDriver == "sqlite" {
		sqlDB.SetMaxOpenConns(1) // SQLite single writer to avoid busy locks
	} else {
		sqlDB.SetMaxOpenConns(25)
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(10 * time.Minute)

	return db, nil
}
