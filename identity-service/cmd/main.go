package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/config"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmigrations"
	jwtDriver "github.com/beka-birhanu/yetbota/identity-service/drivers/jwt"
	logger "github.com/beka-birhanu/yetbota/identity-service/drivers/logger"
	surrealdbDriver "github.com/beka-birhanu/yetbota/identity-service/drivers/surrealdb"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/postgres"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/storage"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/utils"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	"github.com/beka-birhanu/yetbota/identity-service/internal/services/endpoint"
	repoFollow "github.com/beka-birhanu/yetbota/identity-service/internal/services/repository/follow"
	repoPhoto "github.com/beka-birhanu/yetbota/identity-service/internal/services/repository/photo"
	repoUser "github.com/beka-birhanu/yetbota/identity-service/internal/services/repository/user"
	repoDevice "github.com/beka-birhanu/yetbota/identity-service/internal/services/repository/userdevice"
	usecaseAuth "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/auth"
	usecaseUser "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/user"
	usecaseDevice "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/userdevice"
	"github.com/go-redis/redis/v8"
	"github.com/pressly/goose"
	"go.uber.org/zap/zapcore"

	cmdHTTP "github.com/beka-birhanu/yetbota/identity-service/cmd/http"
	transportHTTP "github.com/beka-birhanu/yetbota/identity-service/internal/transport/http"
)

