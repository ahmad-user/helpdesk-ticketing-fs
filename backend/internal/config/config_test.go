package config

import (
	"os"
	"strings"
	"testing"
)

func TestValidateJWTSecret(t *testing.T) {
	tests := []struct {
		name          string
		env           string
		secret        string
		expectErr     bool
		expectedErrSub string
	}{
		{
			name:      "Development env allows default dev secret",
			env:       "development",
			secret:    DefaultDevJWTSecret,
			expectErr: false,
		},
		{
			name:      "Development env allows custom short secret",
			env:       "development",
			secret:    "short",
			expectErr: false,
		},
		{
			name:      "Test env allows empty or arbitrary secret",
			env:       "test",
			secret:    "",
			expectErr: false,
		},
		{
			name:           "Production fails if secret is empty",
			env:            "production",
			secret:         "",
			expectErr:      true,
			expectedErrSub: "JWT_SECRET environment variable is required",
		},
		{
			name:           "Production fails if secret is whitespace only",
			env:            "production",
			secret:         "   \t\n  ",
			expectErr:      true,
			expectedErrSub: "JWT_SECRET environment variable is required",
		},
		{
			name:           "Production fails if secret is default dev secret",
			env:            "production",
			secret:         DefaultDevJWTSecret,
			expectErr:      true,
			expectedErrSub: "cannot use default development or compose fallback secret",
		},
		{
			name:           "Production fails if secret is docker compose fallback",
			env:            "production",
			secret:         DefaultDockerComposeFallback,
			expectErr:      true,
			expectedErrSub: "cannot use default development or compose fallback secret",
		},
		{
			name:           "Production fails if secret is too short (< 32 chars)",
			env:            "production",
			secret:         "too_short_secret_under_32_chars",
			expectErr:      true,
			expectedErrSub: "too weak for production",
		},
		{
			name:      "Production succeeds with strong secret (>= 32 chars)",
			env:       "production",
			secret:    "production_secure_jwt_secret_with_sufficient_entropy_64chars_long",
			expectErr: false,
		},
		{
			name:           "Production case insensitive env match (PRODUCTION)",
			env:            "PRODUCTION",
			secret:         "short",
			expectErr:      true,
			expectedErrSub: "too weak for production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJWTSecret(tt.env, tt.secret)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				if tt.expectedErrSub != "" && !strings.Contains(err.Error(), tt.expectedErrSub) {
					t.Errorf("Expected error containing %q, got %q", tt.expectedErrSub, err.Error())
				}
				// Verify secret value is NEVER leaked in error message
				if tt.secret != "" && strings.Contains(err.Error(), tt.secret) {
					t.Errorf("CRITICAL SECURITY FLAW: Secret was leaked in error message: %q", err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestLoadConfig_ProductionValidation(t *testing.T) {
	// Backup original env vars
	origEnv := os.Getenv("APP_ENV")
	origSecret := os.Getenv("JWT_SECRET")
	defer func() {
		os.Setenv("APP_ENV", origEnv)
		os.Setenv("JWT_SECRET", origSecret)
	}()

	t.Run("Production fails when JWT_SECRET is unset", func(t *testing.T) {
		os.Setenv("APP_ENV", "production")
		os.Unsetenv("JWT_SECRET")

		cfg, err := LoadConfig()
		if err == nil {
			t.Fatalf("Expected LoadConfig to fail in production with unset JWT_SECRET, but succeeded: %+v", cfg)
		}
	})

	t.Run("Production fails when JWT_SECRET uses fallback", func(t *testing.T) {
		os.Setenv("APP_ENV", "production")
		os.Setenv("JWT_SECRET", DefaultDevJWTSecret)

		_, err := LoadConfig()
		if err == nil {
			t.Fatalf("Expected LoadConfig to fail when using DefaultDevJWTSecret")
		}
	})

	t.Run("Production succeeds when strong JWT_SECRET is provided", func(t *testing.T) {
		os.Setenv("APP_ENV", "production")
		os.Setenv("JWT_SECRET", "super_strong_production_secret_key_1234567890_abcdefghij")

		cfg, err := LoadConfig()
		if err != nil {
			t.Fatalf("Expected LoadConfig to succeed with strong secret, got: %v", err)
		}
		if cfg.JWT.Secret != "super_strong_production_secret_key_1234567890_abcdefghij" {
			t.Errorf("Expected JWT secret to match configured value")
		}
	})
}
