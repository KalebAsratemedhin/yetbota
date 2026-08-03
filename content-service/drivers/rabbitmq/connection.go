package rabbitmq

import (
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/beka-birhanu/yetbota/content-service/internal/messaging"
)

type Config struct {
	URL string
}

type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.Mutex
}

func NewClient(cfg *Config) (*Client, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}
	if err := ch.Qos(10, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq qos: %w", err)
	}
	client := &Client{conn: conn, channel: ch}
	if err := client.declareTopology(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}

func (c *Client) declareTopology() error {
	if err := c.channel.ExchangeDeclare(
		messaging.ExchangeName, "direct", true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}
	if err := c.channel.ExchangeDeclare(
		messaging.DeadLetterExchange, "direct", true, false, false, false, nil,
	); err != nil {
		return fmt.Errorf("declare dlx: %w", err)
	}

	for _, queue := range messaging.WorkQueues {
		dlq := queue + ".dlq"
		if err := c.declareQueue(dlq, nil); err != nil {
			return err
		}
		if err := c.declareQueue(queue, amqp.Table{
			"x-dead-letter-exchange":    messaging.DeadLetterExchange,
			"x-dead-letter-routing-key": dlq,
		}); err != nil {
			return err
		}
		if err := c.channel.QueueBind(queue, queue, messaging.ExchangeName, false, nil); err != nil {
			return fmt.Errorf("bind %s: %w", queue, err)
		}
		if err := c.channel.QueueBind(dlq, dlq, messaging.DeadLetterExchange, false, nil); err != nil {
			return fmt.Errorf("bind dlq %s: %w", dlq, err)
		}
	}
	return nil
}

func (c *Client) declareQueue(name string, args amqp.Table) error {
	_, err := c.channel.QueueDeclare(name, true, false, false, false, args)
	if err != nil {
		return fmt.Errorf("declare queue %s: %w", name, err)
	}
	return nil
}

func (c *Client) Channel() *amqp.Channel {
	return c.channel
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.channel != nil {
		_ = c.channel.Close()
		c.channel = nil
	}
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}
