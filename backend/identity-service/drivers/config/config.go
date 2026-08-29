package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTP           HTTPConfig     `yaml:"http"`
	Postgres       PostgresConfig `yaml:"postgres"`
	JWT            JWTConfig      `yaml:"jwt"`
	CORS           CORSConfig     `yaml:"cors"`
	MigrateOnStart bool           `yaml:"migrate_on_start"`
}

type HTTPConfig struct {
	Addr string `yaml:"addr"`
}

type PostgresConfig struct {
	URL string `yaml:"url"`
}

type JWTConfig struct {
	Secret           string `yaml:"secret"`
	AccessTTLMinutes int    `yaml:"access_ttl_minutes"`
	RefreshTTLDays   int    `yaml:"refresh_ttl_days"`
	// Deprecated: use access_ttl_minutes. Kept so old config still loads.
	TTLHours int `yaml:"ttl_hours"`
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
	// Render (and most PaaS) set PORT; prefer it over config defaults.
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
	if cfg.JWT.Secret == "" {
		return Config{}, fmt.Errorf("jwt.secret is required")
	}
	if cfg.Postgres.URL == "" {
		return Config{}, fmt.Errorf("postgres.url is required")
	}
	return cfg, nil
}
