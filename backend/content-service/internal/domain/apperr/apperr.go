package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

type Code string

const (
	CodeValidation   Code = "VALIDATION_FAILED"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeNotFound     Code = "NOT_FOUND"
	CodeInternal     Code = "INTERNAL_ERROR"
)

type Field struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Error struct {
	Code    Code
	Message string
	Details []Field
	err     error
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

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok || e == nil || t == nil {
		return false
	}
	return e.Code == t.Code
}

func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

func Wrap(err error, code Code, message string) *Error {
	return &Error{Code: code, Message: message, err: err}
}

func Validation(details ...Field) *Error {
	return &Error{Code: CodeValidation, Message: "invalid request", Details: details}
}

func Required(field string) Field {
	return Field{Field: field, Code: "REQUIRED", Message: field + " is required"}
}

func As(err error) (*Error, bool) {
	var ae *Error
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

func HTTPStatus(code Code) int {
	switch code {
	case CodeValidation:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

var (
	ErrUnauthorized = New(CodeUnauthorized, "unauthorized")
	ErrNotFound     = New(CodeNotFound, "not found")
	ErrInternal     = New(CodeInternal, "internal error")
)
