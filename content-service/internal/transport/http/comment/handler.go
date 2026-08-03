package comment

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

		mux.Handle("POST /", kithttp.NewServer(auth("POST", "/")(cfg.E.CommentAdd), decodeCommentAddHTTP, encodeCommentAddHTTP, opts...))
		mux.Handle("GET /", kithttp.NewServer(auth("GET", "/")(cfg.E.CommentList), decodeCommentListHTTP, encodeCommentListHTTP, opts...))
		mux.Handle("GET /{id}", kithttp.NewServer(auth("GET", "/{id}")(cfg.E.CommentRead), decodeCommentReadHTTP, encodeCommentReadHTTP, opts...))
		mux.Handle("DELETE /{id}", kithttp.NewServer(auth("DELETE", "/{id}")(cfg.E.CommentDelete), decodeCommentDeleteHTTP, encodeCommentDeleteHTTP, opts...))
		mux.Handle("POST /{id}/vote", kithttp.NewServer(auth("POST", "/{id}/vote")(cfg.E.CommentVote), decodeCommentVoteHTTP, encodeCommentVoteHTTP, opts...))
	}

	return mux, nil
}
