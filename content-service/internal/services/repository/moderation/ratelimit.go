package moderation

import (
	"context"
	"fmt"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	domainModeration "github.com/beka-birhanu/yetbota/content-service/internal/domain/moderation"
	goredis "github.com/redis/go-redis/v9"
)

type rateLimiter struct {
	rdb    *goredis.Client
	prefix string
}

type RateLimitConfig struct {
	RDB    *goredis.Client `validate:"required"`
	Prefix string
}

func NewRateLimiter(cfg *RateLimitConfig) (domainModeration.RateLimiter, error) {
	prefix := cfg.Prefix
	if prefix == "" {
		prefix = "REPORT_RL"
	}
	return &rateLimiter{rdb: cfg.RDB, prefix: prefix}, nil
}

func (l *rateLimiter) Allow(ctx context.Context, key string, max int, window time.Duration) (bool, error) {
	redisKey := fmt.Sprintf("%s:%s", l.prefix, key)
	count, err := l.rdb.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, toddlerr.FromDBError(err, "rate_limit")
	}
	if count == 1 {
		if err := l.rdb.Expire(ctx, redisKey, window).Err(); err != nil {
			return false, toddlerr.FromDBError(err, "rate_limit")
		}
	}
	return count <= int64(max), nil
}
