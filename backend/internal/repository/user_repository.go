package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"helpdesk-ticketing-system/backend/internal/model"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Common repository errors.
var (
	ErrUserNotFound          = errors.New("user not found")
	ErrEmailAlreadyExists    = errors.New("email address is already registered")
	ErrCannotDemoteLastAdmin = errors.New("cannot demote the last remaining administrator")
	ErrCannotDeleteSelf      = errors.New("cannot delete your own account")
	ErrUserHasHistory        = errors.New("user has tickets or comments and cannot be deleted")
)

// UserRepository defines database operations for the User entity.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindAll(ctx context.Context, filter model.UserFilter) ([]*model.User, int64, error)
	UpdateRole(ctx context.Context, id uuid.UUID, role string) error
	CountByRole(ctx context.Context, role string) (int64, error)
	UpdateUserRoleTx(ctx context.Context, id uuid.UUID, newRole string) (*model.User, error)
	UpdateUserTx(ctx context.Context, user *model.User) (*model.User, error)
	DeleteUserTx(ctx context.Context, actorID, id uuid.UUID) error
}

type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// Create inserts a new user record into PostgreSQL.
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if r.db == nil {
		return ErrDatabaseNotConnected
	}

	query := `
		INSERT INTO users (id, name, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at;
	`

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.ID,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// 23505 is PostgreSQL unique_violation code
			if pqErr.Code == "23505" {
				return ErrEmailAlreadyExists
			}
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

