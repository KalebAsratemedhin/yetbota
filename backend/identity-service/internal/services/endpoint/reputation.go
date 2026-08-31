package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"

	domainrep "github.com/yetbota/identity-service/internal/domain/reputation"
)

type ReputationService interface {
	RecordEvent(ctx context.Context, in domainrep.EventInput) (domainrep.Snapshot, error)
}

type RecordReputationRequest struct {
	UserID      string  `json:"user_id"`
	ActorID     *string `json:"actor_id"`
	EventType   string  `json:"event_type"`
	SubjectType string  `json:"subject_type"`
	SubjectID   *string `json:"subject_id"`
}

type ReputationSnapshotResponse struct {
	ReputationScore int64       `json:"reputation_score"`
	ReputationLevel string      `json:"reputation_level"`
	Badges          []BadgeJSON `json:"badges"`
}

type BadgeJSON struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AwardedAt   string `json:"awarded_at,omitempty"`
}

func MakeRecordReputationEndpoint(svc ReputationService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(RecordReputationRequest)
		userID, err := uuid.Parse(req.UserID)
		if err != nil {
			return nil, err
		}
		in := domainrep.EventInput{
			UserID:      userID,
			EventType:   req.EventType,
			SubjectType: req.SubjectType,
		}
		if req.ActorID != nil && *req.ActorID != "" {
			aid, err := uuid.Parse(*req.ActorID)
			if err != nil {
				return nil, err
			}
			in.ActorID = &aid
		}
		if req.SubjectID != nil && *req.SubjectID != "" {
			sid, err := uuid.Parse(*req.SubjectID)
			if err != nil {
				return nil, err
			}
			in.SubjectID = &sid
		}
		snap, err := svc.RecordEvent(ctx, in)
		if err != nil {
			return nil, err
		}
		return snapshotResponse(snap), nil
	}
}

func snapshotResponse(snap domainrep.Snapshot) ReputationSnapshotResponse {
	badges := make([]BadgeJSON, 0, len(snap.Badges))
	for _, b := range snap.Badges {
		bj := BadgeJSON{ID: b.ID, Name: b.Name, Description: b.Description}
		if b.AwardedAt != nil {
			bj.AwardedAt = b.AwardedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00")
		}
		badges = append(badges, bj)
	}
	return ReputationSnapshotResponse{
		ReputationScore: snap.Score,
		ReputationLevel: snap.Level,
		Badges:          badges,
	}
}
