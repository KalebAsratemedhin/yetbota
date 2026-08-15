package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/geotypes"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	rabbitmqDriver "github.com/beka-birhanu/yetbota/content-service/drivers/rabbitmq"
	domainContribution "github.com/beka-birhanu/yetbota/content-service/internal/domain/contribution"
	domainPlace "github.com/beka-birhanu/yetbota/content-service/internal/domain/place"
	domainActivity "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeactivity"
	domainPlaceMatch "github.com/beka-birhanu/yetbota/content-service/internal/domain/placematch"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	processorsDomain "github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
	"github.com/beka-birhanu/yetbota/content-service/internal/messaging"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/google/uuid"
	"github.com/twpayne/go-geom"
)

const (
	defaultDigestLinkRadiusKm = 0.5
	defaultPlaceMatchMinScore = 0.75
)

type PostDigestHandler struct {
	postRepo         domainPost.Repository
	placeRepo        domainPlace.Repository
	contributionRepo domainContribution.Repository
	activityRepo     domainActivity.Repository
	placeMatcher     domainPlaceMatch.Client
	consumer         *rabbitmqDriver.Consumer
	publisher        *rabbitmqDriver.Publisher
	executor         processorsDomain.Executor
	linkRadiusKm     float64
	minMatchScore    float64
}

type PostDigestHandlerConfig struct {
	PostRepo         domainPost.Repository         `validate:"required"`
	PlaceRepo        domainPlace.Repository        `validate:"required"`
	ContributionRepo domainContribution.Repository `validate:"required"`
	ActivityRepo     domainActivity.Repository     `validate:"required"`
	Consumer         *rabbitmqDriver.Consumer      `validate:"required"`
	Publisher        *rabbitmqDriver.Publisher
	Executor         processorsDomain.Executor
	PlaceMatcher     domainPlaceMatch.Client
	LinkRadiusKm     float64
	MinMatchScore    float64
}

