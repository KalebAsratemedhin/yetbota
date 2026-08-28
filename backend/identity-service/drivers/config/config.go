package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTP           HTTPConfig     `yaml:"http"`
	Postgres       PostgresConfig `yaml:"postgres"`
	JWT            JWTConfig      `yaml:"jwt"`
	MigrateOnStart bool           `yaml:"migrate_on_start"`
}

type HTTPConfig struct {
	Addr string `yaml:"addr"`
}

type PostgresConfig struct {
	URL string `yaml:"url"`
}

type JWTConfig struct {
	Secret   string `yaml:"secret"`
	TTLHours int    `yaml:"ttl_hours"`
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
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
	if cfg.HTTP.Addr == "" {
		cfg.HTTP.Addr = ":8081"
	}
	if cfg.JWT.TTLHours <= 0 {
		cfg.JWT.TTLHours = 24
	}
	if cfg.JWT.Secret == "" {
		return Config{}, fmt.Errorf("jwt.secret is required")
	}
	if cfg.Postgres.URL == "" {
		return Config{}, fmt.Errorf("postgres.url is required")
	}
	return cfg, nil
}
