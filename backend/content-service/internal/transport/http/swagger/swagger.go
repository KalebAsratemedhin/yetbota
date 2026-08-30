package swagger

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/yetbota/content-service/api"
)

func Mount(mux *http.ServeMux) {
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(api.OpenAPI)
	})
	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/openapi.yaml"),
	))
}
