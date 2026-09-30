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
	Get     endpoint.Endpoint
	Posts   endpoint.Endpoint
	Entries endpoint.Endpoint
}

func Mount(mux *http.ServeMux, eps Endpoints, authn shared.TokenAuthenticator) {
	opts := []kithttp.ServerOption{kithttp.ServerErrorEncoder(shared.EncodeError)}
	auth := func(ep endpoint.Endpoint, dec kithttp.DecodeRequestFunc) http.Handler {
		return shared.RequireAuth(authn, kithttp.NewServer(ep, dec, shared.EncodeJSONResponse, opts...))
	}
	mux.Handle("GET /places", auth(eps.List, decodeListPlaces))
	mux.Handle("GET /places/{id}", auth(eps.Get, decodePlaceID))
	mux.Handle("GET /places/{id}/posts", auth(eps.Posts, decodePlacePosts))
	mux.Handle("GET /places/{id}/entries", auth(eps.Entries, decodePlaceEntries))
}

func decodeListPlaces(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	req := svcendpoint.ListPlacesRequest{
		Principal: p,
		Query:     r.URL.Query().Get("q"),
		Limit:     limit,
		Cursor:    r.URL.Query().Get("cursor"),
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
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.PlaceIDRequest{Principal: p, PlaceID: id}, nil
}

func decodePlacePosts(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.PlaceIDRequest{
		Principal: p, PlaceID: id, Limit: limit, Cursor: r.URL.Query().Get("cursor"),
	}, nil
}

func decodePlaceEntries(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.PlaceIDRequest{
		Principal: p, PlaceID: id, Limit: limit, Cursor: r.URL.Query().Get("cursor"),
		Category: r.URL.Query().Get("category"),
	}, nil
}

func parseID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperr.Validation(apperr.Field{Field: "id", Code: "INVALID", Message: "id must be a UUID"})
	}
	return id, nil
}
