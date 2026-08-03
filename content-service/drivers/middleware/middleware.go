package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-kit/kit/endpoint"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	domainAuth "github.com/beka-birhanu/yetbota/content-service/internal/domain/auth"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
)

func httpError(w http.ResponseWriter, statusCode int, err string) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(err)
}

func httpSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func authMissingOrInvalid() error {
	return &toddlerr.Error{
		PublicStatusCode:  status.Unauthorized,
		ServiceStatusCode: status.Unauthorized,
		PublicMessage:     "Missing or invalid access token",
		ServiceMessage:    "expected Authorization: Bearer <access_token>",
	}
}

func AuthMiddleware(sessionManager domainAuth.SessionManager) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			data := ctx.Value(ctxRP.AppSession)
			ctxSess := data.(*ctxRP.Context)
			header := ctxSess.Header.(http.Header)["Authorization"]
			var authHeader string
			if len(header) > 0 {
				authHeader = strings.TrimSpace(header[0])
			}

			fields := strings.Fields(authHeader)
			if len(fields) < 2 || !strings.EqualFold(fields[0], "Bearer") {
				return nil, authMissingOrInvalid()
			}
			normalized := "Bearer " + strings.Join(fields[1:], " ")

			userSession, errExtract := sessionManager.ExtractUserSession(ctx, &domainAuth.TokenInfo{
				TokenType: domainAuth.AccessToken,
				Token:     normalized,
			})
			if errExtract != nil {
				return nil, errExtract
			}

			ctxSess.UserSession = *userSession

			return next(ctx, request)
		}
	}
}

// SelectAuthMiddleware mirrors the gRPC interceptor pattern: if the route is in
// constants.SkipAuthHTTP it gets optional auth, otherwise mandatory auth.
// fullPattern must be "METHOD /prefix/local-pattern", e.g. "GET /posts/".
func SelectAuthMiddleware(sm domainAuth.SessionManager, fullPattern string) endpoint.Middleware {
	if _, skip := constants.SkipAuthHTTP[fullPattern]; skip {
		return OptionalAuthMiddleware(sm)
	}
	return AuthMiddleware(sm)
}

// OptionalAuthMiddleware populates the session if a valid Bearer token is present; silent no-op otherwise.
func OptionalAuthMiddleware(sessionManager domainAuth.SessionManager) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			data := ctx.Value(ctxRP.AppSession)
			ctxSess, ok := data.(*ctxRP.Context)
			if !ok || ctxSess == nil {
				return next(ctx, request)
			}
			header, _ := ctxSess.Header.(http.Header)
			var authHeader string
			if vals := header["Authorization"]; len(vals) > 0 {
				authHeader = strings.TrimSpace(vals[0])
			}
			fields := strings.Fields(authHeader)
			if len(fields) >= 2 && strings.EqualFold(fields[0], "Bearer") {
				normalized := "Bearer " + strings.Join(fields[1:], " ")
				if userSession, err := sessionManager.ExtractUserSession(ctx, &domainAuth.TokenInfo{
					TokenType: domainAuth.AccessToken,
					Token:     normalized,
				}); err == nil {
					ctxSess.UserSession = *userSession
				}
			}
			return next(ctx, request)
		}
	}
}

func TokenVerify(sessionManager domainAuth.SessionManager) func(context.Context, http.Handler) http.Handler {
	return func(ctx context.Context, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			fields := strings.Fields(authHeader)
			if len(fields) < 2 || !strings.EqualFold(fields[0], "Bearer") {
				httpError(w, http.StatusUnauthorized, "Invalid token")
				return
			}
			normalized := "Bearer " + strings.Join(fields[1:], " ")

			userSession, err := sessionManager.ExtractUserSession(r.Context(), &domainAuth.TokenInfo{
				TokenType: domainAuth.AccessToken,
				Token:     normalized,
			})
			if err != nil {
				httpError(w, http.StatusUnauthorized, err.Error())
				return
			}

			_ = userSession
			next.ServeHTTP(w, r)
		})
	}
}
