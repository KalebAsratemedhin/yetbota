package admin

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
		statsEp := cfg.E.AdminOverviewStats
		growthEp := cfg.E.AdminOverviewGrowth
		userStatsEp := cfg.E.AdminUserStats
		auditEp := cfg.E.AdminSystemAudit

		if cfg.SessionManager != nil {
			mw := driverMW.AuthMiddleware(cfg.SessionManager)
			statsEp = mw(statsEp)
			growthEp = mw(growthEp)
			userStatsEp = mw(userStatsEp)
			auditEp = mw(auditEp)
		}

		mux.Handle("GET /overview/stats", kithttp.NewServer(statsEp, decodeNoParams, encodeOverviewStatsHTTP, shared.ServerOptions()...))
		mux.Handle("GET /overview/growth", kithttp.NewServer(growthEp, decodeOverviewGrowthHTTP, encodeOverviewGrowthHTTP, shared.ServerOptions()...))
		mux.Handle("GET /users/stats", kithttp.NewServer(userStatsEp, decodeNoParams, encodeUserStatsHTTP, shared.ServerOptions()...))
		mux.Handle("GET /system/audit", kithttp.NewServer(auditEp, decodeSystemAuditHTTP, encodeSystemAuditHTTP, shared.ServerOptions()...))
	}

	return mux, nil
}
