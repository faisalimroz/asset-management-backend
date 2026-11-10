package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application configuration settings
type Config struct {
	ServerPort      string
	Environment     string
	DBDriver        string // "sqlite" or "postgres"
	DBPath          string // for sqlite (e.g., "./asset_management.db")
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	JWTSecret       string
	JWTExpirationHr int
	AutoSeed        bool
}

// LoadConfig loads environment variables from .env file or OS environment
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found or could not be loaded, using environment variables and defaults")
	}

	jwtExp, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	if err != nil {
		jwtExp = 24
	}

	autoSeed, _ := strconv.ParseBool(getEnv("AUTO_SEED", "true"))

	return &Config{
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		DBDriver:        getEnv("DB_DRIVER", "sqlite"),
		DBPath:          getEnv("DB_PATH", "./asset_management.db"),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBUser:          getEnv("DB_USER", "postgres"),
		DBPassword:      getEnv("DB_PASSWORD", "postgres"),
		DBName:          getEnv("DB_NAME", "asset_management"),
		DBSSLMode:       getEnv("DB_SSLMODE", "disable"),
		JWTSecret:       getEnv("JWT_SECRET", "super-secret-company-asset-jwt-key-2026!"),
		JWTExpirationHr: jwtExp,
		AutoSeed:        autoSeed,
	}
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return defaultVal
}
