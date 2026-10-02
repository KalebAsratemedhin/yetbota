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
	EventScoreSync   = "score.sync"

	BadgeNewcomer           = "level_newcomer"
	BadgeRegular            = "level_regular"
	BadgeContributor        = "level_contributor"
	BadgeLocalGuide         = "level_local_guide"
	BadgeExplorer           = "level_explorer"
	BadgeLegendaryExplorer  = "level_legendary_explorer"
	BadgePathfinder         = "level_pathfinder"
)

// ScoreBadge is a one-shot award when reputation score reaches MinScore.
type ScoreBadge struct {
	ID       string
	MinScore int64
}

// ScoreBadges is ordered low → high for evaluation.
var ScoreBadges = []ScoreBadge{
	{ID: BadgeNewcomer, MinScore: 10},
	{ID: BadgeRegular, MinScore: 40},
	{ID: BadgeContributor, MinScore: 100},
	{ID: BadgeLocalGuide, MinScore: 400},
	{ID: BadgeExplorer, MinScore: 550},
	{ID: BadgeLegendaryExplorer, MinScore: 850},
	{ID: BadgePathfinder, MinScore: 1000},
}

// SyncInput sets denormalized score from content-side engagement math (Option 1).
type SyncInput struct {
	UserID        uuid.UUID
	Score         int64
	PostsCreated  int64
	LikesReceived int64
}

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
	SetScore(ctx context.Context, userID uuid.UUID, score int64) (int64, error)
	InsertEvent(ctx context.Context, id uuid.UUID, userID uuid.UUID, actorID *uuid.UUID, eventType, subjectType string, subjectID *uuid.UUID, delta int32) error
	SumPositiveDeltaToday(ctx context.Context, userID uuid.UUID) (int64, error)
	CountEventsByType(ctx context.Context, userID uuid.UUID, eventType string) (int64, error)
	ListUserBadges(ctx context.Context, userID uuid.UUID) ([]Badge, error)
	AwardBadge(ctx context.Context, userID uuid.UUID, badgeID string) (bool, error)
	HasBadge(ctx context.Context, userID uuid.UUID, badgeID string) (bool, error)
	GetScore(ctx context.Context, userID uuid.UUID) (int64, error)
}
