package reputation

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	EventPostCreated = "post.created"
	EventPostLiked   = "post.liked"
	EventPostUnliked = "post.unliked"

	BadgeFirstPost        = "first_post"
	BadgeTenLikesReceived = "ten_likes_received"
	BadgeFirstFollower    = "first_follower"
	BadgeLevelContributor = "level_contributor"
)

type EventInput struct {
	UserID      uuid.UUID
	ActorID     *uuid.UUID
	EventType   string
	SubjectType string
	SubjectID   *uuid.UUID
}

type Badge struct {
	ID          string
	Name        string
	Description string
	AwardedAt   *time.Time // nil if catalog-only
}

type Snapshot struct {
	Score  int64
	Level  string
	Badges []Badge
}

type Repository interface {
	AdjustScore(ctx context.Context, userID uuid.UUID, delta int64) (int64, error)
	InsertEvent(ctx context.Context, id uuid.UUID, userID uuid.UUID, actorID *uuid.UUID, eventType, subjectType string, subjectID *uuid.UUID, delta int32) error
	SumPositiveDeltaToday(ctx context.Context, userID uuid.UUID) (int64, error)
	CountEventsByType(ctx context.Context, userID uuid.UUID, eventType string) (int64, error)
	ListUserBadges(ctx context.Context, userID uuid.UUID) ([]Badge, error)
	AwardBadge(ctx context.Context, userID uuid.UUID, badgeID string) (bool, error)
	HasBadge(ctx context.Context, userID uuid.UUID, badgeID string) (bool, error)
	GetScore(ctx context.Context, userID uuid.UUID) (int64, error)
}
