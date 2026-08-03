package report

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
		createEndpoint := cfg.E.ReportCreate
		if cfg.SessionManager != nil {
			createEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(createEndpoint)
		}

		mux.Handle("POST /reports", kithttp.NewServer(createEndpoint, decodeReportCreateHTTP, encodeReportCreateHTTP, shared.ServerOptions()...))
	}

	return mux, nil
}
