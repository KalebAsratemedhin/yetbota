package places

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/google/uuid"

	"github.com/yetbota/content-service/internal/domain/apperr"
	svcendpoint "github.com/yetbota/content-service/internal/services/endpoint"
	"github.com/yetbota/content-service/internal/transport/http/shared"
)

type Endpoints struct {
	List    endpoint.Endpoint
	Top     endpoint.Endpoint
	Get     endpoint.Endpoint
	Posts   endpoint.Endpoint
	Entries endpoint.Endpoint
}

func Mount(mux *http.ServeMux, eps Endpoints, authn shared.TokenAuthenticator) {
	opts := []kithttp.ServerOption{kithttp.ServerErrorEncoder(shared.EncodeError)}
	public := func(ep endpoint.Endpoint, dec kithttp.DecodeRequestFunc) http.Handler {
		return shared.OptionalAuth(authn, kithttp.NewServer(ep, dec, shared.EncodeJSONResponse, opts...))
	}
	mux.Handle("GET /places", public(eps.List, decodeListPlaces))
	mux.Handle("GET /places/top", public(eps.Top, decodeTopPlaces))
	mux.Handle("GET /places/{id}", public(eps.Get, decodePlaceID))
	mux.Handle("GET /places/{id}/posts", public(eps.Posts, decodePlacePosts))
	mux.Handle("GET /places/{id}/entries", public(eps.Entries, decodePlaceEntries))
}

func decodeTopPlaces(ctx context.Context, r *http.Request) (interface{}, error) {
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.ListTopPlacesRequest{
		Limit: limit, Cursor: r.URL.Query().Get("cursor"),
	}, nil
}

func decodeListPlaces(ctx context.Context, r *http.Request) (interface{}, error) {
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	req := svcendpoint.ListPlacesRequest{
		Query:  r.URL.Query().Get("q"),
		Limit:  limit,
		Cursor: r.URL.Query().Get("cursor"),
	}
	if p, ok := shared.PrincipalFromCtxOptional(ctx); ok {
		req.Principal = p
	}
	latRaw, lngRaw := r.URL.Query().Get("lat"), r.URL.Query().Get("lng")
	if latRaw != "" || lngRaw != "" {
		lat, err1 := strconv.ParseFloat(latRaw, 64)
		lng, err2 := strconv.ParseFloat(lngRaw, 64)
		if err1 != nil || err2 != nil {
			return nil, apperr.Validation(apperr.Field{
				Field: "lat/lng", Code: "INVALID", Message: "lat and lng must be numbers",
			})
		}
		req.Lat, req.Lng = &lat, &lng
	}
	if raw := r.URL.Query().Get("radius_m"); raw != "" {
		radius, err := strconv.ParseFloat(raw, 64)
		if err != nil || radius < 0 {
			return nil, apperr.Validation(apperr.Field{
				Field: "radius_m", Code: "INVALID", Message: "radius_m must be a non-negative number",
			})
		}
		req.RadiusM = radius
	}
	return req, nil
}

func decodePlaceID(ctx context.Context, r *http.Request) (interface{}, error) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	req := svcendpoint.PlaceIDRequest{PlaceID: id}
	if p, ok := shared.PrincipalFromCtxOptional(ctx); ok {
		req.Principal = p
	}
	return req, nil
}

func decodePlacePosts(ctx context.Context, r *http.Request) (interface{}, error) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	req := svcendpoint.PlaceIDRequest{
		PlaceID: id, Limit: limit, Cursor: r.URL.Query().Get("cursor"),
	}
	if p, ok := shared.PrincipalFromCtxOptional(ctx); ok {
		req.Principal = p
	}
	return req, nil
}

func decodePlaceEntries(ctx context.Context, r *http.Request) (interface{}, error) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	req := svcendpoint.PlaceIDRequest{
		PlaceID: id, Limit: limit, Cursor: r.URL.Query().Get("cursor"),
		Category: r.URL.Query().Get("category"),
	}
	if p, ok := shared.PrincipalFromCtxOptional(ctx); ok {
		req.Principal = p
	}
	return req, nil
}

func parseID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperr.Validation(apperr.Field{Field: "id", Code: "INVALID", Message: "id must be a UUID"})
	}
	return id, nil
}
