package seeds

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// SeedUser holds the initial seed user details.
type SeedUser struct {
	Name     string
	Email    string
	Password string
	Role     string
}

// DefaultSeedUsers returns the required default users for system bootstrapping.
func DefaultSeedUsers() []SeedUser {
	return []SeedUser{
		{
			Name:     "Default Requester",
			Email:    "requester@example.com",
			Password: "Requester123!",
			Role:     "requester",
		},
		{
			Name:     "Default Support Agent",
			Email:    "agent@example.com",
			Password: "Agent123!",
			Role:     "agent",
		},
		{
			Name:     "System Administrator",
			Email:    "admin@example.com",
			Password: "Admin123!",
			Role:     "admin",
		},
	}
}

// RunSeeds seeds the database with initial users idempotently.
func RunSeeds(ctx context.Context, db *sql.DB) error {
	users := DefaultSeedUsers()

	query := `
		INSERT INTO users (name, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (email) DO NOTHING;
	`

	for _, u := range users {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password for %s: %w", u.Email, err)
		}

		now := time.Now().UTC()
		result, err := db.ExecContext(ctx, query, u.Name, u.Email, string(hashedPassword), u.Role, now, now)
		if err != nil {
			return fmt.Errorf("failed to seed user %s: %w", u.Email, err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			log.Printf("[SEED] Inserted seed user: %s (role: %s)\n", u.Email, u.Role)
		} else {
			log.Printf("[SEED] User already exists (skipped): %s\n", u.Email)
		}
	}

	return nil
}
