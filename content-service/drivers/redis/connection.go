package redis

import (
	"context"
	"crypto/tls"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	goredis "github.com/redis/go-redis/v9"
)

type Config struct {
	Address  string `validate:"required"`
	Password string `validate:"required"`

	DB           int
	PoolSize     int `validate:"required"`
	MinIdleConns int
	MaxIdleConns int `validate:"required"`
	MaxRetries   int `validate:"required"`

	DialTimeout  time.Duration `validate:"required"`
	ReadTimeout  time.Duration `validate:"required"`
	WriteTimeout time.Duration `validate:"required"`
	PoolTimeout  time.Duration `validate:"required"`

	ConnMaxIdleTime time.Duration `validate:"required"`
	ConnMaxLifetime time.Duration `validate:"required"`

	TLS bool
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}

	return nil
}

func NewConnection(ctx context.Context, c *Config) (*goredis.Client, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	opts := &goredis.Options{
		Addr:         c.Address,
		Password:     c.Password,
		DB:           c.DB,
		PoolSize:     c.PoolSize,
		MinIdleConns: c.MinIdleConns,
		MaxIdleConns: c.MaxIdleConns,
		MaxRetries:   c.MaxRetries,

		DialTimeout:  c.DialTimeout,
		ReadTimeout:  c.ReadTimeout,
		WriteTimeout: c.WriteTimeout,
		PoolTimeout:  c.PoolTimeout,

		ConnMaxIdleTime: c.ConnMaxIdleTime,
		ConnMaxLifetime: c.ConnMaxLifetime,
	}

	if c.TLS {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	rdb := goredis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return nil, err
	}

	return rdb, nil
}
