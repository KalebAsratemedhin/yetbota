package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/beka-birhanu/yetbota/content-service/drivers/rabbitmq"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	processorsDomain "github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
	"github.com/beka-birhanu/yetbota/content-service/internal/messaging"
	"golang.org/x/sync/errgroup"
)

type FeedUpdateHandler struct {
	feedUpdate *feedUpdateActivity
	consumer   *rabbitmq.Consumer
}

type FeedUpdateHandlerConfig struct {
	FeedUpdateActivity *feedUpdateActivity `validate:"required"`
	Consumer           *rabbitmq.Consumer  `validate:"required"`
}

func NewFeedUpdateHandler(cfg *FeedUpdateHandlerConfig) (*FeedUpdateHandler, error) {
	if err := validator.Validate.Struct(cfg); err != nil {
		return nil, err
	}
	return &FeedUpdateHandler{
		feedUpdate: cfg.FeedUpdateActivity,
		consumer:   cfg.Consumer,
	}, nil
}

func (h *FeedUpdateHandler) Run(ctx context.Context) error {
	return h.consumer.Consume(ctx, messaging.QueueFeedUpdate, h.handle)
}

func (h *FeedUpdateHandler) handle(ctx context.Context, body []byte, _ int) error {
	var msg messaging.FeedUpdateMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("decode feed update message: %w", err)
	}
	return h.Handle(ctx, processorsDomain.FeedUpdateWorkflowInput{
		PostID:       msg.PostID,
		InteractorID: msg.InteractorID,
	})
}

func (h *FeedUpdateHandler) Handle(ctx context.Context, input processorsDomain.FeedUpdateWorkflowInput) error {
	fanOutData, err := h.feedUpdate.FetchPostFanOutData(ctx, input.PostID)
	if err != nil {
		return err
	}
	if err := h.feedUpdate.CachePostScore(ctx, input.PostID, fanOutData.Score); err != nil {
		log.Printf("CachePostScore failed: %v", err)
	}

	g, gCtx := errgroup.WithContext(ctx)
	var followerResult *FollowerFanOutResult
	var simResult *SimFanOutResult

	g.Go(func() error {
		var err error
		followerResult, err = h.feedUpdate.GetFollowerTree(gCtx, fanOutData.AuthorID, fanOutData.Score)
		return err
	})
	g.Go(func() error {
		var err error
		simResult, err = h.feedUpdate.GetSimilarPostsTree(gCtx, input.PostID, fanOutData.Score)
		return err
	})
	if err := g.Wait(); err != nil {
		return err
	}

	if followerResult.IsCelebrity {
		if err := h.feedUpdate.PublishToCelebrityFeed(ctx, fanOutData.AuthorID, input.PostID, fanOutData.Score); err != nil {
			log.Printf("PublishToCelebrityFeed failed: %v", err)
		}
	}

	allBatchKeys := append(followerResult.BatchKeys, simResult.BatchUserKeys...)
	fanOutG, fanOutCtx := errgroup.WithContext(ctx)
	for _, key := range allBatchKeys {
		batchKey := key
		fanOutG.Go(func() error {
			if err := h.feedUpdate.FanOutFeedBatch(fanOutCtx, input.PostID, batchKey, fanOutData.Score); err != nil {
				log.Printf("FanOutFeedBatch failed: %v", err)
			}
			return nil
		})
	}
	_ = fanOutG.Wait()

	if len(simResult.SimPosts) > 0 {
		simG, simCtx := errgroup.WithContext(ctx)
		for _, key := range followerResult.BatchKeys {
			depthKey := key
			simG.Go(func() error {
				if err := h.feedUpdate.FanOutSimilarPostsToFollowerBatch(simCtx, simResult.SimPosts, depthKey); err != nil {
					log.Printf("FanOutSimilarPostsToFollowerBatch failed: %v", err)
				}
				return nil
			})
		}
		for _, key := range simResult.BatchUserKeys {
			userKey := key
			simG.Go(func() error {
				if err := h.feedUpdate.FanOutSimilarPostsToUserBatch(simCtx, simResult.SimPosts, userKey); err != nil {
					log.Printf("FanOutSimilarPostsToUserBatch failed: %v", err)
				}
				return nil
			})
		}
		simG.Go(func() error {
			if err := h.feedUpdate.FanOutSimilarPostsToAuthor(simCtx, fanOutData.AuthorID, simResult.SimPosts); err != nil {
				log.Printf("FanOutSimilarPostsToAuthor failed: %v", err)
			}
			return nil
		})
		_ = simG.Wait()
	}

	return nil
}
