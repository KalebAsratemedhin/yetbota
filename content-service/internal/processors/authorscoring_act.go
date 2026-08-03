package processors

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/beka-birhanu/yetbota/content-service/drivers/config"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainNotification "github.com/beka-birhanu/yetbota/content-service/internal/domain/notification"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	domainStorage "github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
	domainUser "github.com/beka-birhanu/yetbota/content-service/internal/domain/user"
)

type ScoringEntry struct {
	StreamID  string
	PostID    string
	CreatedAt time.Time
}

type authorScoringActivity struct {
	stream             domainStorage.Stream
	postRepo           domainPost.Repository
	userRepo           domainUser.Repository
	notificationSender domainNotification.Sender
	badges             []config.Badge
	consumerGroup      string
	batchSize          int64
}

type authorScoringActConfig struct {
	Stream             domainStorage.Stream      `validate:"required"`
	PostRepo           domainPost.Repository     `validate:"required"`
	UserRepo           domainUser.Repository     `validate:"required"`
	NotificationSender domainNotification.Sender `validate:"required"`
	Badges             []config.Badge
	ConsumerGroup      string `validate:"required"`
	BatchSize          int64  `validate:"required"`
}

func (c *authorScoringActConfig) validate() error {
	return validator.Validate.Struct(c)
}

func newAuthorScoringAct(cfg *authorScoringActConfig) (*authorScoringActivity, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &authorScoringActivity{
		stream:             cfg.Stream,
		postRepo:           cfg.PostRepo,
		userRepo:           cfg.UserRepo,
		notificationSender: cfg.NotificationSender,
		badges:             cfg.Badges,
		consumerGroup:      cfg.ConsumerGroup,
		batchSize:          cfg.BatchSize,
	}, nil
}

func (a *authorScoringActivity) EnsureGroup(ctx context.Context) error {
	return a.stream.EnsureGroup(ctx, a.consumerGroup)
}

func (a *authorScoringActivity) ReadScoringBatch(ctx context.Context) ([]ScoringEntry, error) {
	consumer := "worker"

	pending, err := a.stream.ReadGroup(ctx, a.consumerGroup, consumer, "0", a.batchSize)
	if err != nil {
		return nil, err
	}

	remaining := a.batchSize - int64(len(pending))
	var fresh []domainStorage.StreamEntry
	if remaining > 0 {
		fresh, err = a.stream.ReadGroup(ctx, a.consumerGroup, consumer, ">", remaining)
		if err != nil {
			return nil, err
		}
	}

	all := append(pending, fresh...)
	entries := make([]ScoringEntry, 0, len(all))
	for _, e := range all {
		postID := e.Fields["postID"]
		if postID == "" {
			continue
		}
		ts, err := strconv.ParseInt(e.Fields["createdAt"], 10, 64)
		if err != nil {
			continue
		}
		entries = append(entries, ScoringEntry{
			StreamID:  e.ID,
			PostID:    postID,
			CreatedAt: time.Unix(ts, 0),
		})
	}
	return entries, nil
}

func (a *authorScoringActivity) ProcessScoringEntry(ctx context.Context, postID string) error {
	post, err := a.postRepo.Read(ctx, postID)
	if err != nil {
		return err
	}

	user, err := a.userRepo.Read(ctx, post.UserID)
	if err != nil {
		return err
	}

	postCount, err := a.postRepo.Count(ctx, &domainPost.ListOptions{UserID: post.UserID})
	if err != nil {
		return err
	}

	user.Score = computeNewScore(user.Score, post, postCount)

	prevBadges := make(map[string]bool, len(user.Badges))
	for _, b := range user.Badges {
		prevBadges[b] = true
	}

	applyEarnedBadges(user, a.badges)

	if err := a.userRepo.Update(ctx, user); err != nil {
		return err
	}

	newBadges := make([]string, 0)
	for _, b := range user.Badges {
		if !prevBadges[b] {
			newBadges = append(newBadges, b)
		}
	}

	if len(newBadges) > 0 {
		tokens, err := a.userRepo.GetDeviceTokens(ctx, user.ID)
		if err != nil {
			log.Printf("error fetching device tokens for user %s: %v", user.ID, err)
		}
		for _, b := range newBadges {
			if len(tokens) == 0 {
				break
			}
			if err := a.notificationSender.Send(ctx, &domainNotification.NotificationData{
				Title:     "New Badge Earned!",
				Body:      fmt.Sprintf("You earned the %s badge. Keep it up!", b),
				Receivers: tokens,
			}); err != nil {
				log.Printf("error sending badge notification: %v", err)
			}
		}
	}

	return nil
}

func (a *authorScoringActivity) AckEntries(ctx context.Context, ids []string) error {
	return a.stream.Ack(ctx, a.consumerGroup, ids)
}
