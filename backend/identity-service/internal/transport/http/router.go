package httptransport

import (
	"net/http"

	authhttp "github.com/yetbota/identity-service/internal/transport/http/auth"
	"github.com/yetbota/identity-service/internal/transport/http/shared"
	"github.com/yetbota/identity-service/internal/transport/http/swagger"
	usershttp "github.com/yetbota/identity-service/internal/transport/http/users"
)

func NewRouter(authEps authhttp.Endpoints, userEps usershttp.Endpoints, authn shared.TokenAuthenticator) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/auth/", http.StripPrefix("/auth", authhttp.NewHandler(authEps, authn)))
	mux.Handle("/users/", http.StripPrefix("/users", usershttp.NewHandler(userEps, authn)))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	swagger.Mount(mux)
	return mux
}
