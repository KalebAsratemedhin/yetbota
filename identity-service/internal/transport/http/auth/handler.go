package auth

import (
	"net/http"

	driverMW "github.com/beka-birhanu/yetbota/identity-service/drivers/middleware"
	domainAuth "github.com/beka-birhanu/yetbota/identity-service/internal/domain/auth"
	"github.com/beka-birhanu/yetbota/identity-service/internal/services/endpoint"
	"github.com/beka-birhanu/yetbota/identity-service/internal/transport/http/shared"
	kithttp "github.com/go-kit/kit/transport/http"
)

type Config struct {
	E              *endpoint.Endpoints
	SessionManager domainAuth.SessionManager
}

func NewHandler(cfg *Config) (http.Handler, error) {
	mux := http.NewServeMux()

	if cfg != nil && cfg.E != nil {
		loginServer := kithttp.NewServer(
			cfg.E.Login,
			decodeLoginHTTP,
			encodeLoginHTTP,
			shared.ServerOptions()...,
		)
		refreshServer := kithttp.NewServer(
			cfg.E.Refresh,
			decodeRefreshHTTP,
			encodeRefreshHTTP,
			shared.ServerOptions()...,
		)
		logoutEndpoint := cfg.E.Logout
		if cfg.SessionManager != nil {
			logoutEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(logoutEndpoint)
		}
		logoutServer := kithttp.NewServer(
			logoutEndpoint,
			decodeLogoutHTTP,
			encodeLogoutHTTP,
			shared.ServerOptions()...,
		)
		authorizationEndpoint := cfg.E.Authorization
		changePasswordEndpoint := cfg.E.ChangePassword
		if cfg.SessionManager != nil {
			authorizationEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(authorizationEndpoint)
			changePasswordEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(changePasswordEndpoint)
		}
		authorizationServer := kithttp.NewServer(
			authorizationEndpoint,
			decodeAuthorizationHTTP,
			encodeAuthorizationHTTP,
			shared.ServerOptions()...,
		)
		changePasswordServer := kithttp.NewServer(
			changePasswordEndpoint,
			decodeChangePasswordHTTP,
			encodeChangePasswordHTTP,
			shared.ServerOptions()...,
		)

		mux.Handle("POST /login", loginServer)
		mux.Handle("POST /refresh", refreshServer)
		mux.Handle("POST /logout", logoutServer)
		mux.Handle("POST /authorization", authorizationServer)
		mux.Handle("POST /password/change", changePasswordServer)
	}

	return mux, nil
}
