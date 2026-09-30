package httptransport

import (
	"net/http"

	postshttp "github.com/yetbota/content-service/internal/transport/http/posts"
	placeshttp "github.com/yetbota/content-service/internal/transport/http/places"
	internalapi "github.com/yetbota/content-service/internal/transport/http/internalapi"
	qahttp "github.com/yetbota/content-service/internal/transport/http/qa"
	"github.com/yetbota/content-service/internal/transport/http/shared"
	"github.com/yetbota/content-service/internal/transport/http/swagger"
)

func NewRouter(
	postsEps postshttp.Endpoints,
	qaEps qahttp.Endpoints,
	placesEps placeshttp.Endpoints,
	internalEps internalapi.Endpoints,
	authn shared.TokenAuthenticator,
	serviceToken string,
) http.Handler {
	mux := http.NewServeMux()
	postshttp.Mount(mux, postsEps, authn)
	qahttp.Mount(mux, qaEps, authn)
	placeshttp.Mount(mux, placesEps, authn)
	internalapi.Mount(mux, internalEps, serviceToken)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	swagger.Mount(mux)
	return mux
}
