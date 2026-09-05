package httptransport

import (
	"net/http"

	postshttp "github.com/yetbota/content-service/internal/transport/http/posts"
	qahttp "github.com/yetbota/content-service/internal/transport/http/qa"
	"github.com/yetbota/content-service/internal/transport/http/shared"
	"github.com/yetbota/content-service/internal/transport/http/swagger"
)

func NewRouter(postsEps postshttp.Endpoints, qaEps qahttp.Endpoints, authn shared.TokenAuthenticator) http.Handler {
	mux := http.NewServeMux()
	postshttp.Mount(mux, postsEps, authn)
	qahttp.Mount(mux, qaEps, authn)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	swagger.Mount(mux)
	return mux
}
