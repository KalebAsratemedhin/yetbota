package internalapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	kithttp "github.com/go-kit/kit/transport/http"

	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	"github.com/beka-birhanu/yetbota/content-service/internal/services/endpoint"
	placeSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/place"
	"github.com/beka-birhanu/yetbota/content-service/internal/transport/http/shared"
)

type Config struct {
	E              *endpoint.Endpoints
	InternalAPIKey string
	PathPrefix     string
}

func NewHandler(cfg *Config) (http.Handler, error) {
	mux := http.NewServeMux()
	if cfg == nil || cfg.E == nil {
		return mux, nil
	}
	opts := shared.ServerOptions()
	mux.Handle("PATCH /places/{id}/summary", apiKeyMiddleware(cfg.InternalAPIKey, kithttp.NewServer(
		cfg.E.PlaceUpdateSummary,
		decodePlaceUpdateSummaryHTTP,
		encodePlaceUpdateSummaryHTTP,
		opts...,
	)))
	return mux, nil
}

func apiKeyMiddleware(expected string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if expected == "" {
			http.Error(w, "internal api key not configured", http.StatusUnauthorized)
			return
		}
		got := strings.TrimSpace(r.Header.Get("X-Internal-API-Key"))
		if got == "" || got != expected {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func decodePlaceUpdateSummaryHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		SummaryText string `json:"summary_text"`
		Citations   []struct {
			ContributionID string `json:"contribution_id"`
			Kind           string `json:"kind"`
			Excerpt        string `json:"excerpt"`
		} `json:"summary_citations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "invalid json",
			ServiceMessage:    err.Error(),
		}
	}
	citations := make([]placeSvc.SummaryCitation, 0, len(in.Citations))
	for _, c := range in.Citations {
		citations = append(citations, placeSvc.SummaryCitation{
			ContributionID: c.ContributionID,
			Kind:           c.Kind,
			Excerpt:        c.Excerpt,
		})
	}
	req := &placeSvc.UpdateSummaryRequest{
		PlaceID:     r.PathValue("id"),
		SummaryText: in.SummaryText,
		Citations:   citations,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if data := ctx.Value(ctxYB.AppSession); data != nil {
		if ctxSess, ok := data.(*ctxYB.Context); ok && ctxSess != nil {
			ctxSess.SetRequest(req)
		}
	}
	return req, nil
}

func encodePlaceUpdateSummaryHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	env := shared.Envelope{Success: true, Data: map[string]any{}}
	if data := ctx.Value(ctxYB.AppSession); data != nil {
		if ctxSess, ok := data.(*ctxYB.Context); ok && ctxSess != nil {
			ctxSess.Response = env
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}
