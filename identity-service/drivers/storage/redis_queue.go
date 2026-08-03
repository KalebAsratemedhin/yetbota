package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	toddlerError "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	domain "github.com/beka-birhanu/yetbota/identity-service/internal/domain/storage"
	"github.com/go-redis/redis/v8"
)

const (
	maxRead   = 10
	blockTime = 500 * time.Millisecond
)

type queue[T any] struct {
	rdb *redis.Client
}

// NewQueue initializes a new redis queue.
func NewQueue[T any](rdb *redis.Client) domain.Queue[T] {
	return &queue[T]{
		rdb: rdb,
	}
}

// Dequeue implements storage.Queue.
func (n *queue[T]) Dequeue(ctx context.Context, topic string) ([]T, error) {
	topicKey := n.cookKey(topic)

	streams, err := n.rdb.XRead(ctx, &redis.XReadArgs{
		Streams: []string{topicKey, "0"},
		Count:   maxRead,
		Block:   blockTime,
	}).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, &toddlerError.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("Redis XREAD failed for stream %s: %v", topicKey, err),
		}
	}

	var resp []T

	for _, stream := range streams {
		for _, message := range stream.Messages {

			var data T
			err = json.Unmarshal([]byte(message.Values["data"].(string)), &data)
			if err != nil { // I hope this never happens. JK 😂, we know this will never happen!
				continue
			}

			resp = append(resp, data)

			// Delete message after processing
			_, _ = n.rdb.XDel(ctx, topicKey, message.ID).Result()
		}
	}

	return resp, nil
}

// Enqueue implements storage.Queue.
func (n *queue[T]) Enqueue(ctx context.Context, topics []string, data T) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return &toddlerError.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("Failed to marshal data %v: %v", data, err),
		}
	}

	for _, topic := range topics {
		topicKey := n.cookKey(topic)

		_, err := n.rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: topicKey,
			Values: map[string]any{
				"data": string(dataJSON),
			},
		}).Result()
		if err != nil {
			return &toddlerError.Error{
				PublicStatusCode:  status.ServerError,
				ServiceStatusCode: status.ServerError,
				PublicMessage:     "Something went wrong",
				ServiceMessage:    fmt.Sprintf("Failed to enqueue for topic %s: %v", topicKey, err),
			}
		}
	}

	return nil
}

func (n *queue[T]) cookKey(topic string) string {
	var zero T
	return fmt.Sprintf(
		"topic@%s:%s",
		reflect.TypeOf(zero),
		topic,
	)
}
