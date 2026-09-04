package internalhttp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/yetbota/identity-service/internal/domain/apperr"
	svcendpoint "github.com/yetbota/identity-service/internal/services/endpoint"
	"github.com/yetbota/identity-service/internal/transport/http/shared"
)

func Mount(mux *http.ServeMux, record, syncScore endpoint.Endpoint, serviceToken string) {
	opts := []kithttp.ServerOption{kithttp.ServerErrorEncoder(shared.EncodeError)}
	mux.Handle("POST /internal/reputation/events", requireServiceToken(serviceToken,
		kithttp.NewServer(record, decodeRecord, shared.EncodeJSONResponse, opts...),
	))
	mux.Handle("POST /internal/reputation/score", requireServiceToken(serviceToken,
		kithttp.NewServer(syncScore, decodeSync, shared.EncodeJSONResponse, opts...),
	))
}

func requireServiceToken(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token == "" {
			shared.EncodeError(r.Context(), apperr.New(apperr.CodeInternal, "internal service token not configured"), w)
			return
		}
		got := strings.TrimSpace(r.Header.Get("X-Service-Token"))
		if got == "" || got != token {
			shared.EncodeError(r.Context(), apperr.ErrUnauthorized, w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func decodeRecord(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req svcendpoint.RecordReputationRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID", Message: "invalid JSON body"})
	}
	return req, nil
}

func decodeSync(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req svcendpoint.SyncReputationRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID", Message: "invalid JSON body"})
	}
	return req, nil
}
