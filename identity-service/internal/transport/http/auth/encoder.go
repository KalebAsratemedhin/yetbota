package auth

import (
	"context"
	"encoding/json"
	"net/http"

	toddlerr "github.com/beka-birhanu/toddler/error"
	ctxYB "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
	authSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/auth"
	"github.com/beka-birhanu/yetbota/identity-service/internal/transport/http/shared"
)

type tokenData struct {
	AccessToken     string `json:"access_token"`
	AccessTokenTTL  int64  `json:"access_token_ttl"`
	RefreshToken    string `json:"refresh_token"`
	RefreshTokenTTL int64  `json:"refresh_token_ttl"`
}

func setCtxResponse(ctx context.Context, env shared.Envelope) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.Response = env
}

func encodeLoginHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*authSvc.LoginResponse)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{
		Success: true,
		Data: tokenData{
			AccessToken:     out.AccessToken,
			AccessTokenTTL:  out.AccessTokenTTL,
			RefreshToken:    out.RefreshToken,
			RefreshTokenTTL: out.RefreshTokenTTL,
		},
	}
	setCtxResponse(ctx, env)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeRefreshHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*authSvc.RefreshResponse)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{
		Success: true,
		Data: tokenData{
			AccessToken:     out.AccessToken,
			AccessTokenTTL:  out.AccessTokenTTL,
			RefreshToken:    out.RefreshToken,
			RefreshTokenTTL: out.RefreshTokenTTL,
		},
	}
	setCtxResponse(ctx, env)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeLogoutHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	if _, ok := resp.(*authSvc.LogoutResponse); !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{Success: true}
	setCtxResponse(ctx, env)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeAuthorizationHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	if _, ok := resp.(*authSvc.AuthorizationResponse); !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{Success: true}
	setCtxResponse(ctx, env)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeChangePasswordHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	if _, ok := resp.(*authSvc.ChangePasswordResponse); !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{Success: true}
	setCtxResponse(ctx, env)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}
