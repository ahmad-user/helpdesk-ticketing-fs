package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"helpdesk-ticketing-system/backend/internal/config"

	_ "github.com/lib/pq"
)

// ErrInvalidConfig indicates missing mandatory database configuration fields.
var ErrInvalidConfig = errors.New("invalid database configuration: host, port, user, and dbname are required")

// NewPostgresDB initializes and verifies a PostgreSQL connection pool.
func NewPostgresDB(cfg config.DBConfig) (*sql.DB, error) {
	if cfg.Host == "" || cfg.Port == "" || cfg.User == "" || cfg.Name == "" {
		return nil, ErrInvalidConfig
	}

	dsn := cfg.DSN()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database driver: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verify connection with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf(
			"failed to connect to PostgreSQL at %s:%s (database: '%s', user: '%s'): %w",
			cfg.Host, cfg.Port, cfg.Name, cfg.User, err,
		)
	}

	log.Printf("[INFO] Successfully connected to PostgreSQL database '%s' at %s:%s\n", cfg.Name, cfg.Host, cfg.Port)
	return db, nil
}

// ClosePostgresDB closes the database connection gracefully.
func ClosePostgresDB(db *sql.DB) {
	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("[ERROR] Error closing database connection: %v\n", err)
		} else {
			log.Println("[INFO] Database connection closed successfully")
		}
	}
}
