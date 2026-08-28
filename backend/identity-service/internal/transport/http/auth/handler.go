package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	domain "github.com/yetbota/identity-service/internal/domain/user"
	svcendpoint "github.com/yetbota/identity-service/internal/services/endpoint"
	"github.com/yetbota/identity-service/internal/transport/http/shared"
)

func NewHandler(signUp, signIn endpoint.Endpoint) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(shared.EncodeError),
	}

	mux := http.NewServeMux()
	mux.Handle("POST /signup", kithttp.NewServer(
		signUp,
		decodeCredentials,
		encodeCreated,
		opts...,
	))
	mux.Handle("POST /signin", kithttp.NewServer(
		signIn,
		decodeCredentials,
		shared.EncodeJSONResponse,
		opts...,
	))
	return mux
}

func decodeCredentials(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req svcendpoint.CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, domain.ErrValidation
	}
	return req, nil
}

func encodeCreated(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(http.StatusCreated)
	return shared.EncodeJSONResponse(ctx, w, response)
}
