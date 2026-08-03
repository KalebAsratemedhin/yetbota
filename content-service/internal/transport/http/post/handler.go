package post

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

		mux.Handle("POST /", kithttp.NewServer(auth("POST", "/")(cfg.E.PostAdd), decodePostAddHTTP, encodePostAddHTTP, opts...))
		mux.Handle("GET /", kithttp.NewServer(auth("GET", "/")(cfg.E.PostList), decodePostListHTTP, encodePostListHTTP, opts...))
		mux.Handle("GET /{id}", kithttp.NewServer(auth("GET", "/{id}")(cfg.E.PostRead), decodePostReadHTTP, encodePostReadHTTP, opts...))
		mux.Handle("PATCH /{id}", kithttp.NewServer(auth("PATCH", "/{id}")(cfg.E.PostUpdate), decodePostUpdateHTTP, encodePostUpdateHTTP, opts...))
		mux.Handle("POST /{id}/vote", kithttp.NewServer(auth("POST", "/{id}/vote")(cfg.E.PostVote), decodePostVoteHTTP, encodePostVoteHTTP, opts...))
		mux.Handle("POST /{id}/save", kithttp.NewServer(auth("POST", "/{id}/save")(cfg.E.PostSave), decodePostSaveHTTP, encodePostSaveHTTP, opts...))
		mux.Handle("DELETE /{id}/save", kithttp.NewServer(auth("DELETE", "/{id}/save")(cfg.E.PostUnsave), decodePostUnsaveHTTP, encodePostUnsaveHTTP, opts...))
	}

	return mux, nil
}
