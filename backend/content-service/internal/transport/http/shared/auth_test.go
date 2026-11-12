package shared

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yetbota/content-service/internal/domain/apperr"
	"github.com/yetbota/content-service/internal/domain/auth"
)

type stubAuthn struct {
	principal auth.Principal
	err       error
}

func (s stubAuthn) AuthenticateToken(ctx context.Context, token string) (auth.Principal, error) {
	if s.err != nil {
		return auth.Principal{}, s.err
	}
	return s.principal, nil
}

func TestOptionalAuth_MissingTokenAllowsAnonymous(t *testing.T) {
	called := false
	h := OptionalAuth(stubAuthn{}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if _, ok := PrincipalFromCtxOptional(r.Context()); ok {
			t.Fatal("expected no principal")
		}
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/feed", nil))
	if !called || rr.Code != http.StatusOK {
		t.Fatalf("called=%v status=%d", called, rr.Code)
	}
}

func TestOptionalAuth_InvalidTokenUnauthorized(t *testing.T) {
	h := OptionalAuth(stubAuthn{err: apperr.ErrUnauthorized}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next must not run")
	}))
	req := httptest.NewRequest(http.MethodGet, "/feed", nil)
	req.Header.Set("Authorization", "Bearer bad")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestOptionalAuth_ValidTokenSetsPrincipal(t *testing.T) {
	want := auth.Principal{UserID: "11111111-1111-1111-1111-111111111111", Username: "u"}
	h := OptionalAuth(stubAuthn{principal: want}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, ok := PrincipalFromCtxOptional(r.Context())
		if !ok || got.UserID != want.UserID {
			t.Fatalf("got=%v ok=%v", got, ok)
		}
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/feed", nil)
	req.Header.Set("Authorization", "Bearer good")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestRequireAuth_MissingTokenUnauthorized(t *testing.T) {
	h := RequireAuth(stubAuthn{}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next must not run")
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/posts", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rr.Code)
	}
}
