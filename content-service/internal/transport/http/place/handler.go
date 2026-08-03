package place

import (
	"net/http"

	gkendpoint "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	driverMW "github.com/beka-birhanu/yetbota/content-service/drivers/middleware"
	domainAuth "github.com/beka-birhanu/yetbota/content-service/internal/domain/auth"
	"github.com/beka-birhanu/yetbota/content-service/internal/services/endpoint"
	"github.com/beka-birhanu/yetbota/content-service/internal/transport/http/shared"
)

type Config struct {
	E              *endpoint.Endpoints
	SessionManager domainAuth.SessionManager
	PathPrefix     string
}

func NewHandler(cfg *Config) (http.Handler, error) {
	mux := http.NewServeMux()

	if cfg != nil && cfg.E != nil && cfg.SessionManager != nil {
		auth := func(method, localPattern string) gkendpoint.Middleware {
			return driverMW.SelectAuthMiddleware(cfg.SessionManager, method+" "+cfg.PathPrefix+localPattern)
		}
		opts := shared.ServerOptions()

		mux.Handle("POST /", kithttp.NewServer(auth("POST", "/")(cfg.E.PlaceAdd), decodePlaceAddHTTP, encodePlaceAddHTTP, opts...))
		mux.Handle("GET /", kithttp.NewServer(auth("GET", "/")(cfg.E.PlaceList), decodePlaceListHTTP, encodePlaceListHTTP, opts...))
		mux.Handle("GET /{id}", kithttp.NewServer(auth("GET", "/{id}")(cfg.E.PlaceRead), decodePlaceReadHTTP, encodePlaceReadHTTP, opts...))
		mux.Handle("PATCH /{id}", kithttp.NewServer(auth("PATCH", "/{id}")(cfg.E.PlaceUpdate), decodePlaceUpdateHTTP, encodePlaceUpdateHTTP, opts...))
		mux.Handle("POST /{id}/bookmark", kithttp.NewServer(auth("POST", "/{id}/bookmark")(cfg.E.PlaceBookmark), decodePlaceBookmarkHTTP, encodePlaceBookmarkHTTP, opts...))
		mux.Handle("DELETE /{id}/bookmark", kithttp.NewServer(auth("DELETE", "/{id}/bookmark")(cfg.E.PlaceUnbookmark), decodePlaceUnbookmarkHTTP, encodePlaceBookmarkHTTP, opts...))
		mux.Handle("POST /{id}/confirm", kithttp.NewServer(auth("POST", "/{id}/confirm")(cfg.E.PlaceConfirm), decodePlaceConfirmHTTP, encodePlaceConfirmHTTP, opts...))
		mux.Handle("POST /{id}/report-problem", kithttp.NewServer(auth("POST", "/{id}/report-problem")(cfg.E.PlaceReportProblem), decodePlaceReportProblemHTTP, encodePlaceReportProblemHTTP, opts...))
		mux.Handle("POST /{id}/contributions", kithttp.NewServer(auth("POST", "/{id}/contributions")(cfg.E.ContributionAdd), decodePlaceContributionAddHTTP, encodeContributionAddHTTP, opts...))
		mux.Handle("GET /{id}/contributions", kithttp.NewServer(auth("GET", "/{id}/contributions")(cfg.E.ContributionListByPlace), decodePlaceContributionListHTTP, encodeContributionListHTTP, opts...))
	}

	return mux, nil
}
