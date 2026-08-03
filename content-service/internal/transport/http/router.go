package http

import (
	"net/http"
	"time"

	logger "github.com/beka-birhanu/yetbota/content-service/drivers/logger"
	domainAuth "github.com/beka-birhanu/yetbota/content-service/internal/domain/auth"
	"github.com/beka-birhanu/yetbota/content-service/internal/services/endpoint"
	httpAdmin "github.com/beka-birhanu/yetbota/content-service/internal/transport/http/admin"
	httpComment "github.com/beka-birhanu/yetbota/content-service/internal/transport/http/comment"
	httpFeed "github.com/beka-birhanu/yetbota/content-service/internal/transport/http/feed"
	httpModeration "github.com/beka-birhanu/yetbota/content-service/internal/transport/http/moderation"
	httpNotification "github.com/beka-birhanu/yetbota/content-service/internal/transport/http/notification"
	httpContribution "github.com/beka-birhanu/yetbota/content-service/internal/transport/http/contribution"
	httpInternal "github.com/beka-birhanu/yetbota/content-service/internal/transport/http/internalapi"
	httpPlace "github.com/beka-birhanu/yetbota/content-service/internal/transport/http/place"
	httpPlaceFeed "github.com/beka-birhanu/yetbota/content-service/internal/transport/http/placefeed"
	httpPost "github.com/beka-birhanu/yetbota/content-service/internal/transport/http/post"
	httpReport "github.com/beka-birhanu/yetbota/content-service/internal/transport/http/report"
	httpShared "github.com/beka-birhanu/yetbota/content-service/internal/transport/http/shared"
)

func loggingMiddleware(next http.Handler) http.Handler {
	log := logger.Default()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logCtx := logger.Context{
			ReqMethod: r.Method,
			ReqURI:    r.RequestURI,
		}
		ctx := logger.InjectCtx(r.Context(), logCtx)
		r = r.WithContext(ctx)

		log.Info(ctx, "http request")

		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		elapsed := time.Since(start)
		if rw.status >= 500 {
			log.Error(ctx, "http response",
				logger.Field{Key: "status", Val: rw.status},
				logger.Field{Key: "duration_ms", Val: elapsed.Milliseconds()},
			)
		} else {
			log.Info(ctx, "http response",
				logger.Field{Key: "status", Val: rw.status},
				logger.Field{Key: "duration_ms", Val: elapsed.Milliseconds()},
			)
		}
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

type Config struct {
	BasePath       string
	E              *endpoint.Endpoints
	SessionManager domainAuth.SessionManager
	CorsHosts      []string
	InternalAPIKey string
}

func NewRouter(cfg *Config) (http.Handler, error) {
	r := http.NewServeMux()

	v1 := http.NewServeMux()
	if cfg != nil && cfg.E != nil {
		postHandler, err := httpPost.NewHandler(&httpPost.Config{E: cfg.E, SessionManager: cfg.SessionManager, PathPrefix: "/posts"})
		if err != nil {
			return nil, err
		}
		v1.Handle("/posts/", http.StripPrefix("/posts", postHandler))

		placeHandler, err := httpPlace.NewHandler(&httpPlace.Config{E: cfg.E, SessionManager: cfg.SessionManager, PathPrefix: "/places"})
		if err != nil {
			return nil, err
		}
		v1.Handle("/places/", http.StripPrefix("/places", placeHandler))

		contributionHandler, err := httpContribution.NewHandler(&httpContribution.Config{E: cfg.E, SessionManager: cfg.SessionManager, PathPrefix: "/contributions"})
		if err != nil {
			return nil, err
		}
		v1.Handle("/contributions/", http.StripPrefix("/contributions", contributionHandler))

		questionsHandler, err := httpContribution.NewQuestionsHandler(&httpContribution.Config{E: cfg.E, SessionManager: cfg.SessionManager, PathPrefix: "/questions"})
		if err != nil {
			return nil, err
		}
		v1.Handle("/questions/", http.StripPrefix("/questions", questionsHandler))
		v1.Handle("/questions", http.StripPrefix("/questions", questionsHandler))

		placeFeedHandler, err := httpPlaceFeed.NewHandler(&httpPlaceFeed.Config{E: cfg.E, SessionManager: cfg.SessionManager, PathPrefix: "/place-feed"})
		if err != nil {
			return nil, err
		}
		v1.Handle("/place-feed/", http.StripPrefix("/place-feed", placeFeedHandler))
		v1.Handle("/place-feed", http.StripPrefix("/place-feed", placeFeedHandler))

		commentHandler, err := httpComment.NewHandler(&httpComment.Config{E: cfg.E, SessionManager: cfg.SessionManager, PathPrefix: "/comments"})
		if err != nil {
			return nil, err
		}
		v1.Handle("/comments/", http.StripPrefix("/comments", commentHandler))

		feedHandler, err := httpFeed.NewHandler(&httpFeed.Config{E: cfg.E, SessionManager: cfg.SessionManager, PathPrefix: "/feed"})
		if err != nil {
			return nil, err
		}
		v1.Handle("/feed/", http.StripPrefix("/feed", feedHandler))

		notificationHandler, err := httpNotification.NewHandler(&httpNotification.Config{E: cfg.E, SessionManager: cfg.SessionManager, PathPrefix: "/notifications"})
		if err != nil {
			return nil, err
		}
		v1.Handle("/notifications/", http.StripPrefix("/notifications", notificationHandler))

		reportHandler, err := httpReport.NewHandler(&httpReport.Config{E: cfg.E, SessionManager: cfg.SessionManager})
		if err != nil {
			return nil, err
		}
		v1.Handle("/reports", reportHandler)

		moderationHandler, err := httpModeration.NewHandler(&httpModeration.Config{E: cfg.E, SessionManager: cfg.SessionManager})
		if err != nil {
			return nil, err
		}
		v1.Handle("/admin/moderation/", http.StripPrefix("/admin/moderation", moderationHandler))

		adminHandler, err := httpAdmin.NewHandler(&httpAdmin.Config{E: cfg.E, SessionManager: cfg.SessionManager})
		if err != nil {
			return nil, err
		}
		v1.Handle("/admin/", http.StripPrefix("/admin", adminHandler))

		internalHandler, err := httpInternal.NewHandler(&httpInternal.Config{
			E:              cfg.E,
			InternalAPIKey: cfg.InternalAPIKey,
			PathPrefix:     "/internal",
		})
		if err != nil {
			return nil, err
		}
		v1.Handle("/internal/", http.StripPrefix("/internal", internalHandler))
	}

	if cfg != nil && cfg.BasePath != "" {
		r.Handle(cfg.BasePath+"/v1/", http.StripPrefix(cfg.BasePath+"/v1", v1))
	} else {
		r.Handle("/v1/", http.StripPrefix("/v1", v1))
	}

	var out http.Handler = r
	out = loggingMiddleware(out)
	if cfg != nil && len(cfg.CorsHosts) > 0 {
		out = httpShared.CORS(cfg.CorsHosts)(out)
	}

	return out, nil
}
