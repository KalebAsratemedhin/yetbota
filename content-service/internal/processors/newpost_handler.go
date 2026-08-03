package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	rabbitmqDriver "github.com/beka-birhanu/yetbota/content-service/drivers/rabbitmq"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	"github.com/beka-birhanu/yetbota/content-service/internal/messaging"
	"golang.org/x/sync/errgroup"
)

type NewPostHandler struct {
	newPost   *newPostActivity
	postRepo  domainPost.Repository
	consumer  *rabbitmqDriver.Consumer
	publisher *rabbitmqDriver.Publisher
	ingestRPC time.Duration
}

type NewPostHandlerConfig struct {
	NewPostActivity  *newPostActivity          `validate:"required"`
	PostRepo         domainPost.Repository     `validate:"required"`
	Consumer         *rabbitmqDriver.Consumer  `validate:"required"`
	Publisher        *rabbitmqDriver.Publisher `validate:"required"`
	IngestRPCTimeout time.Duration
}

func NewNewPostHandler(cfg *NewPostHandlerConfig) (*NewPostHandler, error) {
	if err := validator.Validate.Struct(cfg); err != nil {
		return nil, err
	}
	timeout := cfg.IngestRPCTimeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	return &NewPostHandler{
		newPost:   cfg.NewPostActivity,
		postRepo:  cfg.PostRepo,
		consumer:  cfg.Consumer,
		publisher: cfg.Publisher,
		ingestRPC: timeout,
	}, nil
}

func (h *NewPostHandler) Run(ctx context.Context) error {
	return h.consumer.Consume(ctx, messaging.QueueNewPost, h.handle)
}

func (h *NewPostHandler) handle(ctx context.Context, body []byte, _ int) error {
	var msg messaging.NewPostMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("decode new post message: %w", err)
	}
	return h.Handle(ctx, msg.PostID)
}

func (h *NewPostHandler) Handle(ctx context.Context, postID string) error {
	photoIDs, err := h.newPost.FetchPostPhotoIDs(ctx, postID)
	if err != nil {
		return err
	}

	if len(photoIDs) > 0 {
		g, gCtx := errgroup.WithContext(ctx)
		for _, id := range photoIDs {
			photoID := id
			g.Go(func() error {
				return h.newPost.ProcessPhoto(gCtx, photoID)
			})
		}
		if err := g.Wait(); err != nil {
			return err
		}
	}

	post, err := h.postRepo.Read(ctx, postID)
	if err != nil {
		return err
	}

	ingestMsg := buildIngestMessage(post)
	replyBody, err := h.consumer.Request(ctx, messaging.QueueAIIngest, ingestMsg, h.ingestRPC)
	if err != nil {
		fmt.Printf("post embedding RPC failed for %s: %v\n", postID, err)
	} else {
		var result messaging.IngestResultMessage
		if err := json.Unmarshal(replyBody, &result); err != nil {
			fmt.Printf("decode ingest result for %s: %v\n", postID, err)
		} else if result.Status == constants.VerdictDuplicate {
			if err := h.newPost.MarkPostDuplicate(ctx, postID); err != nil {
				fmt.Printf("mark post duplicate %s: %v\n", postID, err)
			}
			return nil
		}
	}

	h.publishDigest(ctx, postID)

	return h.publisher.Publish(ctx, messaging.QueueFeedUpdate, messaging.FeedUpdateMessage{
		PostID: postID,
	})
}

func (h *NewPostHandler) publishDigest(ctx context.Context, postID string) {
	if err := h.publisher.Publish(ctx, messaging.QueuePostDigest, messaging.PostDigestMessage{PostID: postID}); err != nil {
		fmt.Printf("unable to publish post digest for %s: %v\n", postID, err)
	}
}

func buildIngestMessage(post *dbmodels.Post) messaging.IngestMessage {
	kind := "post"
	if post.IsQuestion {
		kind = "question"
	}
	text := strings.TrimSpace(post.Title)
	if desc := strings.TrimSpace(post.Description); desc != "" {
		text = strings.TrimSpace(text + "\n\n" + desc)
	}
	tags := make([]string, len(post.Tags))
	copy(tags, post.Tags)

	msg := messaging.IngestMessage{
		ContentID: post.ID,
		Kind:      kind,
		UserID:    post.UserID,
		Text:      text,
		Tags:      tags,
	}
	if post.Location.Valid && post.Location.Point != nil {
		lat := post.Location.Point.Y()
		lon := post.Location.Point.X()
		msg.Latitude = &lat
		msg.Longitude = &lon
	}
	return msg
}

func buildAnswerIngestMessage(comment *dbmodels.Comment, question *dbmodels.Post) messaging.IngestMessage {
	contextText := strings.TrimSpace(question.Title)
	if desc := strings.TrimSpace(question.Description); desc != "" {
		contextText = strings.TrimSpace(contextText + "\n\n" + desc)
	}
	return messaging.IngestMessage{
		ContentID: comment.ID,
		Kind:      "answer",
		UserID:    comment.UserID,
		Text:      comment.Content,
		ParentID:  comment.PostID,
		Context:   contextText,
	}
}
