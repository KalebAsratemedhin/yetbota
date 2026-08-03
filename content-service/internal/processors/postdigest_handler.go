package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	rabbitmqDriver "github.com/beka-birhanu/yetbota/content-service/drivers/rabbitmq"
	domainContribution "github.com/beka-birhanu/yetbota/content-service/internal/domain/contribution"
	domainPlace "github.com/beka-birhanu/yetbota/content-service/internal/domain/place"
	domainActivity "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeactivity"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	"github.com/beka-birhanu/yetbota/content-service/internal/messaging"
	"github.com/google/uuid"
)

const defaultDigestLinkRadiusKm = 0.5

type PostDigestHandler struct {
	postRepo         domainPost.Repository
	placeRepo        domainPlace.Repository
	contributionRepo domainContribution.Repository
	activityRepo     domainActivity.Repository
	consumer         *rabbitmqDriver.Consumer
	publisher        *rabbitmqDriver.Publisher
	linkRadiusKm     float64
}

type PostDigestHandlerConfig struct {
	PostRepo         domainPost.Repository         `validate:"required"`
	PlaceRepo        domainPlace.Repository        `validate:"required"`
	ContributionRepo domainContribution.Repository `validate:"required"`
	ActivityRepo     domainActivity.Repository     `validate:"required"`
	Consumer         *rabbitmqDriver.Consumer      `validate:"required"`
	Publisher        *rabbitmqDriver.Publisher
	LinkRadiusKm     float64
}

func NewPostDigestHandler(cfg *PostDigestHandlerConfig) (*PostDigestHandler, error) {
	if err := validator.Validate.Struct(cfg); err != nil {
		return nil, err
	}
	radius := cfg.LinkRadiusKm
	if radius <= 0 {
		radius = defaultDigestLinkRadiusKm
	}
	return &PostDigestHandler{
		postRepo:         cfg.PostRepo,
		placeRepo:        cfg.PlaceRepo,
		contributionRepo: cfg.ContributionRepo,
		activityRepo:     cfg.ActivityRepo,
		consumer:         cfg.Consumer,
		publisher:        cfg.Publisher,
		linkRadiusKm:     radius,
	}, nil
}

func (h *PostDigestHandler) Run(ctx context.Context) error {
	return h.consumer.Consume(ctx, messaging.QueuePostDigest, h.handle)
}

func (h *PostDigestHandler) handle(ctx context.Context, body []byte, _ int) error {
	var msg messaging.PostDigestMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("decode post digest message: %w", err)
	}
	return h.Handle(ctx, msg.PostID)
}

func (h *PostDigestHandler) Handle(ctx context.Context, postID string) error {
	post, err := h.postRepo.Read(ctx, postID)
	if err != nil {
		return err
	}
	if post.ModerationStatus != constants.ModerationStatusVisible {
		return nil
	}

	placeID := ""
	linkedNow := false
	if post.PlaceID.Valid && post.PlaceID.String != "" {
		placeID = post.PlaceID.String
	} else if post.Location.Valid && post.Location.Point != nil {
		lat := post.Location.Point.Y()
		lon := post.Location.Point.X()
		status := constants.PlaceStatusVisible
		places, listErr := h.placeRepo.List(ctx, &domainPlace.ListOptions{
			NearLat:  &lat,
			NearLon:  &lon,
			RadiusKm: &h.linkRadiusKm,
			Status:   &status,
			Page:     1,
			PageSize: 1,
		})
		if listErr != nil {
			return listErr
		}
		if len(places) > 0 {
			placeID = places[0].ID
			post.PlaceID = null.StringFrom(placeID)
			if err := h.postRepo.Update(ctx, nil, post); err != nil {
				return err
			}
			linkedNow = true
		}
	}

	if placeID == "" {
		return nil
	}

	now := time.Now().UTC()
	if linkedNow {
		_ = h.activityRepo.Add(ctx, nil, &domainActivity.Event{
			ID:        uuid.NewString(),
			PlaceID:   placeID,
			ActorID:   null.StringFrom(post.UserID),
			EventType: domainActivity.EventPostLinked,
			RefID:     null.StringFrom(post.ID),
			Summary:   "A post was linked to this place",
			CreatedAt: now,
		})
	}

	exists, err := h.contributionRepo.ExistsForSourcePost(ctx, post.ID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	content := strings.TrimSpace(post.Title)
	if desc := strings.TrimSpace(post.Description); desc != "" {
		if content != "" {
			content += "\n\n"
		}
		content += desc
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	if len(content) > 10000 {
		content = content[:10000]
	}

	entity := &domainContribution.Contribution{
		ID:               uuid.NewString(),
		PlaceID:          placeID,
		UserID:           post.UserID, // D5: post author owns extracted rows
		Kind:             "tip",
		Content:          content,
		ModerationStatus: constants.ModerationStatusVisible,
		Origin:           constants.ContributionOriginPostExtracted,
		SourcePostID:     null.StringFrom(post.ID),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := h.contributionRepo.Add(ctx, nil, entity); err != nil {
		return err
	}
	_ = h.activityRepo.Add(ctx, nil, &domainActivity.Event{
		ID:        uuid.NewString(),
		PlaceID:   placeID,
		ActorID:   null.StringFrom(post.UserID),
		EventType: domainActivity.EventContributionAdded,
		RefID:     null.StringFrom(entity.ID),
		Summary:   "Knowledge extracted from a linked post",
		CreatedAt: now,
	})
	if h.publisher != nil {
		_ = h.publisher.Publish(ctx, messaging.QueuePlaceIngest, messaging.ContributionIngestMessage{
			ContributionID: entity.ID,
		})
	}
	return nil
}
