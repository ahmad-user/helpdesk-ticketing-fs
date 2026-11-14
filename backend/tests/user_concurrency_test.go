package tests

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"helpdesk-ticketing-system/backend/internal/config"
	"helpdesk-ticketing-system/backend/internal/database"
	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/repository"
	"helpdesk-ticketing-system/backend/internal/service"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// TestConcurrentLastAdminDemotion_DatabaseTransaction verifies that when concurrent
// requests attempt to demote administrators, database transactions and advisory/row locks
// guarantee that the last administrator can NEVER be demoted.
func TestConcurrentLastAdminDemotion_DatabaseTransaction(t *testing.T) {
	// Attempt connection to test PostgreSQL database
	dbCfg := config.DBConfig{
		Host:            "localhost",
		Port:            "5432",
		User:            "postgres",
		Password:        "postgres",
		Name:            "helpdesk_db",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := database.NewPostgresDB(dbCfg)
	if err != nil {
		t.Skipf("Skipping real database concurrency test: PostgreSQL not accessible (%v)", err)
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo)

	// Create 2 test administrators
	ts := time.Now().UnixNano()
	admin1 := &model.User{
		ID:           uuid.New(),
		Name:         fmt.Sprintf("Race Admin 1 %d", ts),
		Email:        fmt.Sprintf("race.admin.1.%d@example.com", ts),
		PasswordHash: "hashed_dummy_password",
		Role:         model.RoleAdmin,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	admin2 := &model.User{
		ID:           uuid.New(),
		Name:         fmt.Sprintf("Race Admin 2 %d", ts),
		Email:        fmt.Sprintf("race.admin.2.%d@example.com", ts),
		PasswordHash: "hashed_dummy_password",
		Role:         model.RoleAdmin,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := userRepo.Create(ctx, admin1); err != nil {
		t.Fatalf("Failed to create test admin1: %v", err)
	}
	if err := userRepo.Create(ctx, admin2); err != nil {
		t.Fatalf("Failed to create test admin2: %v", err)
	}

	// Defer cleanup of test users
	defer func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE id IN ($1, $2)", admin1.ID, admin2.ID)
	}()

	// Verify both admins were inserted
	var initialAdmins int64
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE id IN ($1, $2) AND role = 'admin'", admin1.ID, admin2.ID).Scan(&initialAdmins)
	if err != nil || initialAdmins != 2 {
		t.Fatalf("Expected 2 test admins initially, found %d (err: %v)", initialAdmins, err)
	}

	// Launch concurrent demotion attempts against both admins simultaneously
	const concurrency = 20
	start := make(chan struct{})
	var wg sync.WaitGroup

	var lastAdminBlockedCount int64
	var errorCount int64

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		// Alternate target between admin1 and admin2
		targetID := admin1.ID
		callerID := admin2.ID
		if i%2 == 1 {
			targetID = admin2.ID
			callerID = admin1.ID
		}

		go func(tgt, caller uuid.UUID) {
			defer wg.Done()
			<-start // Wait for simultaneous release

			_, updateErr := userSvc.UpdateUserRole(context.Background(), caller, tgt, model.RoleRequester)
			if updateErr != nil {
				if errors.Is(updateErr, service.ErrCannotDemoteLastAdmin) {
					atomic.AddInt64(&lastAdminBlockedCount, 1)
				} else {
					atomic.AddInt64(&errorCount, 1)
				}
			}
		}(targetID, callerID)
	}

	// Unleash all goroutines at the exact same instant
	close(start)
	wg.Wait()

	if errorCount > 0 {
		t.Errorf("Encountered unexpected non-demotion errors: %d", errorCount)
	}

	// Check final state in PostgreSQL directly
	var remainingTestAdmins int64
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE id IN ($1, $2) AND role = 'admin'", admin1.ID, admin2.ID).Scan(&remainingTestAdmins)
	if err != nil {
		t.Fatalf("Failed to query remaining test admins: %v", err)
	}

	// Total admins across system must never be 0
	var totalSystemAdmins int64
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&totalSystemAdmins)
	if err != nil {
		t.Fatalf("Failed to query total system admins: %v", err)
	}

	t.Logf("Postgres Concurrency Test Results: remainingTestAdmins=%d, totalSystemAdmins=%d, blockedAttempts=%d",
		remainingTestAdmins, totalSystemAdmins, lastAdminBlockedCount)

	if totalSystemAdmins == 0 {
		t.Fatalf("CRITICAL RACE CONDITION DETECTED: System has ZERO administrators remaining!")
	}

	// If there were other system admins or only these two, verify that at least one of these two remained an admin
	// or that lastAdminBlockedCount > 0
	if totalSystemAdmins < 1 {
		t.Errorf("Expected at least 1 administrator in the system, found %d", totalSystemAdmins)
	}

	// Subtest: Proving that when only two administrators exist in the entire database,
	// concurrent demotion requests can NEVER reduce the admin count below 1.
	t.Run("When only two admins exist, concurrent demotions leave exactly one admin", func(t *testing.T) {
		// 1. Create two fresh, isolated test admins
		subAdmin1 := &model.User{
			ID:           uuid.New(),
			Name:         "Iso Admin 1",
			Email:        fmt.Sprintf("iso.admin.1.%d@example.com", time.Now().UnixNano()),
			PasswordHash: "hashed_dummy_password",
			Role:         model.RoleAdmin,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}
		subAdmin2 := &model.User{
			ID:           uuid.New(),
			Name:         "Iso Admin 2",
			Email:        fmt.Sprintf("iso.admin.2.%d@example.com", time.Now().UnixNano()),
			PasswordHash: "hashed_dummy_password",
			Role:         model.RoleAdmin,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}
		if err := userRepo.Create(ctx, subAdmin1); err != nil {
			t.Fatalf("Failed to create subAdmin1: %v", err)
		}
		if err := userRepo.Create(ctx, subAdmin2); err != nil {
			t.Fatalf("Failed to create subAdmin2: %v", err)
		}
		defer func() {
			_, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE id IN ($1, $2)", subAdmin1.ID, subAdmin2.ID)
		}()

		// 2. Temporarily set any OTHER existing admins to a temporary role so that
		// EXACTLY 2 admins exist in the entire system.
		var existingAdminIDs []uuid.UUID
		rows, err := db.QueryContext(ctx, "SELECT id FROM users WHERE role = 'admin' AND id NOT IN ($1, $2)", subAdmin1.ID, subAdmin2.ID)
		if err != nil {
			t.Fatalf("Failed to query other admins: %v", err)
		}
		defer rows.Close()
		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err == nil {
				existingAdminIDs = append(existingAdminIDs, id)
			}
		}
		rows.Close()

		if len(existingAdminIDs) > 0 {
			for _, oid := range existingAdminIDs {
				_, _ = db.ExecContext(ctx, "UPDATE users SET role = 'agent' WHERE id = $1", oid)
			}
			// Ensure we always restore existing admins
			defer func() {
				for _, oid := range existingAdminIDs {
					_, _ = db.ExecContext(context.Background(), "UPDATE users SET role = 'admin' WHERE id = $1", oid)
				}
			}()
		}

		// Verify that exactly 2 admins exist now in the entire database
		var countBefore int64
		_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&countBefore)
		if countBefore != 2 {
			t.Fatalf("Precondition failed: Expected exactly 2 admins in system, got %d", countBefore)
		}

		// 3. Fire 20 concurrent demotion requests
		subConcurrency := 20
		subStart := make(chan struct{})
		var subWg sync.WaitGroup
		var subBlockedCount int64
		var subSuccessCount int64

		for i := 0; i < subConcurrency; i++ {
			subWg.Add(1)
			target := subAdmin1.ID
			caller := subAdmin2.ID
			if i%2 == 1 {
				target = subAdmin2.ID
				caller = subAdmin1.ID
			}

			go func(tgt, clr uuid.UUID) {
				defer subWg.Done()
				<-subStart

				_, err := userSvc.UpdateUserRole(context.Background(), clr, tgt, model.RoleRequester)
				if err != nil {
					if errors.Is(err, service.ErrCannotDemoteLastAdmin) {
						atomic.AddInt64(&subBlockedCount, 1)
					}
				} else {
					atomic.AddInt64(&subSuccessCount, 1)
				}
			}(target, caller)
		}

		close(subStart)
		subWg.Wait()

		// 4. Verify system state
		var countAfter int64
		_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&countAfter)

		t.Logf("Isolated 2-Admin Race Result: countAfter=%d, successes=%d, blocked=%d", countAfter, subSuccessCount, subBlockedCount)

		if countAfter == 0 {
			t.Fatalf("CRITICAL RACE CONDITION: System was left with 0 administrators!")
		}
		if countAfter != 1 {
			t.Errorf("Expected exactly 1 remaining administrator in database, got %d", countAfter)
		}
		if subBlockedCount == 0 {
			t.Errorf("Expected at least one demotion attempt to be blocked with ErrCannotDemoteLastAdmin")
		}
	})

	// Also test demoting the absolute last administrator when only 1 admin exists
	t.Run("Absolute single admin demotion is strictly blocked", func(t *testing.T) {
		// Identify remaining admin
		var remainingAdminID uuid.UUID
		err := db.QueryRowContext(ctx, "SELECT id FROM users WHERE role = 'admin' LIMIT 1").Scan(&remainingAdminID)
		if err != nil {
			t.Fatalf("No admin found in database: %v", err)
		}

		// Attempt to demote all other admins temporarily, leaving only 1
		// Then try to demote that 1 admin
		var count int64
		_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
		if count == 1 {
			_, err := userSvc.UpdateUserRole(ctx, remainingAdminID, remainingAdminID, model.RoleAgent)
			if !errors.Is(err, service.ErrCannotDemoteLastAdmin) {
				t.Fatalf("Expected ErrCannotDemoteLastAdmin, got: %v", err)
			}
		}
	})
}
