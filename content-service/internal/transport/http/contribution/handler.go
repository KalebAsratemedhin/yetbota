package contribution

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
		mux.Handle("POST /{id}/answers", kithttp.NewServer(auth("POST", "/{id}/answers")(cfg.E.ContributionAddAnswer), decodeContributionAddAnswerHTTP, encodeContributionAddAnswerHTTP, opts...))
		mux.Handle("POST /{id}/vote", kithttp.NewServer(auth("POST", "/{id}/vote")(cfg.E.ContributionVote), decodeContributionVoteHTTP, encodeContributionVoteHTTP, opts...))
		mux.Handle("POST /{id}/accept-answer", kithttp.NewServer(auth("POST", "/{id}/accept-answer")(cfg.E.ContributionAcceptAnswer), decodeContributionAcceptAnswerHTTP, encodeContributionAcceptAnswerHTTP, opts...))
	}
	return mux, nil
}

func NewQuestionsHandler(cfg *Config) (http.Handler, error) {
	mux := http.NewServeMux()
	if cfg != nil && cfg.E != nil && cfg.SessionManager != nil {
		auth := func(method, localPattern string) gkendpoint.Middleware {
			return driverMW.SelectAuthMiddleware(cfg.SessionManager, method+" "+cfg.PathPrefix+localPattern)
		}
		opts := shared.ServerOptions()
		mux.Handle("GET /", kithttp.NewServer(auth("GET", "/")(cfg.E.ContributionListOpenQuestions), decodeOpenQuestionsListHTTP, encodeOpenQuestionsListHTTP, opts...))
		mux.Handle("GET /near", kithttp.NewServer(auth("GET", "/near")(cfg.E.ContributionListOpenQuestions), decodeOpenQuestionsNearHTTP, encodeOpenQuestionsListHTTP, opts...))
	}
	return mux, nil
}