func main() {
	validator.InitValidator()
	logger.InitDefault(
		logger.MaskEnabled(),
		logger.WithStdout(),
		logger.WithLevel(zapcore.DebugLevel),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("error load config: %v", err))
	}
	if portStr := os.Getenv("PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Rest.Port = port
		}
	}

	pgdb, err := postgres.NewDB(
		&postgres.Config{
			Host:     cfg.Postgres.Host,
			Port:     cfg.Postgres.Port,
			User:     cfg.Postgres.User,
			Password: cfg.Postgres.Password,
			DB:       cfg.Postgres.DB,
			SslMode:  cfg.Postgres.SslMode,
		})
	if err != nil {
		panic(fmt.Errorf("error connect postgres: %v", err))
	}
	defer func() {
		_ = pgdb.Close()
	}()

	boil.SetDB(pgdb)

	if err := pgdb.Ping(); err != nil {
		panic(fmt.Errorf("error pinging database: %v", err))
	}
	fmt.Println("Database connection successful!")

	dbGoose, err := dbmigrations.RunDBMigrations(&dbmigrations.Config{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		User:     cfg.Postgres.User,
		Password: cfg.Postgres.Password,
		DB:       cfg.Postgres.DB,
		SslMode:  cfg.Postgres.SslMode,
	})
	if err != nil {
		panic(fmt.Errorf("error run DB migrations: %v", err))
	}

	if err := goose.SetDialect("postgres"); err != nil {
		panic(fmt.Errorf("error setting goose dialect: %v", err))
	}
	goose.SetTableName("goose_db_version_identity")

	currentVersion, err := goose.GetDBVersion(dbGoose)
	if err != nil {
		fmt.Printf("Migration table initialization: %v\n", err)
	}
	fmt.Printf("Current migration version: %d\n", currentVersion)

	if err := goose.Up(dbGoose, constants.MigrationFolder); err != nil {
		panic(fmt.Errorf("error running migrations: %v", err))
	}

	redisOpts := &redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
	}
	if cfg.Redis.TLS {
		redisOpts.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	redisConn := redis.NewClient(redisOpts)
	if _, err := redisConn.Ping(ctx).Result(); err != nil {
		panic(fmt.Errorf("error connecting to redis: %v", err))
	}
	fmt.Println("Redis connection successful!")

	sessionManager, err := jwtDriver.NewSessionManager(&jwtDriver.Config{
		AccessKey:  cfg.Jwt.AccessToken.Secret,
		RefreshKey: cfg.Jwt.RefreshToken.Secret,
		AccessTTL:  time.Duration(cfg.Jwt.AccessToken.Expiration) * time.Second,
		RefreshTTL: time.Duration(cfg.Jwt.RefreshToken.Expiration) * time.Second,
		Algo:       cfg.Jwt.Algorithm,
		RedisConn:  redisConn,
	})
	if err != nil {
		panic(fmt.Errorf("error creating session manager: %v", err))
	}

	userRepo, err := repoUser.NewRepo(&repoUser.Config{DB: pgdb})
	if err != nil {
		panic(fmt.Errorf("error creating user repo: %v", err))
	}

	deviceRepo, err := repoDevice.NewRepo(&repoDevice.Config{DB: pgdb})
	if err != nil {
		panic(fmt.Errorf("error creating device repo: %v", err))
	}

	hasher := utils.NewHasher()

	bucket, err := storage.NewCloudinaryBlob(&storage.CloudinaryConfig{
		CloudName: cfg.Cloudinary.CloudName,
		APIKey:    cfg.Cloudinary.APIKey,
		APISecret: cfg.Cloudinary.APISecret,
		Folder:    cfg.Cloudinary.Folder,
	})
	if err != nil {
		panic(fmt.Errorf("error creating cloudinary blob: %v", err))
	}

	photoRepo, err := repoPhoto.NewRepo(&repoPhoto.Config{DB: pgdb})
	if err != nil {
		panic(fmt.Errorf("error creating photo repo: %v", err))
	}

	graphClient, err := surrealdbDriver.NewClient(context.Background(), &surrealdbDriver.Config{
		URL:       cfg.SurrealDB.URL,
		Username:  cfg.SurrealDB.Username,
		Password:  cfg.SurrealDB.Password,
		Namespace: cfg.SurrealDB.Namespace,
		Database:  cfg.SurrealDB.Database,
	})
	if err != nil {
		panic(fmt.Errorf("error creating surrealdb client: %v", err))
	}
	defer func() {
		_ = graphClient.Close(context.Background())
	}()

	followRepo, err := repoFollow.NewRepo(&repoFollow.Config{Graph: graphClient})
	if err != nil {
		panic(fmt.Errorf("error creating follow repo: %v", err))
	}

	deviceService, err := usecaseDevice.NewService(&usecaseDevice.Config{
		DeviceRepo: deviceRepo,
	})
	if err != nil {
		panic(fmt.Errorf("error creating device service: %v", err))
	}

	authService, err := usecaseAuth.NewService(&usecaseAuth.Config{
		UserRepo:       userRepo,
		SessionManager: sessionManager,
		Hasher:         hasher,
		DeviceRepo:     deviceRepo,
		AccessTTL:      time.Duration(cfg.Jwt.AccessToken.Expiration) * time.Second,
		RefreshTTL:     time.Duration(cfg.Jwt.RefreshToken.Expiration) * time.Second,
	})
	if err != nil {
		panic(fmt.Errorf("error creating auth service: %v", err))
	}

	userService, err := usecaseUser.NewService(&usecaseUser.Config{
		UserRepo:   userRepo,
		PhotoRepo:  photoRepo,
		FollowRepo: followRepo,
		Hasher:     hasher,
		Bucket:     bucket,
	})
	if err != nil {
		panic(fmt.Errorf("error creating user service: %v", err))
	}

	endpoints, err := endpoint.NewEndpoints(&endpoint.Config{
		AuthService:   authService,
		UserService:   userService,
		DeviceService: deviceService,
	})
	if err != nil {
		panic(fmt.Errorf("error creating endpoints: %v", err))
	}

	httpRouter, err := transportHTTP.NewRouter(&transportHTTP.Config{
		E:              endpoints,
		SessionManager: sessionManager,
		CorsHosts:      cfg.Cors.Hosts,
	})
	if err != nil {
		panic(fmt.Errorf("error creating http router: %v", err))
	}

	if err := cmdHTTP.RunServer(ctx, &cmdHTTP.Config{
		Port:         cfg.Rest.Port,
		Handler:      httpRouter,
		ReadTimeout:  time.Duration(cfg.Rest.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Rest.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Rest.IdleTimeout) * time.Second,
	}); err != nil {
		panic(fmt.Errorf("server error: %v", err))
	}
}
