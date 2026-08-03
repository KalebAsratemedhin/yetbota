package storage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainStorage "github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
	"github.com/redis/go-redis/v9"
)

type set struct {
	rdb    *redis.Client
	prefix string
}

type Config struct {
	RDB    *redis.Client `validate:"required"`
	Prefix string
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewSet(c *Config) (domainStorage.Set, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	return &set{rdb: c.RDB, prefix: c.Prefix}, nil
}

func (s *set) Add(ctx context.Context, key string, ttl int64) error {
	key = s.cookKey(key)
	if err := s.rdb.Set(ctx, key, 1, time.Duration(ttl)*time.Second).Err(); err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("redis set: add failed: %v", err),
		}
	}
	return nil
}

func (s *set) TryAdd(ctx context.Context, key string, ttl int64) (bool, error) {
	key = s.cookKey(key)
	ok, err := s.rdb.SetNX(ctx, key, 1, time.Duration(ttl)*time.Second).Result()
	if err != nil {
		return false, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("redis set: try add failed: %v", err),
		}
	}
	return ok, nil
}

func (s *set) Delete(ctx context.Context, key string) error {
	s.cookKey(key)
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("redis set: delete failed: %v", err),
		}
	}
	return nil
}

func (s *set) Exists(ctx context.Context, keys []string) (map[string]bool, error) {
	result := make(map[string]bool, len(keys))
	if len(keys) == 0 {
		return result, nil
	}

	pipe := s.rdb.Pipeline()
	cmds := make([]*redis.IntCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Exists(ctx, s.cookKey(key))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("redis set: exists failed: %v", err),
		}
	}

	for i, key := range keys {
		result[key] = cmds[i].Val() > 0
	}
	return result, nil
}

func (s *set) StoreBatch(ctx context.Context, key string, data map[string]float64, ttlSeconds int64) error {
	key = s.cookKey(key)
	args := make([]any, 0, len(data)*2)
	for k, v := range data {
		args = append(args, k, v)
	}
	pipe := s.rdb.Pipeline()
	pipe.HSet(ctx, key, args...)
	pipe.Expire(ctx, key, time.Duration(ttlSeconds)*time.Second)
	if _, err := pipe.Exec(ctx); err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("redis set: store batch failed: %v", err),
		}
	}
	return nil
}

func (s *set) ReadBatch(ctx context.Context, key string) (map[string]float64, error) {
	key = s.cookKey(key)
	raw, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("redis set: read batch failed: %v", err),
		}
	}
	result := make(map[string]float64, len(raw))
	for k, v := range raw {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, &toddlerr.Error{
				PublicStatusCode:  status.ServerError,
				ServiceStatusCode: status.ServerError,
				PublicMessage:     "something went wrong",
				ServiceMessage:    fmt.Sprintf("redis set: parse batch score for key %q: %v", k, err),
			}
		}
		result[k] = f
	}
	return result, nil
}

func (s *set) cookKey(key string) string {
	return fmt.Sprintf("%s:%s", s.prefix, key)
}
