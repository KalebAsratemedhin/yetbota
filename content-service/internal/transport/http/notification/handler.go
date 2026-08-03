package notification

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

		mux.Handle("GET /", kithttp.NewServer(auth("GET", "/")(cfg.E.NotificationList), decodeNotificationListHTTP, encodeNotificationListHTTP, opts...))
		mux.Handle("POST /mark-read", kithttp.NewServer(auth("POST", "/mark-read")(cfg.E.NotificationMarkAsRead), decodeNotificationMarkAsReadHTTP, encodeNotificationMarkAsReadHTTP, opts...))
		mux.Handle("DELETE /{id}", kithttp.NewServer(auth("DELETE", "/{id}")(cfg.E.NotificationDelete), decodeNotificationDeleteHTTP, encodeNotificationDeleteHTTP, opts...))
	}

	return mux, nil
}
