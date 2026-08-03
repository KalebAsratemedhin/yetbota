package storage

import "context"

type StreamEntry struct {
	ID     string
	Fields map[string]string
}

type Stream interface {
	// Add appends a new entry to the stream.
	Add(ctx context.Context, fields map[string]any) error
	// ReadGroup reads up to count messages. id: ">" for new messages, "0" for pending (delivered but unacked).
	ReadGroup(ctx context.Context, group, consumer, id string, count int64) ([]StreamEntry, error)
	// Ack acknowledges processed entries so they leave the pending-entry list.
	Ack(ctx context.Context, group string, ids []string) error
	// EnsureGroup creates the consumer group if it does not exist.
	EnsureGroup(ctx context.Context, group string) error
}
