package admin

import (
	"context"
	"time"
)

type PeriodCount struct {
	Current  int64
	Previous int64
}

type GrowthBucket struct {
	Day   time.Time
	Count int64
}

type AuditOptions struct {
	ActionType string
	Actor      string
	From       *time.Time
	To         *time.Time
	Page       int
	PageSize   int
}

type AuditEntry struct {
	ID         string
	Timestamp  time.Time
	ActorID    string
	ActorName  string
	ActionType string
	Details    string
}

type Repository interface {
	CountUsers(ctx context.Context) (int64, error)
	CountQuestions(ctx context.Context) (int64, error)
	CountPostsWithLocation(ctx context.Context) (int64, error)
	UserPeriodCount(ctx context.Context, window time.Duration) (*PeriodCount, error)
	CountUsersBefore(ctx context.Context, t time.Time) (int64, error)
	UserGrowthDaily(ctx context.Context, since time.Time) ([]GrowthBucket, error)
	CountUsersSince(ctx context.Context, since time.Time) (int64, error)
	HighReputationCount(ctx context.Context, threshold int64) (int64, error)
	ListAudit(ctx context.Context, opts *AuditOptions) ([]*AuditEntry, error)
	CountAudit(ctx context.Context, opts *AuditOptions) (int64, error)
}
