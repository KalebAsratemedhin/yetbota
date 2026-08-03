package contributionvote

import (
	"context"
	"database/sql"
	"time"
)

type Vote struct {
	ID             string
	ContributionID string
	UserID         string
	VoteType       string
	CreatedAt      time.Time
}

type ListOptions struct {
	UserID          string
	ContributionIDs []string
}

type Repository interface {
	Exists(ctx context.Context, userID, contributionID string) (bool, error)
	Read(ctx context.Context, userID, contributionID string) (*Vote, error)
	Add(ctx context.Context, tx *sql.Tx, entity *Vote) error
	Update(ctx context.Context, tx *sql.Tx, entity *Vote) error
	List(ctx context.Context, opts *ListOptions) (map[string]*Vote, error)
}
