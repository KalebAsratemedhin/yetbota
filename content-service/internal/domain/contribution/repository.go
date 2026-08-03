package contribution

import (
	"context"
	"database/sql"
	"time"
)

type ListOptions struct {
	PlaceID      string
	ParentID     *string
	Kind         string
	Page         int
	PageSize     int
	OnlyVisible  bool
	NonAnswerTop bool
	OpenOnly     bool
	NearLat      *float64
	NearLon      *float64
	RadiusKm     *float64
	SourcePostID string
}

type Repository interface {
	Add(ctx context.Context, tx *sql.Tx, entity *Contribution) error
	Read(ctx context.Context, id string) (*Contribution, error)
	List(ctx context.Context, opts *ListOptions) ([]*Contribution, error)
	Count(ctx context.Context, opts *ListOptions) (int64, error)
	Update(ctx context.Context, tx *sql.Tx, entity *Contribution) error
	UpdateVoteCounts(ctx context.Context, tx *sql.Tx, id string, upDelta, downDelta, expectedUp, expectedDown int) error
	UpdateModerationStatus(ctx context.Context, tx *sql.Tx, id, status string) error
	ClearAcceptedAnswer(ctx context.Context, tx *sql.Tx, questionID string) error
	AcceptAnswer(ctx context.Context, tx *sql.Tx, answerID, acceptedBy string, acceptedAt time.Time) error
	ExistsForSourcePost(ctx context.Context, sourcePostID string) (bool, error)
}
