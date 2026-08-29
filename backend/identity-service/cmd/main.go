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
	"github.com/yetbota/identity-service/drivers/cloudinary"
	"github.com/yetbota/identity-service/drivers/config"
	"github.com/yetbota/identity-service/drivers/envfile"
	"github.com/yetbota/identity-service/drivers/logger"
	"github.com/yetbota/identity-service/drivers/postgres"
	"github.com/yetbota/identity-service/drivers/surreal"
	"github.com/yetbota/identity-service/internal/services/endpoint"
	followrepo "github.com/yetbota/identity-service/internal/services/repository/follow"
	userrepo "github.com/yetbota/identity-service/internal/services/repository/user"
	authuc "github.com/yetbota/identity-service/internal/services/usecase/auth"
	socialuc "github.com/yetbota/identity-service/internal/services/usecase/social"
	httptransport "github.com/yetbota/identity-service/internal/transport/http"
	authhttp "github.com/yetbota/identity-service/internal/transport/http/auth"
	"github.com/yetbota/identity-service/internal/transport/http/shared"
	usershttp "github.com/yetbota/identity-service/internal/transport/http/users"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	log := logger.New()
	logger.SetDefault(log)

	if err := envfile.LoadDotEnv(); err != nil {
		log.Error("dotenv_load_failed", "err", err.Error())
		os.Exit(1)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Error("config_load_failed", "err", err.Error())
		os.Exit(1)
	}
	images := cloudinary.New(cfg.Cloudinary)
	log.Info("config_loaded",
		"http_addr", cfg.HTTP.Addr,
		"migrate_on_start", cfg.MigrateOnStart,
		"cors_origins", len(cfg.CORS.AllowedOrigins),
		"access_ttl_minutes", cfg.JWT.AccessTTLMinutes,
		"refresh_ttl_days", cfg.JWT.RefreshTTLDays,
		"surreal_url", cfg.Surreal.URL,
		"surreal_ns", cfg.Surreal.Namespace,
		"surreal_db", cfg.Surreal.Database,
		"cloudinary_configured", images.Configured(),
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

	sdb, err := surreal.Connect(ctx, cfg.Surreal)
	if err != nil {
		log.Error("surreal_failed", "err", err.Error())
		os.Exit(1)
	}
	defer func() { _ = sdb.Close(ctx) }()
	log.Info("surreal_connected")

	if err := surreal.EnsureSchema(ctx, sdb); err != nil {
		log.Error("surreal_schema_failed", "err", err.Error())
		os.Exit(1)
	}
	log.Info("surreal_schema_ok")

	repo := userrepo.New(pool)
	follows := followrepo.NewStore(sdb)
	hasher := auth.NewBcryptHasher()
	tokens := auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.AccessTTLMinutes)
	authSvc := authuc.NewService(repo, hasher, tokens, repo, repo, images, cfg.JWT.RefreshTTLDays)
	socialSvc := socialuc.NewService(repo, follows)

	authEps := authhttp.Endpoints{
		SignUp:         endpoint.MakeSignUpEndpoint(authSvc),
		SignIn:         endpoint.MakeSignInEndpoint(authSvc),
		Me:             endpoint.MakeMeEndpoint(authSvc),
		UpdateProfile:  endpoint.MakeUpdateProfileEndpoint(authSvc),
		UploadAvatar:   endpoint.MakeUploadAvatarEndpoint(authSvc),
		DeleteAvatar:   endpoint.MakeDeleteAvatarEndpoint(authSvc),
		Introspect:     endpoint.MakeIntrospectEndpoint(authSvc),
		Logout:         endpoint.MakeLogoutEndpoint(authSvc),
		ChangePassword: endpoint.MakeChangePasswordEndpoint(authSvc),
		Refresh:        endpoint.MakeRefreshEndpoint(authSvc),
		DeleteAccount:  endpoint.MakeDeleteAccountEndpoint(authSvc),
	}
	userEps := usershttp.Endpoints{
		GetProfile: endpoint.MakeGetPublicProfileEndpoint(socialSvc),
		Search:     endpoint.MakeSearchUsersEndpoint(socialSvc),
		Follow:     endpoint.MakeFollowEndpoint(socialSvc),
		Unfollow:   endpoint.MakeUnfollowEndpoint(socialSvc),
		Followers:  endpoint.MakeFollowersEndpoint(socialSvc),
		Following:  endpoint.MakeFollowingEndpoint(socialSvc),
	}

	router := httptransport.NewRouter(authEps, userEps, authSvc)
	handler := shared.AccessLog(shared.CORS(cfg.CORS.AllowedOrigins, router))

	log.Info("listening", "addr", cfg.HTTP.Addr)
	if err := httpserver.ListenAndServe(ctx, cfg.HTTP.Addr, handler); err != nil {
		log.Error("http_server_failed", "err", err.Error())
		os.Exit(1)
	}
	slog.Info("shutdown_complete")
}
