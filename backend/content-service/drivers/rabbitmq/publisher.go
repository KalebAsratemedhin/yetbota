package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeEvents   = "yetbota.events"
	RoutingPostCreated = "post.created"
	QueuePostIngest  = "ai.post.ingest"
)

// PostCreatedEvent is published by content-service after a post is created.
type PostCreatedEvent struct {
	Event      string `json:"event"`
	PostID     string `json:"post_id"`
	OccurredAt string `json:"occurred_at"`
}

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func ConnectPublisher(url string) (*Publisher, error) {
	if url == "" {
		return nil, fmt.Errorf("rabbitmq url is empty")
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}
	if err := ch.ExchangeDeclare(ExchangeEvents, "topic", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}
	return &Publisher{conn: conn, ch: ch}, nil
}

func (p *Publisher) Close() error {
	if p == nil {
		return nil
	}
	var err error
	if p.ch != nil {
		err = p.ch.Close()
	}
	if p.conn != nil {
		if e := p.conn.Close(); e != nil && err == nil {
			err = e
		}
	}
	return err
}

func (p *Publisher) PublishPostCreated(ctx context.Context, postID uuid.UUID) error {
	if p == nil || p.ch == nil {
		return nil
	}
	body, err := json.Marshal(PostCreatedEvent{
		Event:      RoutingPostCreated,
		PostID:     postID.String(),
		OccurredAt: time.Now().UTC().Format(time.RFC3339Nano),
	})
	if err != nil {
		return err
	}
	return p.ch.PublishWithContext(ctx, ExchangeEvents, RoutingPostCreated, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
		Body:         body,
	})
}
