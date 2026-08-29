package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	httpserver "github.com/yetbota/identity-service/cmd/http"
	"github.com/yetbota/identity-service/drivers/auth"
	"github.com/yetbota/identity-service/drivers/config"
	"github.com/yetbota/identity-service/drivers/postgres"
	"github.com/yetbota/identity-service/internal/services/endpoint"
	userrepo "github.com/yetbota/identity-service/internal/services/repository/user"
	authuc "github.com/yetbota/identity-service/internal/services/usecase/auth"
	httptransport "github.com/yetbota/identity-service/internal/transport/http"
	authhttp "github.com/yetbota/identity-service/internal/transport/http/auth"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.MigrateOnStart {
		if err := postgres.Migrate(cfg.Postgres.URL, "migrations"); err != nil {
			log.Fatalf("migrate: %v", err)
		}
	}

	pool, err := postgres.NewPool(ctx, cfg.Postgres.URL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	repo := userrepo.New(pool)
	hasher := auth.NewBcryptHasher()
	tokens := auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.AccessTTLMinutes)
	svc := authuc.NewService(repo, hasher, tokens, repo, repo, cfg.JWT.RefreshTTLDays)

	eps := authhttp.Endpoints{
		SignUp:         endpoint.MakeSignUpEndpoint(svc),
		SignIn:         endpoint.MakeSignInEndpoint(svc),
		Me:             endpoint.MakeMeEndpoint(svc),
		UpdateProfile:  endpoint.MakeUpdateProfileEndpoint(svc),
		Introspect:     endpoint.MakeIntrospectEndpoint(svc),
		Logout:         endpoint.MakeLogoutEndpoint(svc),
		ChangePassword: endpoint.MakeChangePasswordEndpoint(svc),
		Refresh:        endpoint.MakeRefreshEndpoint(svc),
		DeleteAccount:  endpoint.MakeDeleteAccountEndpoint(svc),
	}

	router := httptransport.NewRouter(eps, svc)

	fmt.Printf("identity-service listening on %s\n", cfg.HTTP.Addr)
	if err := httpserver.ListenAndServe(ctx, cfg.HTTP.Addr, router); err != nil {
		log.Fatalf("http: %v", err)
	}
}