// FindByEmail retrieves a user by email address.
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	if r.db == nil {
		return nil, ErrDatabaseNotConnected
	}

	query := `
		SELECT id, name, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE email = $1;
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return user, nil
}

// FindByID retrieves a user by their UUID primary key.
func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if r.db == nil {
		return nil, ErrDatabaseNotConnected
	}

	query := `
		SELECT id, name, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1;
	`

	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	return user, nil
}

// FindAll retrieves users with pagination and search filtering.
func (r *userRepository) FindAll(ctx context.Context, filter model.UserFilter) ([]*model.User, int64, error) {
	if r.db == nil {
		return nil, 0, ErrDatabaseNotConnected
	}

	baseQuery := "FROM users"
	var whereClauses []string
	var args []interface{}
	argIndex := 1

	if filter.Search != "" {
		searchPattern := "%" + strings.ToLower(filter.Search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(name) LIKE $%d OR LOWER(email) LIKE $%d)", argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	if len(whereClauses) > 0 {
		baseQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	dataQuery := fmt.Sprintf(`
		SELECT id, name, email, password_hash, role, created_at, updated_at
		%s
		ORDER BY created_at ASC
		LIMIT $%d OFFSET $%d;
	`, baseQuery, argIndex, argIndex+1)

	dataArgs := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("users row error: %w", err)
	}

	return users, total, nil
}

// UpdateUserRoleTx safely updates a user's role within an ACID database transaction.
// When an administrator is being demoted, it acquires a PostgreSQL transaction-level
// advisory lock and a row-level lock (FOR UPDATE), ensuring concurrent demotion requests
// cannot reduce the administrator count to zero.
func (r *userRepository) UpdateUserRoleTx(ctx context.Context, id uuid.UUID, newRole string) (*model.User, error) {
	if r.db == nil {
		return nil, ErrDatabaseNotConnected
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Acquire transaction-level advisory lock to serialize admin role modifications across instances
	const adminDemotionLockQuery = `SELECT pg_advisory_xact_lock(hashtext('helpdesk_admin_role_protection'));`
	if _, err := tx.ExecContext(ctx, adminDemotionLockQuery); err != nil {
		return nil, fmt.Errorf("failed to acquire advisory lock: %w", err)
	}

	// 2. Row-level lock on the target user (FOR UPDATE)
	targetUserQuery := `
		SELECT id, name, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1
		FOR UPDATE;
	`
	targetUser := &model.User{}
	err = tx.QueryRowContext(ctx, targetUserQuery, id).Scan(
		&targetUser.ID,
		&targetUser.Name,
		&targetUser.Email,
		&targetUser.PasswordHash,
		&targetUser.Role,
		&targetUser.CreatedAt,
		&targetUser.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to lock target user: %w", err)
	}

	// 3. Safety check: Prevent demoting the last remaining administrator
	if targetUser.Role == model.RoleAdmin && newRole != model.RoleAdmin {
		var adminCount int64
		countQuery := `SELECT COUNT(*) FROM users WHERE role = $1;`
		if err := tx.QueryRowContext(ctx, countQuery, model.RoleAdmin).Scan(&adminCount); err != nil {
			return nil, fmt.Errorf("failed to count administrators: %w", err)
		}

		if adminCount <= 1 {
			return nil, ErrCannotDemoteLastAdmin
		}
	}

	// 4. Update the target user's role and return updated model
	updateQuery := `
		UPDATE users
		SET role = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING id, name, email, password_hash, role, created_at, updated_at;
	`
	updatedUser := &model.User{}
	err = tx.QueryRowContext(ctx, updateQuery, newRole, id).Scan(
		&updatedUser.ID,
		&updatedUser.Name,
		&updatedUser.Email,
		&updatedUser.PasswordHash,
		&updatedUser.Role,
		&updatedUser.CreatedAt,
		&updatedUser.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to update user role: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return updatedUser, nil
}

// UpdateRole updates the role and updated_at timestamp of a user.
func (r *userRepository) UpdateRole(ctx context.Context, id uuid.UUID, role string) error {
	_, err := r.UpdateUserRoleTx(ctx, id, role)
	return err
}

// CountByRole returns the total number of users who have a specific role.
func (r *userRepository) CountByRole(ctx context.Context, role string) (int64, error) {
	if r.db == nil {
		return 0, ErrDatabaseNotConnected
	}

	query := `SELECT COUNT(*) FROM users WHERE role = $1;`
	var count int64
	err := r.db.QueryRowContext(ctx, query, role).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users by role: %w", err)
	}

	return count, nil
}

func (r *userRepository) UpdateUserTx(ctx context.Context, user *model.User) (*model.User, error) {
	if r.db == nil {
		return nil, ErrDatabaseNotConnected
	}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('helpdesk_admin_role_protection'))`); err != nil {
		return nil, err
	}
	current := &model.User{}
	err = tx.QueryRowContext(ctx, `SELECT id,name,email,password_hash,role,created_at,updated_at FROM users WHERE id=$1 FOR UPDATE`, user.ID).Scan(&current.ID, &current.Name, &current.Email, &current.PasswordHash, &current.Role, &current.CreatedAt, &current.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if current.Role == model.RoleAdmin && user.Role != model.RoleAdmin {
		var count int64
		if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role=$1`, model.RoleAdmin).Scan(&count); err != nil {
			return nil, err
		}
		if count <= 1 {
			return nil, ErrCannotDemoteLastAdmin
		}
	}
	updated := &model.User{}
	err = tx.QueryRowContext(ctx, `UPDATE users SET name=$1,email=$2,role=$3,updated_at=CURRENT_TIMESTAMP WHERE id=$4 RETURNING id,name,email,password_hash,role,created_at,updated_at`, user.Name, user.Email, user.Role, user.ID).Scan(&updated.ID, &updated.Name, &updated.Email, &updated.PasswordHash, &updated.Role, &updated.CreatedAt, &updated.UpdatedAt)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrEmailAlreadyExists
		}
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return updated, nil
}

func (r *userRepository) DeleteUserTx(ctx context.Context, actorID, id uuid.UUID) error {
	if r.db == nil {
		return ErrDatabaseNotConnected
	}
	if actorID == id {
		return ErrCannotDeleteSelf
	}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('helpdesk_admin_role_protection'))`); err != nil {
		return err
	}
	var role string
	err = tx.QueryRowContext(ctx, `SELECT role FROM users WHERE id=$1 FOR UPDATE`, id).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}
	if role == model.RoleAdmin {
		var count int64
		if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role=$1`, model.RoleAdmin).Scan(&count); err != nil {
			return err
		}
		if count <= 1 {
			return ErrCannotDemoteLastAdmin
		}
	}
	var refs int64
	if err = tx.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM tickets WHERE requester_id=$1) + (SELECT COUNT(*) FROM comments WHERE author_id=$1)`, id).Scan(&refs); err != nil {
		return err
	}
	if refs > 0 {
		return ErrUserHasHistory
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return tx.Commit()
}
