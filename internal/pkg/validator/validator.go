package validator

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.Message)
}

type Validator struct {
	errors []ValidationError
}

func New() *Validator {
	return &Validator{
		errors: make([]ValidationError, 0),
	}
}

func (v *Validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: "is required",
		})
	}
}

func (v *Validator) MaxLength(field, value string, max int) {
	if len(value) > max {
		v.errors = append(v.errors, ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must not exceed %d characters", max),
		})
	}
}

func (v *Validator) IsValid() bool {
	return len(v.errors) == 0
}

func (v *Validator) Errors() []ValidationError {
	return v.errors
}

func (v *Validator) Error() string {
	if len(v.errors) == 0 {
		return ""
	}
	messages := make([]string, len(v.errors))
	for i, err := range v.errors {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}
