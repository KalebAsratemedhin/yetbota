package redis

import (
	"context"
	"time"

	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/go-redis/redis/v8"
)

type Config struct {
	Address  string `yaml:"address" mapstructure:"address" validate:"required"`
	Password string `yaml:"password" mapstructure:"password"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewConnection(ctx context.Context, c *Config) (*redis.Client, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Address,
		Password:     c.Password,
		PoolTimeout:  2 * time.Second,
		IdleTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return rdb, nil
}
