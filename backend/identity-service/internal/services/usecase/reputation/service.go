package reputation

import (
	"context"

	"github.com/google/uuid"

	"github.com/yetbota/identity-service/internal/domain/apperr"
	domain "github.com/yetbota/identity-service/internal/domain/reputation"
)

type Config struct {
	PostCreatedDelta int
	PostLikedDelta   int
	DailyCap         int64
}

type Service struct {
	repo domain.Repository
	cfg  Config
}

func NewService(repo domain.Repository, cfg Config) *Service {
	if cfg.PostCreatedDelta <= 0 {
		cfg.PostCreatedDelta = 2
	}
	if cfg.PostLikedDelta <= 0 {
		cfg.PostLikedDelta = 1
	}
	if cfg.DailyCap <= 0 {
		cfg.DailyCap = 50
	}
	return &Service{repo: repo, cfg: cfg}
}

func LevelForScore(score int64) string {
	switch {
	case score >= 500:
		return "Pathfinder"
	case score >= 200:
		return "Local Guide"
	case score >= 50:
		return "Contributor"
	default:
		return "Explorer"
	}
}

func (s *Service) Snapshot(ctx context.Context, userID uuid.UUID) (domain.Snapshot, error) {
	score, err := s.repo.GetScore(ctx, userID)
	if err != nil {
		return domain.Snapshot{}, err
	}
	badges, err := s.repo.ListUserBadges(ctx, userID)
	if err != nil {
		return domain.Snapshot{}, err
	}
	if badges == nil {
		badges = []domain.Badge{}
	}
	return domain.Snapshot{Score: score, Level: LevelForScore(score), Badges: badges}, nil
}

func (s *Service) RecordEvent(ctx context.Context, in domain.EventInput) (domain.Snapshot, error) {
	if in.UserID == uuid.Nil {
		return domain.Snapshot{}, apperr.Validation(apperr.Required("user_id"))
	}
	delta, ok := s.deltaFor(in.EventType)
	if !ok {
		return domain.Snapshot{}, apperr.Validation(apperr.Field{
			Field: "event_type", Code: "INVALID", Message: "unknown event_type",
		})
	}
	// Self-actions never award.
	if in.ActorID != nil && *in.ActorID == in.UserID {
		return s.Snapshot(ctx, in.UserID)
	}
	if delta > 0 {
		earned, err := s.repo.SumPositiveDeltaToday(ctx, in.UserID)
		if err != nil {
			return domain.Snapshot{}, err
		}
		remaining := s.cfg.DailyCap - earned
		if remaining <= 0 {
			return s.Snapshot(ctx, in.UserID)
		}
		if int64(delta) > remaining {
			delta = int(remaining)
		}
	}
	if delta == 0 {
		return s.Snapshot(ctx, in.UserID)
	}

	if err := s.repo.InsertEvent(ctx, uuid.New(), in.UserID, in.ActorID, in.EventType, in.SubjectType, in.SubjectID, int32(delta)); err != nil {
		return domain.Snapshot{}, err
	}
	score, err := s.repo.AdjustScore(ctx, in.UserID, int64(delta))
	if err != nil {
		return domain.Snapshot{}, err
	}
	if err := s.evaluateBadges(ctx, in.UserID, in.EventType, score); err != nil {
		return domain.Snapshot{}, err
	}
	return s.Snapshot(ctx, in.UserID)
}

// AwardFirstFollowerBadge is called when someone follows userID.
func (s *Service) AwardFirstFollowerBadge(ctx context.Context, userID uuid.UUID) error {
	_, err := s.repo.AwardBadge(ctx, userID, domain.BadgeFirstFollower)
	return err
}

func (s *Service) deltaFor(eventType string) (int, bool) {
	switch eventType {
	case domain.EventPostCreated:
		return s.cfg.PostCreatedDelta, true
	case domain.EventPostLiked:
		return s.cfg.PostLikedDelta, true
	case domain.EventPostUnliked:
		return -s.cfg.PostLikedDelta, true
	default:
		return 0, false
	}
}

func (s *Service) evaluateBadges(ctx context.Context, userID uuid.UUID, eventType string, score int64) error {
	switch eventType {
	case domain.EventPostCreated:
		if _, err := s.repo.AwardBadge(ctx, userID, domain.BadgeFirstPost); err != nil {
			return err
		}
	case domain.EventPostLiked:
		n, err := s.repo.CountEventsByType(ctx, userID, domain.EventPostLiked)
		if err != nil {
			return err
		}
		if n >= 10 {
			if _, err := s.repo.AwardBadge(ctx, userID, domain.BadgeTenLikesReceived); err != nil {
				return err
			}
		}
	}
	if score >= 50 {
		if _, err := s.repo.AwardBadge(ctx, userID, domain.BadgeLevelContributor); err != nil {
			return err
		}
	}
	return nil
}
