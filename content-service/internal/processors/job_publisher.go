package processors

import (
	"context"
	"fmt"
	"net/http"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/rabbitmq"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainComment "github.com/beka-birhanu/yetbota/content-service/internal/domain/comment"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	processorsDomain "github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
	domainStorage "github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
	"github.com/beka-birhanu/yetbota/content-service/internal/messaging"
)

type JobPublisher struct {
	publisher            *rabbitmq.Publisher
	postRepo             domainPost.Repository
	commentRepo          domainComment.Repository
	summarizeDebounce    domainStorage.Set
	summarizeDebounceSec int64
}

type JobPublisherConfig struct {
	Publisher            *rabbitmq.Publisher      `validate:"required"`
	PostRepo             domainPost.Repository    `validate:"required"`
	CommentRepo          domainComment.Repository `validate:"required"`
	SummarizeDebounce    domainStorage.Set
	SummarizeDebounceSec int64
}

func NewJobPublisher(cfg *JobPublisherConfig) (*JobPublisher, error) {
	if err := validator.Validate.Struct(cfg); err != nil {
		return nil, err
	}
	debounceSec := cfg.SummarizeDebounceSec
	if debounceSec <= 0 {
		debounceSec = 15 * 60
	}
	return &JobPublisher{
		publisher:            cfg.Publisher,
		postRepo:             cfg.PostRepo,
		commentRepo:          cfg.CommentRepo,
		summarizeDebounce:    cfg.SummarizeDebounce,
		summarizeDebounceSec: debounceSec,
	}, nil
}

func (p *JobPublisher) TriggerNewPostWorkflow(ctx context.Context, input processorsDomain.NewPostWorkflowInput) error {
	if err := p.publisher.Publish(ctx, messaging.QueueNewPost, messaging.NewPostMessage{PostID: input.PostID}); err != nil {
		return publishError("new post", err)
	}
	return nil
}

func (p *JobPublisher) TriggerPlaceIngestWorkflow(ctx context.Context, input processorsDomain.PlaceIngestWorkflowInput) error {
	if err := p.publisher.Publish(ctx, messaging.QueuePlaceIngest, messaging.PlaceIngestMessage{PlaceID: input.PlaceID}); err != nil {
		return publishError("place ingest", err)
	}
	return nil
}

func (p *JobPublisher) TriggerContributionIngestWorkflow(ctx context.Context, input processorsDomain.ContributionIngestWorkflowInput) error {
	if err := p.publisher.Publish(ctx, messaging.QueuePlaceIngest, messaging.ContributionIngestMessage{
		ContributionID: input.ContributionID,
	}); err != nil {
		return publishError("contribution ingest", err)
	}
	return nil
}

func (p *JobPublisher) TriggerPlaceSummarizeWorkflow(ctx context.Context, input processorsDomain.PlaceSummarizeWorkflowInput) error {
	if input.PlaceID == "" {
		return nil
	}
	if p.summarizeDebounce != nil {
		ok, err := p.summarizeDebounce.TryAdd(ctx, input.PlaceID, p.summarizeDebounceSec)
		if err != nil {
			fmt.Printf("place summarize debounce failed for %s: %v\n", input.PlaceID, err)
		} else if !ok {
			return nil
		}
	}
	if err := p.publisher.Publish(ctx, messaging.QueuePlaceSummarize, messaging.PlaceSummarizeMessage{PlaceID: input.PlaceID}); err != nil {
		return publishError("place summarize", err)
	}
	return nil
}

func (p *JobPublisher) TriggerFeedUpdateWorkflow(ctx context.Context, input processorsDomain.FeedUpdateWorkflowInput) error {
	if err := p.publisher.Publish(ctx, messaging.QueueFeedUpdate, messaging.FeedUpdateMessage{
		PostID:       input.PostID,
		InteractorID: input.InteractorID,
	}); err != nil {
		return publishError("feed update", err)
	}
	return nil
}

func (p *JobPublisher) TriggerAnswerEmbeddingWorkflow(ctx context.Context, input processorsDomain.AnswerEmbeddingWorkflowInput) error {
	comment, err := p.commentRepo.Read(ctx, input.CommentID)
	if err != nil {
		return err
	}
	question, err := p.postRepo.Read(ctx, comment.PostID)
	if err != nil {
		return err
	}
	msg := buildAnswerIngestMessage(comment, question)
	if err := p.publisher.Publish(ctx, messaging.QueueAIIngest, msg); err != nil {
		return publishError("answer embedding", err)
	}
	return nil
}

func publishError(kind string, err error) error {
	return &toddlerr.Error{
		PublicStatusCode:  http.StatusInternalServerError,
		ServiceStatusCode: http.StatusInternalServerError,
		PublicMessage:     "Something went wrong.",
		ServiceMessage:    fmt.Sprintf("Error publishing %s job: %v", kind, err),
	}
}
