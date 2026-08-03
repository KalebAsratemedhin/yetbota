package userdevice

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
		registerEndpoint := cfg.E.UserDeviceRegister
		if cfg.SessionManager != nil {
			registerEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(registerEndpoint)
		}
		registerServer := kithttp.NewServer(
			registerEndpoint,
			decodeRegisterDeviceHTTP,
			encodeRegisterDeviceHTTP,
			shared.ServerOptions()...,
		)
		mux.Handle("POST /", registerServer)
	}

	return mux, nil
}
