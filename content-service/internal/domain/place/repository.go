package place

import (
	"context"
	"database/sql"
	"time"
)

type ListOptions struct {
	NearLat    *float64
	NearLon    *float64
	RadiusKm   *float64
	Search     string
	Page       int
	PageSize   int
	Status     *string
	CreatedBy  string
	NeedsCheck *bool
}

type Repository interface {
	Add(ctx context.Context, tx *sql.Tx, entity *Place) error
	Read(ctx context.Context, id string) (*Place, error)
	Update(ctx context.Context, tx *sql.Tx, entity *Place) error
	List(ctx context.Context, opts *ListOptions) ([]*Place, error)
	Count(ctx context.Context, opts *ListOptions) (int64, error)
	UpdateStatus(ctx context.Context, tx *sql.Tx, id, status string) error
	RecordConfirmation(ctx context.Context, tx *sql.Tx, id string, confirmedAt time.Time, confidence int16) error
	RecordContributionActivity(ctx context.Context, tx *sql.Tx, id string, at time.Time) error
	MarkStale(ctx context.Context, decayDays int) (int64, error)
	UpdateSummary(ctx context.Context, tx *sql.Tx, id, summaryText string, generatedAt time.Time, citationsJSON []byte) error
}
