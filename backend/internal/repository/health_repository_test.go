package repository

import (
	"context"
	"testing"
)

func TestHealthRepository_NilDB(t *testing.T) {
	repo := NewHealthRepository(nil)
	err := repo.Ping(context.Background())
	if err == nil {
		t.Errorf("Expected error when pinging nil DB, got nil")
	}
	if err != ErrDatabaseNotConnected {
		t.Errorf("Expected ErrDatabaseNotConnected, got %v", err)
	}
}
