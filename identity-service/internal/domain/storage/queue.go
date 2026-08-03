package storage

import "context"

type Queue[T any] interface {
	Enqueue(ctx context.Context, topics []string, data T) error
	Dequeue(ctx context.Context, topic string) ([]T, error)
}
