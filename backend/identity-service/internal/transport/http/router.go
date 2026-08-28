package httptransport

import (
	"net/http"

	"github.com/go-kit/kit/endpoint"

	authhttp "github.com/yetbota/identity-service/internal/transport/http/auth"
)

func NewRouter(signUp, signIn endpoint.Endpoint) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/auth/", http.StripPrefix("/auth", authhttp.NewHandler(signUp, signIn)))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}
