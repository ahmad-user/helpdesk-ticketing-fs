package validator

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
)

type sampleStruct struct {
	Email string `validate:"required,email"`
}

func TestFormatValidationErrors_GenericError(t *testing.T) {
	err := errors.New("custom non-validation error")
	formatted := FormatValidationErrors(err)

	if len(formatted) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(formatted))
	}
	if formatted[0].Field != "general" {
		t.Errorf("Expected field 'general', got '%s'", formatted[0].Field)
	}
}

func TestFormatValidationErrors_ValidatorErrors(t *testing.T) {
	validate := validator.New()
	sample := sampleStruct{Email: "invalid-email"}

	err := validate.Struct(sample)
	if err == nil {
		t.Fatalf("Expected validation error, got nil")
	}

	formatted := FormatValidationErrors(err)
	if len(formatted) != 1 {
		t.Fatalf("Expected 1 formatted error, got %d", len(formatted))
	}
	if formatted[0].Field != "email" {
		t.Errorf("Expected field 'email', got '%s'", formatted[0].Field)
	}
}
