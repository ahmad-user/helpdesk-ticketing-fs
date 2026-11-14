package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError represents an individual field validation error message.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FormatValidationErrors formats go-playground validator errors into clean structured errors.
func FormatValidationErrors(err error) []ValidationError {
	var errs []ValidationError

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		errs = append(errs, ValidationError{
			Field:   "general",
			Message: err.Error(),
		})
		return errs
	}

	for _, fe := range validationErrors {
		field := strings.ToLower(fe.Field())
		var message string

		switch fe.Tag() {
		case "required":
			message = fmt.Sprintf("Field '%s' is required", fe.Field())
		case "email":
			message = fmt.Sprintf("Field '%s' must be a valid email address", fe.Field())
		case "min":
			message = fmt.Sprintf("Field '%s' must be at least %s characters", fe.Field(), fe.Param())
		case "max":
			message = fmt.Sprintf("Field '%s' must not exceed %s characters", fe.Field(), fe.Param())
		default:
			message = fmt.Sprintf("Field '%s' failed validation for tag '%s'", fe.Field(), fe.Tag())
		}

		errs = append(errs, ValidationError{
			Field:   field,
			Message: message,
		})
	}

	return errs
}
