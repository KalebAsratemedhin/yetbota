package users

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	svcendpoint "github.com/yetbota/identity-service/internal/services/endpoint"
	"github.com/yetbota/identity-service/internal/transport/http/shared"
)

type Endpoints struct {
	GetProfile  endpoint.Endpoint
	Leaderboard endpoint.Endpoint
	Search      endpoint.Endpoint
	Follow      endpoint.Endpoint
	Unfollow    endpoint.Endpoint
	Followers   endpoint.Endpoint
	Following   endpoint.Endpoint
}

func NewHandler(eps Endpoints, authn shared.TokenAuthenticator) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(shared.EncodeError),
	}

	mux := http.NewServeMux()
	mux.Handle("GET /search", shared.RequireAuth(authn, kithttp.NewServer(eps.Search, decodeSearch, shared.EncodeJSONResponse, opts...)))
	mux.Handle("GET /leaderboard", shared.OptionalAuth(authn, kithttp.NewServer(eps.Leaderboard, decodeLeaderboard, shared.EncodeJSONResponse, opts...)))
	mux.Handle("GET /{id}", shared.OptionalAuth(authn, kithttp.NewServer(eps.GetProfile, decodeGetProfile, shared.EncodeJSONResponse, opts...)))
	mux.Handle("POST /{id}/follow", shared.RequireAuth(authn, kithttp.NewServer(eps.Follow, decodeUserID, shared.EncodeJSONResponse, opts...)))
	mux.Handle("DELETE /{id}/follow", shared.RequireAuth(authn, kithttp.NewServer(eps.Unfollow, decodeUserID, shared.EncodeJSONResponse, opts...)))
	mux.Handle("GET /{id}/followers", shared.RequireAuth(authn, kithttp.NewServer(eps.Followers, decodeUserList, shared.EncodeJSONResponse, opts...)))
	mux.Handle("GET /{id}/following", shared.RequireAuth(authn, kithttp.NewServer(eps.Following, decodeUserList, shared.EncodeJSONResponse, opts...)))
	return mux
}

func decodeLeaderboard(ctx context.Context, r *http.Request) (interface{}, error) {
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.LeaderboardRequest{Limit: limit}, nil
}

func decodeGetProfile(ctx context.Context, r *http.Request) (interface{}, error) {
	id, err := svcendpoint.ParseUserID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	req := svcendpoint.UserIDRequest{UserID: id}
	if claims, ok := shared.ClaimsFromCtxOptional(ctx); ok {
		req.Claims = claims
	}
	return req, nil
}

func decodeUserID(ctx context.Context, r *http.Request) (interface{}, error) {
	claims, err := shared.ClaimsFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := svcendpoint.ParseUserID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.UserIDRequest{Claims: claims, UserID: id}, nil
}

func decodeUserList(ctx context.Context, r *http.Request) (interface{}, error) {
	claims, err := shared.ClaimsFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := svcendpoint.ParseUserID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.UserListRequest{
		Claims: claims,
		UserID: id,
		Limit:  limit,
		Cursor: r.URL.Query().Get("cursor"),
	}, nil
}

func decodeSearch(ctx context.Context, r *http.Request) (interface{}, error) {
	claims, err := shared.ClaimsFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	return svcendpoint.SearchRequest{
		Claims: claims,
		Query:  r.URL.Query().Get("q"),
	}, nil
}
