package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"helpdesk-ticketing-system/backend/internal/config"
	"helpdesk-ticketing-system/backend/internal/model"
	"helpdesk-ticketing-system/backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Authentication domain errors.
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

// JWTClaims defines the standard and custom claims encoded in the auth token.
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AuthService defines business operations for user authentication and authorization.
type AuthService interface {
	Register(ctx context.Context, req *model.RegisterRequest) (*model.UserResponse, error)
	Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*model.UserResponse, error)
	GenerateToken(user *model.User) (string, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
}

type authService struct {
	userRepo  repository.UserRepository
	jwtConfig config.JWTConfig
}

// NewAuthService creates a new AuthService instance.
func NewAuthService(userRepo repository.UserRepository, jwtConfig config.JWTConfig) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
	}
}

// Register creates a new user account with role forced to requester.
func (s *authService) Register(ctx context.Context, req *model.RegisterRequest) (*model.UserResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, repository.ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now().UTC()
	newUser := &model.User{
		ID:           uuid.New(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         model.RoleRequester, // Enforce requester role for public registration
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	res := newUser.ToResponse()
	return &res, nil
}

// Login verifies credentials and issues a signed JWT.
func (s *authService) Login(ctx context.Context, req *model.LoginRequest) (*model.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("database error during login: %w", err)
	}

	// Verify password hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate JWT
	tokenString, err := s.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth token: %w", err)
	}

	return &model.AuthResponse{
		Token: tokenString,
		User:  user.ToResponse(),
	}, nil
}

// GetProfile retrieves safe user profile details by ID.
func (s *authService) GetProfile(ctx context.Context, userID uuid.UUID) (*model.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := user.ToResponse()
	return &res, nil
}

// GenerateToken creates a signed JWT containing user ID, email, role, and expiry.
func (s *authService) GenerateToken(user *model.User) (string, error) {
	expirationTime := time.Now().UTC().Add(time.Duration(s.jwtConfig.ExpirationHours) * time.Hour)

	claims := &JWTClaims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtConfig.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken parses and validates a JWT token string.
func (s *authService) ValidateToken(tokenString string) (*JWTClaims, error) {
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtConfig.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Validate required claims are present and well-formed
	if claims.UserID == "" {
		return nil, ErrInvalidToken
	}
	if _, err := uuid.Parse(claims.UserID); err != nil {
		return nil, ErrInvalidToken
	}
	if claims.Email == "" {
		return nil, ErrInvalidToken
	}
	if !model.IsValidRole(claims.Role) {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
