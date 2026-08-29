package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTP           HTTPConfig        `yaml:"http"`
	Postgres       PostgresConfig    `yaml:"postgres"`
	Surreal        SurrealConfig     `yaml:"surreal"`
	Cloudinary     CloudinaryConfig  `yaml:"cloudinary"`
	JWT            JWTConfig         `yaml:"jwt"`
	CORS           CORSConfig        `yaml:"cors"`
	MigrateOnStart bool              `yaml:"migrate_on_start"`
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

type CloudinaryConfig struct {
	CloudName string `yaml:"cloud_name"`
	APIKey    string `yaml:"api_key"`
	APISecret string `yaml:"api_secret"`
	Folder    string `yaml:"folder"`
}

type JWTConfig struct {
	Secret           string `yaml:"secret"`
	AccessTTLMinutes int    `yaml:"access_ttl_minutes"`
	RefreshTTLDays   int    `yaml:"refresh_ttl_days"`
	TTLHours         int    `yaml:"ttl_hours"`
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
		if v[0] == ':' {
			cfg.HTTP.Addr = v
		} else {
			cfg.HTTP.Addr = ":" + v
		}
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if v := os.Getenv("MIGRATE_ON_START"); v != "" {
		cfg.MigrateOnStart = v == "1" || v == "true" || v == "TRUE"
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
	if cfg.HTTP.Addr == "" {
		cfg.HTTP.Addr = ":8081"
	}
	if cfg.JWT.AccessTTLMinutes <= 0 {
		if cfg.JWT.TTLHours > 0 {
			cfg.JWT.AccessTTLMinutes = cfg.JWT.TTLHours * 60
		} else {
			cfg.JWT.AccessTTLMinutes = 15
		}
	}
	if cfg.JWT.RefreshTTLDays <= 0 {
		cfg.JWT.RefreshTTLDays = 30
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
	if cfg.Cloudinary.Folder == "" {
		cfg.Cloudinary.Folder = "yetbota/avatars"
	}
	if cfg.JWT.Secret == "" {
		return Config{}, fmt.Errorf("jwt.secret is required")
	}
	if cfg.Postgres.URL == "" {
		return Config{}, fmt.Errorf("postgres.url is required")
	}
	return cfg, nil
}
