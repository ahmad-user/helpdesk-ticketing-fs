package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User service domain errors.
var (
	ErrInvalidRole           = errors.New("invalid role, allowed roles: requester, agent, admin")
	ErrCannotDemoteLastAdmin = repository.ErrCannotDemoteLastAdmin
)

// UserService defines business operations for administrative user management.
type UserService interface {
	ListUsers(ctx context.Context, filter model.UserFilter) ([]model.UserResponse, *model.PaginationMetadata, error)
	UpdateUserRole(ctx context.Context, adminID uuid.UUID, targetID uuid.UUID, newRole string) (*model.UserResponse, error)
	CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.UserResponse, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req model.UpdateUserRequest) (*model.UserResponse, error)
	DeleteUser(ctx context.Context, actorID, id uuid.UUID) error
}

func (s *userService) CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.UserResponse, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Role = strings.ToLower(strings.TrimSpace(req.Role))
	if !model.IsValidRole(req.Role) {
		return nil, ErrInvalidRole
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	user := &model.User{ID: uuid.New(), Name: req.Name, Email: req.Email, PasswordHash: string(hash), Role: req.Role, CreatedAt: now, UpdatedAt: now}
	if err = s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	resp := user.ToResponse()
	return &resp, nil
}

func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, req model.UpdateUserRequest) (*model.UserResponse, error) {
	current, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	current.Name = strings.TrimSpace(req.Name)
	current.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Role != nil {
		current.Role = strings.ToLower(strings.TrimSpace(*req.Role))
	}
	if !model.IsValidRole(current.Role) {
		return nil, ErrInvalidRole
	}
	updated, err := s.userRepo.UpdateUserTx(ctx, current)
	if err != nil {
		return nil, err
	}
	resp := updated.ToResponse()
	return &resp, nil
}

func (s *userService) DeleteUser(ctx context.Context, actorID, id uuid.UUID) error {
	return s.userRepo.DeleteUserTx(ctx, actorID, id)
}

type userService struct {
	userRepo repository.UserRepository
}

// NewUserService constructs a new UserService instance.
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// ListUsers retrieves users matching filters with pagination metadata.
// Password hashes are guaranteed not to be exposed in the response.
func (s *userService) ListUsers(ctx context.Context, filter model.UserFilter) ([]model.UserResponse, *model.PaginationMetadata, error) {
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
	filter.Page = page
	filter.Limit = limit

	users, total, err := s.userRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	resps := make([]model.UserResponse, 0, len(users))
	for _, u := range users {
		resps = append(resps, u.ToResponse())
	}

	pagination := &model.PaginationMetadata{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	return resps, pagination, nil
}

// UpdateUserRole changes a user's system role safely within a database transaction.
// Prevents accidentally or concurrently eliminating the last administrator from the system.
func (s *userService) UpdateUserRole(ctx context.Context, adminID uuid.UUID, targetID uuid.UUID, newRole string) (*model.UserResponse, error) {
	if !model.IsValidRole(newRole) {
		return nil, ErrInvalidRole
	}

	updatedUser, err := s.userRepo.UpdateUserRoleTx(ctx, targetID, newRole)
	if err != nil {
		return nil, err
	}

	resp := updatedUser.ToResponse()
	return &resp, nil
}
