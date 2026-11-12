package shared

import (
	"context"
	"net/http"
	"strings"

	"github.com/yetbota/content-service/internal/domain/apperr"
	"github.com/yetbota/content-service/internal/domain/auth"
)

type TokenAuthenticator interface {
	AuthenticateToken(ctx context.Context, token string) (auth.Principal, error)
}

type principalCtxKey int

const principalKey principalCtxKey = 1

func RequireAuth(authn TokenAuthenticator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := BearerToken(r)
		if raw == "" {
			EncodeError(r.Context(), apperr.ErrUnauthorized, w)
			return
		}
		principal, err := authn.AuthenticateToken(r.Context(), raw)
		if err != nil {
			EncodeError(r.Context(), err, w)
			return
		}
		ctx := context.WithValue(r.Context(), principalKey, principal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth allows anonymous access when Authorization is absent.
// A present but invalid/malformed Bearer token still returns 401.
func OptionalAuth(authn TokenAuthenticator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := BearerToken(r)
		if raw == "" {
			next.ServeHTTP(w, r)
			return
		}
		principal, err := authn.AuthenticateToken(r.Context(), raw)
		if err != nil {
			EncodeError(r.Context(), err, w)
			return
		}
		ctx := context.WithValue(r.Context(), principalKey, principal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func BearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func PrincipalFromCtx(ctx context.Context) (auth.Principal, error) {
	p, ok := ctx.Value(principalKey).(auth.Principal)
	if !ok {
		return auth.Principal{}, apperr.ErrUnauthorized
	}
	return p, nil
}

func PrincipalFromCtxOptional(ctx context.Context) (auth.Principal, bool) {
	p, ok := ctx.Value(principalKey).(auth.Principal)
	return p, ok
}
