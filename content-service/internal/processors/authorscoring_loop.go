package processors

import (
	"context"
	"log"
	"time"

	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	processorsDomain "github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
)

type AuthorScoringLoop struct {
	activity *authorScoringActivity
	input    processorsDomain.AuthorScoringWorkflowInput
}

type AuthorScoringLoopConfig struct {
	Activity *authorScoringActivity                    `validate:"required"`
	Input    processorsDomain.AuthorScoringWorkflowInput `validate:"required"`
}

func NewAuthorScoringLoop(cfg *AuthorScoringLoopConfig) (*AuthorScoringLoop, error) {
	if err := validator.Validate.Struct(cfg); err != nil {
		return nil, err
	}
	return &AuthorScoringLoop{
		activity: cfg.Activity,
		input:    cfg.Input,
	}, nil
}

func (l *AuthorScoringLoop) Run(ctx context.Context) error {
	if err := l.activity.EnsureGroup(ctx); err != nil {
		log.Printf("author scoring EnsureGroup failed: %v", err)
	}

	window := time.Duration(l.input.StabilizingWindowHours) * time.Hour
	poll := time.Duration(l.input.PollIntervalSec) * time.Second

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		entries, err := l.activity.ReadScoringBatch(ctx)
		if err != nil {
			log.Printf("ReadScoringBatch failed: %v", err)
		} else {
			now := time.Now()
			readyIDs := make([]string, 0, len(entries))
			for _, e := range entries {
				if now.Sub(e.CreatedAt) < window {
					continue
				}
				if err := l.activity.ProcessScoringEntry(ctx, e.PostID); err != nil {
					log.Printf("ProcessScoringEntry failed postID=%s: %v", e.PostID, err)
					continue
				}
				readyIDs = append(readyIDs, e.StreamID)
			}
			if len(readyIDs) > 0 {
				if err := l.activity.AckEntries(ctx, readyIDs); err != nil {
					log.Printf("AckEntries failed: %v", err)
				}
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(poll):
		}
	}
}
