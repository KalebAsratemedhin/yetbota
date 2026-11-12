package shared

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yetbota/identity-service/internal/domain/apperr"
	domain "github.com/yetbota/identity-service/internal/domain/user"
)

type stubAuthn struct {
	claims domain.TokenClaims
	err    error
}

func (s stubAuthn) AuthenticateToken(ctx context.Context, token string) (domain.TokenClaims, error) {
	if s.err != nil {
		return domain.TokenClaims{}, s.err
	}
	return s.claims, nil
}

func TestOptionalAuth_MissingTokenAllowsAnonymous(t *testing.T) {
	called := false
	h := OptionalAuth(stubAuthn{}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if _, ok := ClaimsFromCtxOptional(r.Context()); ok {
			t.Fatal("expected no claims")
		}
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/users/x", nil))
	if !called || rr.Code != http.StatusOK {
		t.Fatalf("called=%v status=%d", called, rr.Code)
	}
}

func TestOptionalAuth_InvalidTokenUnauthorized(t *testing.T) {
	h := OptionalAuth(stubAuthn{err: apperr.ErrUnauthorized}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next must not run")
	}))
	req := httptest.NewRequest(http.MethodGet, "/users/x", nil)
	req.Header.Set("Authorization", "Bearer bad")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestRequireAuth_MissingTokenUnauthorized(t *testing.T) {
	h := RequireAuth(stubAuthn{}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next must not run")
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/users/x/follow", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rr.Code)
	}
}
