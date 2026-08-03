package moderation

import (
	"net/http"

	driverMW "github.com/beka-birhanu/yetbota/content-service/drivers/middleware"
	domainAuth "github.com/beka-birhanu/yetbota/content-service/internal/domain/auth"
	"github.com/beka-birhanu/yetbota/content-service/internal/services/endpoint"
	"github.com/beka-birhanu/yetbota/content-service/internal/transport/http/shared"
	kithttp "github.com/go-kit/kit/transport/http"
)

type Config struct {
	E              *endpoint.Endpoints
	SessionManager domainAuth.SessionManager
}

func NewHandler(cfg *Config) (http.Handler, error) {
	mux := http.NewServeMux()

	if cfg != nil && cfg.E != nil {
		listEndpoint := cfg.E.ModerationListCases
		getEndpoint := cfg.E.ModerationGetCase
		actEndpoint := cfg.E.ModerationAct

		if cfg.SessionManager != nil {
			listEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(listEndpoint)
			getEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(getEndpoint)
			actEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(actEndpoint)
		}

		mux.Handle("GET /cases", kithttp.NewServer(listEndpoint, decodeListCasesHTTP, encodeListCasesHTTP, shared.ServerOptions()...))
		mux.Handle("GET /cases/{id}", kithttp.NewServer(getEndpoint, decodeGetCaseHTTP, encodeGetCaseHTTP, shared.ServerOptions()...))
		mux.Handle("POST /cases/{id}/actions", kithttp.NewServer(actEndpoint, decodeActHTTP, encodeActHTTP, shared.ServerOptions()...))
	}

	return mux, nil
}
