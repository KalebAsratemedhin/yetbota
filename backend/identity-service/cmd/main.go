package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	httpserver "github.com/yetbota/identity-service/cmd/http"
	"github.com/yetbota/identity-service/drivers/auth"
	"github.com/yetbota/identity-service/drivers/config"
	"github.com/yetbota/identity-service/drivers/logger"
	"github.com/yetbota/identity-service/drivers/postgres"
	"github.com/yetbota/identity-service/internal/services/endpoint"
	userrepo "github.com/yetbota/identity-service/internal/services/repository/user"
	authuc "github.com/yetbota/identity-service/internal/services/usecase/auth"
	httptransport "github.com/yetbota/identity-service/internal/transport/http"
	authhttp "github.com/yetbota/identity-service/internal/transport/http/auth"
	"github.com/yetbota/identity-service/internal/transport/http/shared"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	log := logger.New()
	logger.SetDefault(log)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Error("config_load_failed", "err", err.Error())
		os.Exit(1)
	}
	log.Info("config_loaded",
		"http_addr", cfg.HTTP.Addr,
		"migrate_on_start", cfg.MigrateOnStart,
		"cors_origins", len(cfg.CORS.AllowedOrigins),
		"access_ttl_minutes", cfg.JWT.AccessTTLMinutes,
		"refresh_ttl_days", cfg.JWT.RefreshTTLDays,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.MigrateOnStart {
		if err := postgres.Migrate(cfg.Postgres.URL, "migrations"); err != nil {
			log.Error("migrate_failed", "err", err.Error())
			os.Exit(1)
		}
		log.Info("migrate_ok")
	}

	pool, err := postgres.NewPool(ctx, cfg.Postgres.URL)
	if err != nil {
		log.Error("postgres_failed", "err", err.Error())
		os.Exit(1)
	}
	defer pool.Close()
	log.Info("postgres_connected")

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
	handler := shared.AccessLog(shared.CORS(cfg.CORS.AllowedOrigins, router))

	log.Info("listening", "addr", cfg.HTTP.Addr)
	if err := httpserver.ListenAndServe(ctx, cfg.HTTP.Addr, handler); err != nil {
		log.Error("http_server_failed", "err", err.Error())
		os.Exit(1)
	}
	slog.Info("shutdown_complete")
}
