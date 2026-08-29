package httptransport

import (
	"net/http"

	authhttp "github.com/yetbota/identity-service/internal/transport/http/auth"
	"github.com/yetbota/identity-service/internal/transport/http/swagger"
)

func NewRouter(eps authhttp.Endpoints, authn authhttp.TokenAuthenticator) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/auth/", http.StripPrefix("/auth", authhttp.NewHandler(eps, authn)))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	swagger.Mount(mux)
	return mux
}
