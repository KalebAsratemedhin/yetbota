package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	domain "github.com/yetbota/identity-service/internal/domain/user"
	svcendpoint "github.com/yetbota/identity-service/internal/services/endpoint"
	"github.com/yetbota/identity-service/internal/transport/http/shared"
)

type TokenAuthenticator interface {
	AuthenticateToken(ctx context.Context, token string) (domain.TokenClaims, error)
}

type Endpoints struct {
	SignUp         endpoint.Endpoint
	SignIn         endpoint.Endpoint
	Me             endpoint.Endpoint
	UpdateProfile  endpoint.Endpoint
	Introspect     endpoint.Endpoint
	Logout         endpoint.Endpoint
	ChangePassword endpoint.Endpoint
	Refresh        endpoint.Endpoint
	DeleteAccount  endpoint.Endpoint
}

func NewHandler(eps Endpoints, authn TokenAuthenticator) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(shared.EncodeError),
	}

	mux := http.NewServeMux()
	mux.Handle("POST /signup", kithttp.NewServer(eps.SignUp, decodeCredentials, encodeCreated, opts...))
	mux.Handle("POST /signin", kithttp.NewServer(eps.SignIn, decodeCredentials, shared.EncodeJSONResponse, opts...))
	mux.Handle("POST /introspect", kithttp.NewServer(eps.Introspect, decodeIntrospect, shared.EncodeJSONResponse, opts...))
	mux.Handle("POST /refresh", kithttp.NewServer(eps.Refresh, decodeRefresh, shared.EncodeJSONResponse, opts...))

	mux.Handle("GET /me", requireAuth(authn, kithttp.NewServer(eps.Me, decodeClaimsOnly, shared.EncodeJSONResponse, opts...)))
	mux.Handle("PATCH /me", requireAuth(authn, kithttp.NewServer(eps.UpdateProfile, decodeUpdateProfile, shared.EncodeJSONResponse, opts...)))
	mux.Handle("POST /logout", requireAuth(authn, kithttp.NewServer(eps.Logout, decodeLogout, shared.EncodeJSONResponse, opts...)))
	mux.Handle("POST /password", requireAuth(authn, kithttp.NewServer(eps.ChangePassword, decodeChangePassword, shared.EncodeJSONResponse, opts...)))
	mux.Handle("DELETE /me", requireAuth(authn, kithttp.NewServer(eps.DeleteAccount, decodeClaimsOnly, shared.EncodeJSONResponse, opts...)))

	return mux
}

type ctxKey int

const claimsKey ctxKey = 1

func requireAuth(authn TokenAuthenticator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := bearerToken(r)
		if raw == "" {
			shared.EncodeError(r.Context(), domain.ErrUnauthorized, w)
			return
		}
		claims, err := authn.AuthenticateToken(r.Context(), raw)
		if err != nil {
			shared.EncodeError(r.Context(), err, w)
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func claimsFromCtx(ctx context.Context) (domain.TokenClaims, error) {
	c, ok := ctx.Value(claimsKey).(domain.TokenClaims)
	if !ok {
		return domain.TokenClaims{}, domain.ErrUnauthorized
	}
	return c, nil
}

func decodeCredentials(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req svcendpoint.CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, domain.ErrValidation
	}
	return req, nil
}

func decodeIntrospect(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req svcendpoint.IntrospectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, domain.ErrValidation
	}
	return req, nil
}

func decodeRefresh(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req svcendpoint.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, domain.ErrValidation
	}
	return req, nil
}

func decodeClaimsOnly(ctx context.Context, _ *http.Request) (interface{}, error) {
	claims, err := claimsFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	return svcendpoint.ClaimsRequest{Claims: claims}, nil
}

func decodeLogout(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	claims, err := claimsFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, domain.ErrValidation
	}
	return svcendpoint.LogoutRequest{Claims: claims, RefreshToken: body.RefreshToken}, nil
}

func decodeUpdateProfile(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	claims, err := claimsFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var body struct {
		FullName *string `json:"full_name"`
		Phone    *string `json:"phone"`
		Address  *string `json:"address"`
		Bio      *string `json:"bio"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, domain.ErrValidation
	}
	return svcendpoint.UpdateProfileRequest{
		Claims:   claims,
		FullName: body.FullName,
		Phone:    body.Phone,
		Address:  body.Address,
		Bio:      body.Bio,
	}, nil
}

func decodeChangePassword(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	claims, err := claimsFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var body svcendpoint.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, domain.ErrValidation
	}
	return struct {
		Claims domain.TokenClaims
		Body   svcendpoint.ChangePasswordRequest
	}{Claims: claims, Body: body}, nil
}

func encodeCreated(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(http.StatusCreated)
	return shared.EncodeJSONResponse(ctx, w, response)
}
