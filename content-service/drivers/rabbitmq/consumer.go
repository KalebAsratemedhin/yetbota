package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/beka-birhanu/yetbota/content-service/internal/messaging"
)

type Consumer struct {
	client *Client
}

func NewConsumer(client *Client) *Consumer {
	return &Consumer{client: client}
}

type Handler func(ctx context.Context, body []byte, deliveryCount int) error

func (c *Consumer) Consume(ctx context.Context, queue string, handler Handler) error {
	deliveries, err := c.client.Channel().Consume(
		queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("consume %s: %w", queue, err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("delivery channel closed for %s", queue)
			}
			count := deliveryCount(delivery)
			err := handler(ctx, delivery.Body, count)
			if err != nil {
				if count >= 3 {
					_ = delivery.Nack(false, false)
				} else {
					_ = delivery.Nack(false, true)
				}
				continue
			}
			_ = delivery.Ack(false)
		}
	}
}

func deliveryCount(d amqp.Delivery) int {
	if deaths, ok := d.Headers["x-death"].([]any); ok && len(deaths) > 0 {
		if death, ok := deaths[0].(amqp.Table); ok {
			if count, ok := death["count"].(int64); ok && count > 0 {
				return int(count) + 1
			}
		}
	}
	if d.Redelivered {
		return 2
	}
	return 1
}

func (c *Consumer) Request(ctx context.Context, queue string, payload any, timeout time.Duration) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	replyQueue, err := c.client.Channel().QueueDeclare(
		"",
		false,
		true,
		true,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("declare reply queue: %w", err)
	}

	msgs, err := c.client.Channel().Consume(
		replyQueue.Name,
		"",
		true,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("consume reply queue: %w", err)
	}

	corrID := uuid.NewString()
	if err := c.client.Channel().PublishWithContext(
		ctx,
		messaging.ExchangeName,
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			DeliveryMode:  amqp.Persistent,
			CorrelationId: corrID,
			ReplyTo:       replyQueue.Name,
			Body:          body,
		},
	); err != nil {
		return nil, fmt.Errorf("publish request: %w", err)
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
			return nil, fmt.Errorf("rpc timeout after %s", timeout)
		case msg, ok := <-msgs:
			if !ok {
				return nil, fmt.Errorf("reply channel closed")
			}
			if msg.CorrelationId == corrID {
				return msg.Body, nil
			}
		}
	}
}
