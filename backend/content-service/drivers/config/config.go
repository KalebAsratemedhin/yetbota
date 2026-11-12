package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTP           HTTPConfig       `yaml:"http"`
	Postgres       PostgresConfig   `yaml:"postgres"`
	Surreal        SurrealConfig    `yaml:"surreal"`
	Identity       IdentityConfig   `yaml:"identity"`
	Cloudinary     CloudinaryConfig `yaml:"cloudinary"`
	Media          MediaConfig      `yaml:"media"`
	Share          ShareConfig      `yaml:"share"`
	Feed           FeedConfig       `yaml:"feed"`
	Trending       TrendingConfig   `yaml:"trending"`
	Reputation     ReputationConfig `yaml:"reputation"`
	RabbitMQ       RabbitMQConfig   `yaml:"rabbitmq"`
	CORS           CORSConfig       `yaml:"cors"`
	MigrateOnStart bool             `yaml:"migrate_on_start"`
}

type HTTPConfig struct {
	Addr string `yaml:"addr"`
}

type PostgresConfig struct {
	URL string `yaml:"url"`
}

type SurrealConfig struct {
	URL       string `yaml:"url"`
	Namespace string `yaml:"namespace"`
	Database  string `yaml:"database"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
}

type IdentityConfig struct {
	BaseURL      string `yaml:"base_url"`
	ServiceToken string `yaml:"service_token"`
}

type FeedConfig struct {
	Gravity                 float64 `yaml:"gravity"`
	CommentWeight           float64 `yaml:"comment_weight"`
	ShareWeight             float64 `yaml:"share_weight"`
	BookmarkWeight          float64 `yaml:"bookmark_weight"`
	FollowBoost             float64 `yaml:"follow_boost"`
	GeoD0Km                 float64 `yaml:"geo_d0_km"`
	NearbyRadiusKm          float64 `yaml:"nearby_radius_km"`
	CandidateLimit          int `yaml:"candidate_limit"`
	LookbackDays            int `yaml:"lookback_days"`
	SuggestedCandidateLimit int `yaml:"suggested_candidate_limit"`
}

type TrendingConfig struct {
	LookbackDays      int `yaml:"lookback_days"`
	CandidateLimit    int `yaml:"candidate_limit"`
	CreatorEventLimit int `yaml:"creator_event_limit"`
}

type ReputationConfig struct {
	LikeWeight     float64 `yaml:"like_weight"`
	CommentWeight  float64 `yaml:"comment_weight"`
	ShareWeight    float64 `yaml:"share_weight"`
	BookmarkWeight float64 `yaml:"bookmark_weight"`
	DislikeWeight  float64 `yaml:"dislike_weight"`
	CreateWeight   float64 `yaml:"create_weight"`
	Alpha          float64 `yaml:"alpha"`
	Gravity        float64 `yaml:"gravity"`
	TimeBaseHours  float64 `yaml:"time_base_hours"`
	Scale          float64 `yaml:"scale"`
	LookbackDays   int     `yaml:"lookback_days"`
	CandidateLimit int     `yaml:"candidate_limit"`
}

type CloudinaryConfig struct {
	CloudName string `yaml:"cloud_name"`
	APIKey    string `yaml:"api_key"`
	APISecret string `yaml:"api_secret"`
	Folder    string `yaml:"folder"`
}

type MediaConfig struct {
	MaxImages        int   `yaml:"max_images"`
	MaxAVDurationSec int   `yaml:"max_av_duration_sec"`
	// MaxUploadBytes matches Cloudinary Free plan upload cap (10 MiB).
	MaxUploadBytes  int64 `yaml:"max_upload_bytes"`
	DefaultPageSize int   `yaml:"default_page_size"`
	MaxPageSize     int   `yaml:"max_page_size"`
}

type ShareConfig struct {
	PublicBaseURL string `yaml:"public_base_url"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type RabbitMQConfig struct {
	URL string `yaml:"url"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if v := os.Getenv("DATABASE_URL"); v != "" {
		cfg.Postgres.URL = v
	}
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		cfg.HTTP.Addr = v
	}
	if v := os.Getenv("PORT"); v != "" {
		if strings.HasPrefix(v, ":") {
			cfg.HTTP.Addr = v
		} else {
			cfg.HTTP.Addr = ":" + v
		}
	}
	if v := os.Getenv("IDENTITY_BASE_URL"); v != "" {
		cfg.Identity.BaseURL = v
	}
	if v := os.Getenv("IDENTITY_SERVICE_TOKEN"); v != "" {
		cfg.Identity.ServiceToken = v
	}
	if v := os.Getenv("INTERNAL_SERVICE_TOKEN"); v != "" && cfg.Identity.ServiceToken == "" {
		cfg.Identity.ServiceToken = v
	}
	if v := os.Getenv("SURREAL_URL"); v != "" {
		cfg.Surreal.URL = v
	}
	if v := os.Getenv("SURREAL_NS"); v != "" {
		cfg.Surreal.Namespace = v
	}
	if v := os.Getenv("SURREAL_DB"); v != "" {
		cfg.Surreal.Database = v
	}
	if v := os.Getenv("SURREAL_USER"); v != "" {
		cfg.Surreal.Username = v
	}
	if v := os.Getenv("SURREAL_PASS"); v != "" {
		cfg.Surreal.Password = v
	}
	if v := os.Getenv("CLOUDINARY_CLOUD_NAME"); v != "" {
		cfg.Cloudinary.CloudName = v
	}
	if v := os.Getenv("CLOUDINARY_API_KEY"); v != "" {
		cfg.Cloudinary.APIKey = v
	}
	if v := os.Getenv("CLOUDINARY_API_SECRET"); v != "" {
		cfg.Cloudinary.APISecret = v
	}
	if v := os.Getenv("CLOUDINARY_FOLDER"); v != "" {
		cfg.Cloudinary.Folder = v
	}
	if v := os.Getenv("SHARE_PUBLIC_BASE_URL"); v != "" {
		cfg.Share.PublicBaseURL = v
	}
	if v := os.Getenv("MEDIA_MAX_IMAGES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Media.MaxImages = n
		}
	}
	if v := os.Getenv("MEDIA_MAX_AV_DURATION_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Media.MaxAVDurationSec = n
		}
	}
	if v := os.Getenv("MEDIA_MAX_UPLOAD_BYTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			cfg.Media.MaxUploadBytes = n
		}
	}
	if v := os.Getenv("MIGRATE_ON_START"); v != "" {
		cfg.MigrateOnStart = v == "1" || strings.EqualFold(v, "true")
	}
	if v := os.Getenv("RABBITMQ_URL"); v != "" {
		cfg.RabbitMQ.URL = v
	}
	if cfg.RabbitMQ.URL == "" {
		cfg.RabbitMQ.URL = "amqp://yetbota:yetbota@localhost:5672/"
	}
	if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
		var origins []string
		for _, part := range strings.Split(v, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				origins = append(origins, part)
			}
		}
		cfg.CORS.AllowedOrigins = origins
	}

	if cfg.HTTP.Addr == "" {
		cfg.HTTP.Addr = ":8082"
	}
	if cfg.Identity.BaseURL == "" {
		cfg.Identity.BaseURL = "http://localhost:8081"
	}
	if cfg.Identity.ServiceToken == "" {
		cfg.Identity.ServiceToken = "dev-internal-token"
	}
	if cfg.Surreal.URL == "" {
		cfg.Surreal.URL = "ws://localhost:8000"
	}
	if cfg.Surreal.Namespace == "" {
		cfg.Surreal.Namespace = "yetbota"
	}
	if cfg.Surreal.Database == "" {
		cfg.Surreal.Database = "main"
	}
	if cfg.Surreal.Username == "" {
		cfg.Surreal.Username = "root"
	}
	if cfg.Surreal.Password == "" {
		cfg.Surreal.Password = "root"
	}
	if cfg.Feed.Gravity <= 0 {
		cfg.Feed.Gravity = 1.5
	}
	if cfg.Feed.CommentWeight <= 0 {
		cfg.Feed.CommentWeight = 3
	}
	if cfg.Feed.ShareWeight <= 0 {
		cfg.Feed.ShareWeight = 4
	}
	if cfg.Feed.BookmarkWeight <= 0 {
		cfg.Feed.BookmarkWeight = 3
	}
	if cfg.Feed.FollowBoost <= 0 {
		cfg.Feed.FollowBoost = 0.35
	}
	if cfg.Feed.GeoD0Km <= 0 {
		cfg.Feed.GeoD0Km = 5
	}
	if cfg.Feed.NearbyRadiusKm <= 0 {
		cfg.Feed.NearbyRadiusKm = 25
	}
	if cfg.Feed.CandidateLimit <= 0 {
		cfg.Feed.CandidateLimit = 200
	}
	if cfg.Feed.LookbackDays <= 0 {
		cfg.Feed.LookbackDays = 14
	}
	if cfg.Feed.SuggestedCandidateLimit <= 0 {
		cfg.Feed.SuggestedCandidateLimit = 500
	}
	if cfg.Trending.LookbackDays <= 0 {
		cfg.Trending.LookbackDays = 7
	}
	if cfg.Trending.CandidateLimit <= 0 {
		cfg.Trending.CandidateLimit = 500
	}
	if cfg.Trending.CreatorEventLimit <= 0 {
		cfg.Trending.CreatorEventLimit = 5000
	}
	if cfg.Reputation.LikeWeight <= 0 {
		cfg.Reputation.LikeWeight = 1
	}
	if cfg.Reputation.CommentWeight <= 0 {
		cfg.Reputation.CommentWeight = cfg.Feed.CommentWeight
	}
	if cfg.Reputation.ShareWeight <= 0 {
		cfg.Reputation.ShareWeight = cfg.Feed.ShareWeight
	}
	if cfg.Reputation.BookmarkWeight <= 0 {
		cfg.Reputation.BookmarkWeight = cfg.Feed.BookmarkWeight
	}
	if cfg.Reputation.DislikeWeight <= 0 {
		cfg.Reputation.DislikeWeight = 1
	}
	if cfg.Reputation.CreateWeight <= 0 {
		cfg.Reputation.CreateWeight = 0.1
	}
	if cfg.Reputation.Alpha <= 0 {
		cfg.Reputation.Alpha = 0.8
	}
	if cfg.Reputation.Gravity <= 0 {
		cfg.Reputation.Gravity = cfg.Feed.Gravity
	}
	if cfg.Reputation.TimeBaseHours <= 0 {
		cfg.Reputation.TimeBaseHours = 4
	}
	if cfg.Reputation.Scale <= 0 {
		cfg.Reputation.Scale = 40
	}
	if cfg.Reputation.LookbackDays <= 0 {
		cfg.Reputation.LookbackDays = 90
	}
	if cfg.Reputation.CandidateLimit <= 0 {
		cfg.Reputation.CandidateLimit = 5000
	}
	if cfg.Cloudinary.Folder == "" {
		cfg.Cloudinary.Folder = "yetbota/posts"
	}
	if cfg.Media.MaxImages <= 0 {
		cfg.Media.MaxImages = 10
	}
	if cfg.Media.MaxAVDurationSec <= 0 {
		cfg.Media.MaxAVDurationSec = 10
	}
	if cfg.Media.MaxUploadBytes <= 0 {
		cfg.Media.MaxUploadBytes = 10 * 1024 * 1024 // Cloudinary Free plan
	}
	if cfg.Media.DefaultPageSize <= 0 {
		cfg.Media.DefaultPageSize = 20
	}
	if cfg.Media.MaxPageSize <= 0 {
		cfg.Media.MaxPageSize = 50
	}
	if cfg.Share.PublicBaseURL == "" {
		cfg.Share.PublicBaseURL = "http://localhost:3000"
	}
	if cfg.Postgres.URL == "" {
		return Config{}, fmt.Errorf("postgres.url is required")
	}
	return cfg, nil
}
