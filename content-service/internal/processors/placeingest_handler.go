package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	rabbitmqDriver "github.com/beka-birhanu/yetbota/content-service/drivers/rabbitmq"
	domainContribution "github.com/beka-birhanu/yetbota/content-service/internal/domain/contribution"
	domainPlace "github.com/beka-birhanu/yetbota/content-service/internal/domain/place"
	processorsDomain "github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
	"github.com/beka-birhanu/yetbota/content-service/internal/messaging"
)

type PlaceIngestHandler struct {
	placeRepo        domainPlace.Repository
	contributionRepo domainContribution.Repository
	publisher        *rabbitmqDriver.Publisher
	executor         processorsDomain.Executor
	consumer         *rabbitmqDriver.Consumer
}

type PlaceIngestHandlerConfig struct {
	PlaceRepo        domainPlace.Repository        `validate:"required"`
	ContributionRepo domainContribution.Repository `validate:"required"`
	Publisher        *rabbitmqDriver.Publisher     `validate:"required"`
	Executor         processorsDomain.Executor     `validate:"required"`
	Consumer         *rabbitmqDriver.Consumer      `validate:"required"`
}

func NewPlaceIngestHandler(cfg *PlaceIngestHandlerConfig) (*PlaceIngestHandler, error) {
	if err := validator.Validate.Struct(cfg); err != nil {
		return nil, err
	}
	return &PlaceIngestHandler{
		placeRepo:        cfg.PlaceRepo,
		contributionRepo: cfg.ContributionRepo,
		publisher:        cfg.Publisher,
		executor:         cfg.Executor,
		consumer:         cfg.Consumer,
	}, nil
}

func (h *PlaceIngestHandler) Run(ctx context.Context) error {
	return h.consumer.Consume(ctx, messaging.QueuePlaceIngest, h.handle)
}

func (h *PlaceIngestHandler) handle(ctx context.Context, body []byte, _ int) error {
	var asMap map[string]any
	if err := json.Unmarshal(body, &asMap); err != nil {
		return fmt.Errorf("decode place ingest message: %w", err)
	}
	if cid, ok := asMap["contribution_id"].(string); ok && strings.TrimSpace(cid) != "" {
		return h.handleContribution(ctx, cid)
	}
	if pid, ok := asMap["place_id"].(string); ok && strings.TrimSpace(pid) != "" {
		return h.handlePlace(ctx, pid)
	}
	return fmt.Errorf("place ingest message missing place_id or contribution_id")
}

func (h *PlaceIngestHandler) handlePlace(ctx context.Context, placeID string) error {
	place, err := h.placeRepo.Read(ctx, placeID)
	if err != nil {
		return err
	}
	text := strings.TrimSpace(place.Name)
	if desc := strings.TrimSpace(place.Description); desc != "" {
		text = strings.TrimSpace(text + "\n\n" + desc)
	}
	if place.Address.Valid && strings.TrimSpace(place.Address.String) != "" {
		text = strings.TrimSpace(text + "\n\n" + place.Address.String)
	}
	if text == "" {
		return nil
	}
	msg := messaging.IngestMessage{
		ContentID: place.ID,
		PlaceID:   place.ID,
		Kind:      "place",
		UserID:    place.CreatedBy,
		Text:      text,
	}
	if place.Location.Valid && place.Location.Point != nil {
		lat := place.Location.Point.Y()
		lon := place.Location.Point.X()
		msg.Latitude = &lat
		msg.Longitude = &lon
	}
	if err := h.publisher.Publish(ctx, messaging.QueueAIIngest, msg); err != nil {
		return err
	}
	_ = h.executor.TriggerPlaceSummarizeWorkflow(ctx, processorsDomain.PlaceSummarizeWorkflowInput{PlaceID: place.ID})
	return nil
}

func (h *PlaceIngestHandler) handleContribution(ctx context.Context, contributionID string) error {
	c, err := h.contributionRepo.Read(ctx, contributionID)
	if err != nil {
		return err
	}
	place, err := h.placeRepo.Read(ctx, c.PlaceID)
	if err != nil {
		return err
	}
	content := strings.TrimSpace(c.Content)
	if content == "" {
		return nil
	}
	msg := messaging.IngestMessage{
		ContentID: c.ID,
		PlaceID:   c.PlaceID,
		Kind:      c.Kind,
		UserID:    c.UserID,
		Text:      content,
	}
	if c.ParentID.Valid {
		msg.ParentID = c.ParentID.String
	}
	if place.Location.Valid && place.Location.Point != nil {
		lat := place.Location.Point.Y()
		lon := place.Location.Point.X()
		msg.Latitude = &lat
		msg.Longitude = &lon
	}
	if err := h.publisher.Publish(ctx, messaging.QueueAIIngest, msg); err != nil {
		return err
	}
	_ = h.executor.TriggerPlaceSummarizeWorkflow(ctx, processorsDomain.PlaceSummarizeWorkflowInput{PlaceID: c.PlaceID})
	return nil
}
