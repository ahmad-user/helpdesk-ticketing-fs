package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/repository"

	"github.com/google/uuid"
)

func TestUserService_ListUsers(t *testing.T) {
	userRepo := newMockUserRepoForTicket()
	svc := NewUserService(userRepo)

	// Seed 3 users
	_ = userRepo.Create(context.Background(), &model.User{
		ID:        uuid.New(),
		Name:      "Alice Admin",
		Email:     "alice@example.com",
		Role:      model.RoleAdmin,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	_ = userRepo.Create(context.Background(), &model.User{
		ID:        uuid.New(),
		Name:      "Bob Agent",
		Email:     "bob@example.com",
		Role:      model.RoleAgent,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	_ = userRepo.Create(context.Background(), &model.User{
		ID:        uuid.New(),
		Name:      "Charlie Requester",
		Email:     "charlie@example.com",
		Role:      model.RoleRequester,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	users, pagination, err := svc.ListUsers(context.Background(), model.UserFilter{
		Page:  1,
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}
	if pagination.Total != 3 {
		t.Errorf("Expected total 3, got %d", pagination.Total)
	}
}

func TestUserService_UpdateUserRole(t *testing.T) {
	userRepo := newMockUserRepoForTicket()
	svc := NewUserService(userRepo)

	admin1 := &model.User{
		ID:        uuid.New(),
		Name:      "Admin One",
		Email:     "admin1@example.com",
		Role:      model.RoleAdmin,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	user := &model.User{
		ID:        uuid.New(),
		Name:      "User Regular",
		Email:     "regular@example.com",
		Role:      model.RoleRequester,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	_ = userRepo.Create(context.Background(), admin1)
	_ = userRepo.Create(context.Background(), user)

	t.Run("Invalid role returns ErrInvalidRole", func(t *testing.T) {
		_, err := svc.UpdateUserRole(context.Background(), admin1.ID, user.ID, "superadmin")
		if err != ErrInvalidRole {
			t.Errorf("Expected ErrInvalidRole, got %v", err)
		}
	})

	t.Run("Non-existent user returns ErrUserNotFound", func(t *testing.T) {
		_, err := svc.UpdateUserRole(context.Background(), admin1.ID, uuid.New(), model.RoleAgent)
		if err != repository.ErrUserNotFound {
			t.Errorf("Expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("Demoting the only administrator fails with ErrCannotDemoteLastAdmin", func(t *testing.T) {
		_, err := svc.UpdateUserRole(context.Background(), admin1.ID, admin1.ID, model.RoleAgent)
		if err != ErrCannotDemoteLastAdmin {
			t.Errorf("Expected ErrCannotDemoteLastAdmin, got %v", err)
		}
	})

	t.Run("Admin successfully changes regular user role to agent", func(t *testing.T) {
		updated, err := svc.UpdateUserRole(context.Background(), admin1.ID, user.ID, model.RoleAgent)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if updated.Role != model.RoleAgent {
			t.Errorf("Expected role agent, got %s", updated.Role)
		}
	})

	t.Run("Can demote admin when another admin exists", func(t *testing.T) {
		admin2 := &model.User{
			ID:        uuid.New(),
			Name:      "Admin Two",
			Email:     "admin2@example.com",
			Role:      model.RoleAdmin,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		_ = userRepo.Create(context.Background(), admin2)

		// Now there are 2 admins (admin1 and admin2). We can demote admin1.
		updated, err := svc.UpdateUserRole(context.Background(), admin2.ID, admin1.ID, model.RoleAgent)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if updated.Role != model.RoleAgent {
			t.Errorf("Expected role agent, got %s", updated.Role)
		}
	})

	t.Run("Concurrent demotions cannot eliminate the last administrator", func(t *testing.T) {
		concurrencyRepo := newMockUserRepoForTicket()
		concurrencySvc := NewUserService(concurrencyRepo)

		cAdmin1 := &model.User{
			ID:        uuid.New(),
			Name:      "Concurrent Admin 1",
			Email:     "cadmin1@example.com",
			Role:      model.RoleAdmin,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		cAdmin2 := &model.User{
			ID:        uuid.New(),
			Name:      "Concurrent Admin 2",
			Email:     "cadmin2@example.com",
			Role:      model.RoleAdmin,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		_ = concurrencyRepo.Create(context.Background(), cAdmin1)
		_ = concurrencyRepo.Create(context.Background(), cAdmin2)

		numRoutines := 10
		start := make(chan struct{})
		var wg sync.WaitGroup
		successCount := 0
		failCount := 0
		var mu sync.Mutex

		for i := 0; i < numRoutines; i++ {
			wg.Add(1)
			targetID := cAdmin1.ID
			if i%2 == 1 {
				targetID = cAdmin2.ID
			}

			go func(tgt uuid.UUID) {
				defer wg.Done()
				<-start
				_, err := concurrencySvc.UpdateUserRole(context.Background(), tgt, tgt, model.RoleRequester)
				mu.Lock()
				defer mu.Unlock()
				if err == nil {
					successCount++
				} else if errors.Is(err, ErrCannotDemoteLastAdmin) {
					failCount++
				}
			}(targetID)
		}

		close(start)
		wg.Wait()

		// Exactly one of the two admins may have been demoted, but never both.
		remainingAdmins, _ := concurrencyRepo.CountByRole(context.Background(), model.RoleAdmin)
		if remainingAdmins < 1 {
			t.Fatalf("CRITICAL RACE CONDITION: System was left with %d administrators!", remainingAdmins)
		}
		if remainingAdmins != 1 {
			t.Fatalf("Expected exactly 1 remaining administrator, got %d", remainingAdmins)
		}
		if failCount == 0 {
			t.Errorf("Expected at least one ErrCannotDemoteLastAdmin failure, got %d", failCount)
		}
	})
}
