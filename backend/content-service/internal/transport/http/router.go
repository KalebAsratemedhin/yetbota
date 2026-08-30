package httptransport

import (
	"net/http"

	postshttp "github.com/yetbota/content-service/internal/transport/http/posts"
	"github.com/yetbota/content-service/internal/transport/http/shared"
	"github.com/yetbota/content-service/internal/transport/http/swagger"
)

func NewRouter(eps postshttp.Endpoints, authn shared.TokenAuthenticator) http.Handler {
	mux := http.NewServeMux()
	postshttp.Mount(mux, eps, authn)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	swagger.Mount(mux)
	return mux
}
