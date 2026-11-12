package shared

import (
	"context"
	"net/http"
	"strings"

	"github.com/yetbota/identity-service/internal/domain/apperr"
	domain "github.com/yetbota/identity-service/internal/domain/user"
)

type TokenAuthenticator interface {
	AuthenticateToken(ctx context.Context, token string) (domain.TokenClaims, error)
}

type claimsCtxKey int

const claimsKey claimsCtxKey = 1

func RequireAuth(authn TokenAuthenticator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := BearerToken(r)
		if raw == "" {
			EncodeError(r.Context(), apperr.ErrUnauthorized, w)
			return
		}
		claims, err := authn.AuthenticateToken(r.Context(), raw)
		if err != nil {
			EncodeError(r.Context(), err, w)
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth allows anonymous access when Authorization is absent.
// A present but invalid Bearer token still returns 401.
func OptionalAuth(authn TokenAuthenticator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := BearerToken(r)
		if raw == "" {
			next.ServeHTTP(w, r)
			return
		}
		claims, err := authn.AuthenticateToken(r.Context(), raw)
		if err != nil {
			EncodeError(r.Context(), err, w)
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
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

func ClaimsFromCtx(ctx context.Context) (domain.TokenClaims, error) {
	c, ok := ctx.Value(claimsKey).(domain.TokenClaims)
	if !ok {
		return domain.TokenClaims{}, apperr.ErrUnauthorized
	}
	return c, nil
}

func ClaimsFromCtxOptional(ctx context.Context) (domain.TokenClaims, bool) {
	c, ok := ctx.Value(claimsKey).(domain.TokenClaims)
	return c, ok
}
