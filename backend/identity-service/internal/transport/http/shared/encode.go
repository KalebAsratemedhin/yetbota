package shared

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	domain "github.com/yetbota/identity-service/internal/domain/user"
)

type errorBody struct {
	Error string `json:"error"`
}

func EncodeJSONResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

func EncodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	status := http.StatusInternalServerError
	msg := "internal error"

	switch {
	case errors.Is(err, domain.ErrValidation):
		status = http.StatusBadRequest
		msg = "invalid request"
	case errors.Is(err, domain.ErrUsernameTaken):
		status = http.StatusConflict
		msg = "username already taken"
	case errors.Is(err, domain.ErrInvalidCredentials):
		status = http.StatusUnauthorized
		msg = "invalid credentials"
	case errors.Is(err, domain.ErrUnauthorized):
		status = http.StatusUnauthorized
		msg = "unauthorized"
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		msg = "not found"
	}

	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorBody{Error: msg})
}
