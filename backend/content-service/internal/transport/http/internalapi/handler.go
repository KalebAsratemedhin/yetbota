package internalapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/google/uuid"

	"github.com/yetbota/content-service/internal/domain/apperr"
	svcendpoint "github.com/yetbota/content-service/internal/services/endpoint"
	"github.com/yetbota/content-service/internal/transport/http/shared"
)

type Endpoints struct {
	Nearby    endpoint.Endpoint
	Attach    endpoint.Endpoint
	Summary   endpoint.Endpoint
	Entries   endpoint.Endpoint
	GetByPost endpoint.Endpoint
}

func Mount(mux *http.ServeMux, eps Endpoints, serviceToken string) {
	opts := []kithttp.ServerOption{kithttp.ServerErrorEncoder(shared.EncodeError)}
	wrap := func(ep endpoint.Endpoint, dec kithttp.DecodeRequestFunc) http.Handler {
		return requireServiceToken(serviceToken, kithttp.NewServer(ep, dec, shared.EncodeJSONResponse, opts...))
	}
	mux.Handle("GET /internal/places/nearby", wrap(eps.Nearby, decodeNearby))
	mux.Handle("POST /internal/places/attach", wrap(eps.Attach, decodeAttach))
	mux.Handle("PATCH /internal/places/{id}/summary", wrap(eps.Summary, decodeSummary))
	mux.Handle("POST /internal/places/{id}/entries", wrap(eps.Entries, decodeEntries))
	mux.Handle("GET /internal/places/by-post/{post_id}", wrap(eps.GetByPost, decodeByPost))
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

func decodeNearby(_ context.Context, r *http.Request) (interface{}, error) {
	q := r.URL.Query()
	lat, err1 := parseFloat(q.Get("lat"), "lat")
	lng, err2 := parseFloat(q.Get("lng"), "lng")
	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}
	var radius float64
	if raw := q.Get("radius_m"); raw != "" {
		v, err := parseFloat(raw, "radius_m")
		if err != nil {
			return nil, err
		}
		radius = v
	}
	limit, err := svcendpoint.ParseListLimit(q.Get("limit"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.InternalNearbyRequest{Lat: lat, Lng: lng, RadiusM: radius, Limit: limit}, nil
}

func decodeAttach(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req svcendpoint.InternalAttachRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID", Message: "invalid JSON body"})
	}
	return req, nil
}

func decodeSummary(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "id", Code: "INVALID", Message: "id must be a UUID"})
	}
	defer r.Body.Close()
	var body struct {
		Summary string `json:"summary"`
		Model   string `json:"model"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID", Message: "invalid JSON body"})
	}
	return svcendpoint.InternalSummaryRequest{PlaceID: id, Summary: body.Summary, Model: body.Model}, nil
}

func decodeEntries(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "id", Code: "INVALID", Message: "id must be a UUID"})
	}
	defer r.Body.Close()
	var body svcendpoint.InternalEntriesRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID", Message: "invalid JSON body"})
	}
	body.PlaceID = id
	return body, nil
}

func decodeByPost(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := uuid.Parse(r.PathValue("post_id"))
	if err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "post_id", Code: "INVALID", Message: "post_id must be a UUID"})
	}
	return id, nil
}

func parseFloat(raw, field string) (float64, error) {
	if raw == "" {
		return 0, apperr.Validation(apperr.Required(field))
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, apperr.Validation(apperr.Field{Field: field, Code: "INVALID", Message: field + " must be a number"})
	}
	return v, nil
}
