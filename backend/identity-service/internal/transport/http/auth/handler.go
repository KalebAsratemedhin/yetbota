package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/yetbota/identity-service/internal/domain/apperr"
	domain "github.com/yetbota/identity-service/internal/domain/user"
	svcendpoint "github.com/yetbota/identity-service/internal/services/endpoint"
	"github.com/yetbota/identity-service/internal/transport/http/shared"
)

type TokenAuthenticator = shared.TokenAuthenticator

type Endpoints struct {
	SignUp         endpoint.Endpoint
	SignIn         endpoint.Endpoint
	Me             endpoint.Endpoint
	UpdateProfile  endpoint.Endpoint
	UploadAvatar   endpoint.Endpoint
	DeleteAvatar   endpoint.Endpoint
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

	mux.Handle("GET /me", shared.RequireAuth(authn, kithttp.NewServer(eps.Me, decodeClaimsOnly, shared.EncodeJSONResponse, opts...)))
	mux.Handle("PATCH /me", shared.RequireAuth(authn, kithttp.NewServer(eps.UpdateProfile, decodeUpdateProfile, shared.EncodeJSONResponse, opts...)))
	mux.Handle("POST /me/avatar", shared.RequireAuth(authn, kithttp.NewServer(eps.UploadAvatar, decodeUploadAvatar, shared.EncodeJSONResponse, opts...)))
	mux.Handle("DELETE /me/avatar", shared.RequireAuth(authn, kithttp.NewServer(eps.DeleteAvatar, decodeClaimsOnly, shared.EncodeJSONResponse, opts...)))
	mux.Handle("POST /logout", shared.RequireAuth(authn, kithttp.NewServer(eps.Logout, decodeLogout, shared.EncodeJSONResponse, opts...)))
	mux.Handle("POST /password", shared.RequireAuth(authn, kithttp.NewServer(eps.ChangePassword, decodeChangePassword, shared.EncodeJSONResponse, opts...)))
	mux.Handle("DELETE /me", shared.RequireAuth(authn, kithttp.NewServer(eps.DeleteAccount, decodeClaimsOnly, shared.EncodeJSONResponse, opts...)))

	return mux
}

func decodeCredentials(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req svcendpoint.CredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID_JSON", Message: "request body must be valid JSON"})
	}
	return req, nil
}

func decodeIntrospect(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req svcendpoint.IntrospectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID_JSON", Message: "request body must be valid JSON"})
	}
	return req, nil
}

func decodeRefresh(_ context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	var req svcendpoint.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID_JSON", Message: "request body must be valid JSON"})
	}
	return req, nil
}

func decodeClaimsOnly(ctx context.Context, _ *http.Request) (interface{}, error) {
	claims, err := shared.ClaimsFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	return svcendpoint.ClaimsRequest{Claims: claims}, nil
}

func decodeLogout(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	claims, err := shared.ClaimsFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID_JSON", Message: "request body must be valid JSON"})
	}
	return svcendpoint.LogoutRequest{Claims: claims, RefreshToken: body.RefreshToken}, nil
}

func decodeUpdateProfile(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	claims, err := shared.ClaimsFromCtx(ctx)
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
		return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID_JSON", Message: "request body must be valid JSON"})
	}
	return svcendpoint.UpdateProfileRequest{
		Claims:   claims,
		FullName: body.FullName,
		Phone:    body.Phone,
		Address:  body.Address,
		Bio:      body.Bio,
	}, nil
}

func decodeUploadAvatar(ctx context.Context, r *http.Request) (interface{}, error) {
	claims, err := shared.ClaimsFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	const maxMem = 2 << 20
	if err := r.ParseMultipartForm(maxMem); err != nil {
		return nil, apperr.Validation(apperr.Field{
			Field: "file", Code: "INVALID", Message: "expected multipart form with file field",
		})
	}
	file, hdr, err := r.FormFile("file")
	if err != nil {
		return nil, apperr.Validation(apperr.Field{
			Field: "file", Code: "REQUIRED", Message: "file is required",
		})
	}
	return svcendpoint.UploadAvatarRequest{
		Claims:   claims,
		Filename: hdr.Filename,
		Reader:   file,
		Size:     hdr.Size,
	}, nil
}

func decodeChangePassword(ctx context.Context, r *http.Request) (interface{}, error) {
	defer r.Body.Close()
	claims, err := shared.ClaimsFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var body svcendpoint.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID_JSON", Message: "request body must be valid JSON"})
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
