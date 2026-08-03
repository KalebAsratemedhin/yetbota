package placeconfirmation

import (
	"context"
	"database/sql"
	"time"

	"github.com/aarondl/null/v8"
)

type Confirmation struct {
	ID          string
	PlaceID     string
	UserID      string
	Note        null.String
	ConfirmedAt time.Time
}

type Repository interface {
	Upsert(ctx context.Context, tx *sql.Tx, entity *Confirmation) error
	CountSince(ctx context.Context, placeID string, since time.Time) (int64, error)
}
