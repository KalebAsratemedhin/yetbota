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
	ExchangeEvents = "yetbota.events"

	RoutingPostCreated     = "post.created"
	RoutingPostUpdated     = "post.updated"
	RoutingCommentCreated  = "comment.created"
	RoutingQuestionCreated = "question.created"
	RoutingQuestionUpdated = "question.updated"
	RoutingAnswerCreated   = "answer.created"
	RoutingAnswerUpdated   = "answer.updated"
)

type contentEvent struct {
	Event      string  `json:"event"`
	OccurredAt string  `json:"occurred_at"`
	PostID     *string `json:"post_id,omitempty"`
	CommentID  *string `json:"comment_id,omitempty"`
	QuestionID *string `json:"question_id,omitempty"`
	AnswerID   *string `json:"answer_id,omitempty"`
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
	id := postID.String()
	return p.publish(ctx, RoutingPostCreated, contentEvent{
		Event: RoutingPostCreated, PostID: &id,
	})
}

func (p *Publisher) PublishPostUpdated(ctx context.Context, postID uuid.UUID) error {
	id := postID.String()
	return p.publish(ctx, RoutingPostUpdated, contentEvent{
		Event: RoutingPostUpdated, PostID: &id,
	})
}

func (p *Publisher) PublishCommentCreated(ctx context.Context, commentID, postID uuid.UUID) error {
	cid, pid := commentID.String(), postID.String()
	return p.publish(ctx, RoutingCommentCreated, contentEvent{
		Event: RoutingCommentCreated, CommentID: &cid, PostID: &pid,
	})
}

func (p *Publisher) PublishQuestionCreated(ctx context.Context, questionID uuid.UUID, postID *uuid.UUID) error {
	qid := questionID.String()
	ev := contentEvent{Event: RoutingQuestionCreated, QuestionID: &qid}
	if postID != nil {
		pid := postID.String()
		ev.PostID = &pid
	}
	return p.publish(ctx, RoutingQuestionCreated, ev)
}

func (p *Publisher) PublishQuestionUpdated(ctx context.Context, questionID uuid.UUID, postID *uuid.UUID) error {
	qid := questionID.String()
	ev := contentEvent{Event: RoutingQuestionUpdated, QuestionID: &qid}
	if postID != nil {
		pid := postID.String()
		ev.PostID = &pid
	}
	return p.publish(ctx, RoutingQuestionUpdated, ev)
}

func (p *Publisher) PublishAnswerCreated(ctx context.Context, answerID, questionID uuid.UUID) error {
	aid, qid := answerID.String(), questionID.String()
	return p.publish(ctx, RoutingAnswerCreated, contentEvent{
		Event: RoutingAnswerCreated, AnswerID: &aid, QuestionID: &qid,
	})
}

func (p *Publisher) PublishAnswerUpdated(ctx context.Context, answerID, questionID uuid.UUID) error {
	aid, qid := answerID.String(), questionID.String()
	return p.publish(ctx, RoutingAnswerUpdated, contentEvent{
		Event: RoutingAnswerUpdated, AnswerID: &aid, QuestionID: &qid,
	})
}

func (p *Publisher) publish(ctx context.Context, routingKey string, ev contentEvent) error {
	if p == nil || p.ch == nil {
		return nil
	}
	ev.OccurredAt = time.Now().UTC().Format(time.RFC3339Nano)
	body, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	return p.ch.PublishWithContext(ctx, ExchangeEvents, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
		Body:         body,
	})
}
