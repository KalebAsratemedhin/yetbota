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
	Identity       IdentityConfig   `yaml:"identity"`
	Cloudinary     CloudinaryConfig `yaml:"cloudinary"`
	Media          MediaConfig      `yaml:"media"`
	Share          ShareConfig      `yaml:"share"`
	CORS           CORSConfig       `yaml:"cors"`
	MigrateOnStart bool             `yaml:"migrate_on_start"`
}

type HTTPConfig struct {
	Addr string `yaml:"addr"`
}

type PostgresConfig struct {
	URL string `yaml:"url"`
}

type IdentityConfig struct {
	BaseURL string `yaml:"base_url"`
}

type CloudinaryConfig struct {
	CloudName string `yaml:"cloud_name"`
	APIKey    string `yaml:"api_key"`
	APISecret string `yaml:"api_secret"`
	Folder    string `yaml:"folder"`
}

type MediaConfig struct {
	MaxImages        int `yaml:"max_images"`
	MaxAVDurationSec int `yaml:"max_av_duration_sec"`
	DefaultPageSize  int `yaml:"default_page_size"`
	MaxPageSize      int `yaml:"max_page_size"`
}

type ShareConfig struct {
	PublicBaseURL string `yaml:"public_base_url"`
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
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
	if v := os.Getenv("MIGRATE_ON_START"); v != "" {
		cfg.MigrateOnStart = v == "1" || strings.EqualFold(v, "true")
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
	if cfg.Cloudinary.Folder == "" {
		cfg.Cloudinary.Folder = "yetbota/posts"
	}
	if cfg.Media.MaxImages <= 0 {
		cfg.Media.MaxImages = 10
	}
	if cfg.Media.MaxAVDurationSec <= 0 {
		cfg.Media.MaxAVDurationSec = 10
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
