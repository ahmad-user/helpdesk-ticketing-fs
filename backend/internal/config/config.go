package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	// DefaultDevJWTSecret is the fallback secret used in development and test environments.
	DefaultDevJWTSecret = "super_secret_jwt_key_helpdesk_change_in_production"
	// DefaultDockerComposeFallback is the historical docker compose fallback secret.
	DefaultDockerComposeFallback = "enterprise_production_jwt_secret_key_helpdesk_2026_xyz"
	// MinProductionJWTSecretLength defines the minimum required character length for production JWT secrets.
	MinProductionJWTSecretLength = 32
)

// Config holds the application, database, and JWT configuration.
type Config struct {
	App AppConfig
	DB  DBConfig
	JWT JWTConfig
}

// AppConfig contains settings for the HTTP application server.
type AppConfig struct {
	Name string
	Env  string
	Port string
	Host string
}

// DBConfig contains settings for PostgreSQL database connection and connection pool.
type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// JWTConfig contains settings for JSON Web Token creation and verification.
type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

// DSN returns the PostgreSQL connection string.
func (db DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode,
	)
}

// LoadConfig reads configuration from .env file (if present) and environment variables.
func LoadConfig() (*Config, error) {
	// Attempt to load .env file if it exists, ignore error if it doesn't exist
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] No .env file found or unable to load, reading from environment variables")
	}

	cfg := &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "helpdesk-ticketing-system"),
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8080"),
			Host: getEnv("APP_HOST", "0.0.0.0"),
		},
		DB: DBConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			Name:            getEnv("DB_NAME", "helpdesk_db"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 15*time.Minute),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", DefaultDevJWTSecret),
			ExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
		},
	}

	// Strictly validate production JWT secrets. Startup fails immediately if invalid.
	if strings.EqualFold(cfg.App.Env, "production") {
		rawSecret := os.Getenv("JWT_SECRET")
		if err := ValidateJWTSecret(cfg.App.Env, rawSecret); err != nil {
			return nil, fmt.Errorf("production configuration error: %w", err)
		}
		cfg.JWT.Secret = rawSecret
	} else {
		// Non-production fallback
		if cfg.JWT.Secret == "" {
			cfg.JWT.Secret = DefaultDevJWTSecret
		}
	}

	return cfg, nil
}

// ValidateJWTSecret validates the JWT secret according to the target environment.
// In production, the secret must not be empty, must not use development/compose fallbacks, and must meet minimum length.
// Under no circumstances is the actual secret value logged or exposed in returned error messages.
func ValidateJWTSecret(env, secret string) error {
	if !strings.EqualFold(env, "production") {
		return nil
	}

	trimmed := strings.TrimSpace(secret)
	if trimmed == "" {
		return errors.New("JWT_SECRET environment variable is required and cannot be empty in production")
	}
	if trimmed == DefaultDevJWTSecret || trimmed == DefaultDockerComposeFallback {
		return errors.New("JWT_SECRET cannot use default development or compose fallback secret in production")
	}
	if len(trimmed) < MinProductionJWTSecretLength {
		return fmt.Errorf("JWT_SECRET is too weak for production: minimum length is %d characters", MinProductionJWTSecretLength)
	}
	return nil
}

// Helper functions for reading environment variables with defaults.
func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		log.Printf("[WARN] Invalid integer value for %s: %s. Using default: %d\n", key, valStr, defaultVal)
		return defaultVal
	}
	return val
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := time.ParseDuration(valStr)
	if err != nil {
		log.Printf("[WARN] Invalid duration value for %s: %s. Using default: %v\n", key, valStr, defaultVal)
		return defaultVal
	}
	return val
}
