package httptransport

import (
	"net/http"

	"github.com/go-kit/kit/endpoint"

	authhttp "github.com/yetbota/identity-service/internal/transport/http/auth"
	internalapi "github.com/yetbota/identity-service/internal/transport/http/internalapi"
	"github.com/yetbota/identity-service/internal/transport/http/shared"
	"github.com/yetbota/identity-service/internal/transport/http/swagger"
	usershttp "github.com/yetbota/identity-service/internal/transport/http/users"
)

func NewRouter(
	authEps authhttp.Endpoints,
	userEps usershttp.Endpoints,
	authn shared.TokenAuthenticator,
	recordReputation endpoint.Endpoint,
	syncReputation endpoint.Endpoint,
	serviceToken string,
) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/auth/", http.StripPrefix("/auth", authhttp.NewHandler(authEps, authn)))
	mux.Handle("/users/", http.StripPrefix("/users", usershttp.NewHandler(userEps, authn)))
	internalapi.Mount(mux, recordReputation, syncReputation, serviceToken)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	swagger.Mount(mux)
	return mux
}