func NewPostDigestHandler(cfg *PostDigestHandlerConfig) (*PostDigestHandler, error) {
	if err := validator.Validate.Struct(cfg); err != nil {
		return nil, err
	}
	radius := cfg.LinkRadiusKm
	if radius <= 0 {
		radius = defaultDigestLinkRadiusKm
	}
	minScore := cfg.MinMatchScore
	if minScore <= 0 {
		minScore = defaultPlaceMatchMinScore
	}
	return &PostDigestHandler{
		postRepo:         cfg.PostRepo,
		placeRepo:        cfg.PlaceRepo,
		contributionRepo: cfg.ContributionRepo,
		activityRepo:     cfg.ActivityRepo,
		placeMatcher:     cfg.PlaceMatcher,
		consumer:         cfg.Consumer,
		publisher:        cfg.Publisher,
		executor:         cfg.Executor,
		linkRadiusKm:     radius,
		minMatchScore:    minScore,
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
	createdPlace := false
	if post.PlaceID.Valid && post.PlaceID.String != "" {
		placeID = post.PlaceID.String
	} else {
		placeID, createdPlace, err = h.resolvePlace(ctx, post)
		if err != nil {
			return err
		}
		if placeID == "" {
			return nil
		}
		post.PlaceID = null.StringFrom(placeID)
		if err := h.postRepo.Update(ctx, nil, post); err != nil {
			return err
		}
		linkedNow = true
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
		h.publishPostIngest(ctx, post, placeID)
		if createdPlace && h.publisher != nil {
			_ = h.publisher.Publish(ctx, messaging.QueuePlaceIngest, messaging.PlaceIngestMessage{PlaceID: placeID})
		}
		if h.executor != nil {
			_ = h.executor.TriggerPlaceSummarizeWorkflow(ctx, processorsDomain.PlaceSummarizeWorkflowInput{PlaceID: placeID})
		}
	}

	exists, err := h.contributionRepo.ExistsForSourcePost(ctx, post.ID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	content := postText(post)
	if content == "" {
		return nil
	}
	if len(content) > 10000 {
		content = content[:10000]
	}

	entity := &domainContribution.Contribution{
		ID:               uuid.NewString(),
		PlaceID:          placeID,
		UserID:           post.UserID,
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

func (h *PostDigestHandler) resolvePlace(ctx context.Context, post *dbmodels.Post) (placeID string, created bool, err error) {
	text := postText(post)
	var lat, lon *float64
	if post.Location.Valid && post.Location.Point != nil {
		y := post.Location.Point.Y()
		x := post.Location.Point.X()
		lat, lon = &y, &x
	}

	if h.placeMatcher != nil && text != "" {
		radius := h.linkRadiusKm
		match, matchErr := h.placeMatcher.Match(ctx, &domainPlaceMatch.Request{
			Text:            text,
			ExcludeSourceID: post.ID,
			Latitude:        lat,
			Longitude:       lon,
			RadiusKm:        &radius,
		})
		if matchErr == nil && match != nil && match.PlaceID != "" && match.Score >= h.minMatchScore {
			return match.PlaceID, false, nil
		}
	}

	if lat != nil && lon != nil {
		status := constants.PlaceStatusVisible
		places, listErr := h.placeRepo.List(ctx, &domainPlace.ListOptions{
			NearLat:  lat,
			NearLon:  lon,
			RadiusKm: &h.linkRadiusKm,
			Status:   &status,
			Page:     1,
			PageSize: 1,
		})
		if listErr != nil {
			return "", false, listErr
		}
		if len(places) > 0 {
			return places[0].ID, false, nil
		}
	}

	place, createErr := h.createPlaceFromPost(ctx, post)
	if createErr != nil {
		return "", false, createErr
	}
	return place.ID, true, nil
}

func (h *PostDigestHandler) createPlaceFromPost(ctx context.Context, post *dbmodels.Post) (*domainPlace.Place, error) {
	now := time.Now().UTC()
	name := strings.TrimSpace(post.Title)
	if name == "" {
		name = "Untitled place"
	}
	if len(name) > 255 {
		name = name[:255]
	}
	desc := strings.TrimSpace(post.Description)
	place := &domainPlace.Place{
		ID:              uuid.NewString(),
		Name:            name,
		Description:     desc,
		Address:         post.Address,
		CreatedBy:       post.UserID,
		Status:          constants.PlaceStatusVisible,
		ConfidenceScore: 0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if post.Location.Valid && post.Location.Point != nil {
		place.Location = geotypes.NullPoint{
			Point: geom.NewPoint(geom.XY).MustSetCoords([]float64{post.Location.Point.X(), post.Location.Point.Y()}),
			Valid: true,
		}
	}
	if err := h.placeRepo.Add(ctx, nil, place); err != nil {
		return nil, err
	}
	return place, nil
}

func (h *PostDigestHandler) publishPostIngest(ctx context.Context, post *dbmodels.Post, placeID string) {
	if h.publisher == nil {
		return
	}
	kind := "post"
	if post.IsQuestion {
		kind = "question"
	}
	text := postText(post)
	if text == "" {
		return
	}
	msg := messaging.IngestMessage{
		ContentID: post.ID,
		PlaceID:   placeID,
		Kind:      kind,
		UserID:    post.UserID,
		Text:      text,
		Tags:      []string(post.Tags),
	}
	if post.Location.Valid && post.Location.Point != nil {
		lat := post.Location.Point.Y()
		lon := post.Location.Point.X()
		msg.Latitude = &lat
		msg.Longitude = &lon
	}
	_ = h.publisher.Publish(ctx, messaging.QueueAIIngest, msg)
}

func postText(post *dbmodels.Post) string {
	content := strings.TrimSpace(post.Title)
	if desc := strings.TrimSpace(post.Description); desc != "" {
		if content != "" {
			content += "\n\n"
		}
		content += desc
	}
	return strings.TrimSpace(content)
}
