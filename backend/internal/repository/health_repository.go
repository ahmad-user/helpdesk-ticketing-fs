package repository

import (
	"context"
	"database/sql"
	"errors"
)

// ErrDatabaseNotConnected is returned when pinging a nil database instance.
var ErrDatabaseNotConnected = errors.New("database connection is not established")

// HealthRepository defines the contract for health data source operations.
type HealthRepository interface {
	Ping(ctx context.Context) error
}

type healthRepository struct {
	db *sql.DB
}

// NewHealthRepository creates a new instance of HealthRepository.
func NewHealthRepository(db *sql.DB) HealthRepository {
	return &healthRepository{db: db}
}

// Ping checks whether the database is responsive.
func (r *healthRepository) Ping(ctx context.Context) error {
	if r.db == nil {
		return ErrDatabaseNotConnected
	}
	return r.db.PingContext(ctx)
}
