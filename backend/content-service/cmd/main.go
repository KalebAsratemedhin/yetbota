package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	httpserver "github.com/yetbota/content-service/cmd/http"
	"github.com/yetbota/content-service/drivers/cloudinary"
	"github.com/yetbota/content-service/drivers/config"
	"github.com/yetbota/content-service/drivers/envfile"
	"github.com/yetbota/content-service/drivers/identity"
	"github.com/yetbota/content-service/drivers/logger"
	"github.com/yetbota/content-service/drivers/postgres"
	"github.com/yetbota/content-service/internal/services/endpoint"
	postrepo "github.com/yetbota/content-service/internal/services/repository/post"
	postuc "github.com/yetbota/content-service/internal/services/usecase/post"
	httptransport "github.com/yetbota/content-service/internal/transport/http"
	postshttp "github.com/yetbota/content-service/internal/transport/http/posts"
	"github.com/yetbota/content-service/internal/transport/http/shared"
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
	signer := cloudinary.NewSigner(cfg.Cloudinary)
	log.Info("config_loaded",
		"http_addr", cfg.HTTP.Addr,
		"migrate_on_start", cfg.MigrateOnStart,
		"identity_base_url", cfg.Identity.BaseURL,
		"cloudinary_configured", signer.Configured(),
		"max_images", cfg.Media.MaxImages,
		"max_av_duration_sec", cfg.Media.MaxAVDurationSec,
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

	idClient := identity.NewClient(cfg.Identity.BaseURL)
	svc := postuc.NewService(
		postrepo.New(pool),
		signer,
		cfg.Media.MaxImages,
		cfg.Media.MaxAVDurationSec,
		cfg.Media.DefaultPageSize,
		cfg.Media.MaxPageSize,
		cfg.Share.PublicBaseURL,
	)

	eps := postshttp.Endpoints{
		SignMedia:          endpoint.MakeSignMediaEndpoint(svc),
		Create:             endpoint.MakeCreatePostEndpoint(svc),
		Get:                endpoint.MakeGetPostEndpoint(svc),
		ListUser:           endpoint.MakeListUserPostsEndpoint(svc),
		SetReaction:        endpoint.MakeSetReactionEndpoint(svc),
		AddComment:         endpoint.MakeAddCommentEndpoint(svc),
		ListComments:       endpoint.MakeListCommentsEndpoint(svc),
		ListReplies:        endpoint.MakeListRepliesEndpoint(svc),
		SetCommentReaction: endpoint.MakeSetCommentReactionEndpoint(svc),
		AddBookmark:        endpoint.MakeAddBookmarkEndpoint(svc),
		RemoveBookmark:     endpoint.MakeRemoveBookmarkEndpoint(svc),
		ListBookmarks:      endpoint.MakeListBookmarksEndpoint(svc),
		Share:              endpoint.MakeShareEndpoint(svc),
		Search:             endpoint.MakeSearchEndpoint(svc),
		ListHashtag:        endpoint.MakeListByHashtagEndpoint(svc),
	}

	router := httptransport.NewRouter(eps, idClient)
	handler := shared.AccessLog(shared.CORS(cfg.CORS.AllowedOrigins, router))

	log.Info("listening", "addr", cfg.HTTP.Addr)
	if err := httpserver.ListenAndServe(ctx, cfg.HTTP.Addr, handler); err != nil {
		log.Error("http_server_failed", "err", err.Error())
		os.Exit(1)
	}
	slog.Info("shutdown_complete")
}
