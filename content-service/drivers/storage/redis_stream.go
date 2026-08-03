package storage

import (
	"context"
	"fmt"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainStorage "github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
	"github.com/redis/go-redis/v9"
)

type redisStream struct {
	rdb *redis.Client
	key string
}

type StreamConfig struct {
	RDB *redis.Client `validate:"required"`
	Key string        `validate:"required"`
}

func (c *StreamConfig) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewStream(c *StreamConfig) (domainStorage.Stream, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &redisStream{rdb: c.RDB, key: c.Key}, nil
}

func (s *redisStream) Add(ctx context.Context, fields map[string]any) error {
	if err := s.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: s.key,
		Values: fields,
	}).Err(); err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("redis stream: add failed: %v", err),
		}
	}
	return nil
}

func (s *redisStream) ReadGroup(ctx context.Context, group, consumer, id string, count int64) ([]domainStorage.StreamEntry, error) {
	msgs, err := s.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{s.key, id},
		Count:    count,
		Block:    3 * time.Second,
		NoAck:    false,
	}).Result()
	if err == redis.Nil || err == context.DeadlineExceeded {
		return nil, nil
	}
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("redis stream: read group failed: %v", err),
		}
	}

	var entries []domainStorage.StreamEntry
	for _, stream := range msgs {
		for _, msg := range stream.Messages {
			fields := make(map[string]string, len(msg.Values))
			for k, v := range msg.Values {
				if str, ok := v.(string); ok {
					fields[k] = str
				}
			}
			entries = append(entries, domainStorage.StreamEntry{ID: msg.ID, Fields: fields})
		}
	}
	return entries, nil
}

func (s *redisStream) Ack(ctx context.Context, group string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	if err := s.rdb.XAck(ctx, s.key, group, ids...).Err(); err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("redis stream: ack failed: %v", err),
		}
	}
	return nil
}

func (s *redisStream) EnsureGroup(ctx context.Context, group string) error {
	err := s.rdb.XGroupCreateMkStream(ctx, s.key, group, "$").Err()
	if err != nil && err.Error() == "BUSYGROUP Consumer Group name already exists" {
		return nil
	}
	if err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("redis stream: ensure group failed: %v", err),
		}
	}
	return nil
}
