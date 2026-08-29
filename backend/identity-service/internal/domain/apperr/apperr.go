package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

// Code is a stable, client-facing error identifier.
type Code string

const (
	CodeValidation         Code = "VALIDATION_FAILED"
	CodeUsernameTaken      Code = "USERNAME_TAKEN"
	CodeInvalidCredentials Code = "INVALID_CREDENTIALS"
	CodeUnauthorized       Code = "UNAUTHORIZED"
	CodeNotFound           Code = "NOT_FOUND"
	CodeInternal           Code = "INTERNAL_ERROR"
)

// Field describes a single validation problem.
type Field struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error is the application error type used across layers.
type Error struct {
	Code    Code
	Message string
	Details []Field
	err     error // optional wrapped cause (not exposed to clients)
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.err)
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.err }

// Is allows errors.Is(err, ErrUnauthorized) by matching Code.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok || e == nil || t == nil {
		return false
	}
	return e.Code == t.Code
}

// New builds a sentinel-style app error.
func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Wrap attaches a cause while preserving code/message for HTTP mapping.
func Wrap(err error, code Code, message string) *Error {
	return &Error{Code: code, Message: message, err: err}
}

// Validation returns VALIDATION_FAILED with field details.
func Validation(details ...Field) *Error {
	return &Error{
		Code:    CodeValidation,
		Message: "invalid request",
		Details: details,
	}
}

func Required(field string) Field {
	return Field{Field: field, Code: "REQUIRED", Message: field + " is required"}
}

// As extracts *Error from err (including wrapped).
func As(err error) (*Error, bool) {
	var ae *Error
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

// HTTPStatus maps code → status. Unknown → 500.
func HTTPStatus(code Code) int {
	switch code {
	case CodeValidation:
		return http.StatusBadRequest
	case CodeUsernameTaken:
		return http.StatusConflict
	case CodeInvalidCredentials, CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

var (
	ErrUsernameTaken      = New(CodeUsernameTaken, "username already taken")
	ErrInvalidCredentials = New(CodeInvalidCredentials, "invalid credentials")
	ErrUnauthorized       = New(CodeUnauthorized, "unauthorized")
	ErrNotFound           = New(CodeNotFound, "not found")
	ErrInternal           = New(CodeInternal, "internal error")
)

// IsCode reports whether err is an *Error with the given code.
func IsCode(err error, code Code) bool {
	ae, ok := As(err)
	return ok && ae.Code == code
}
