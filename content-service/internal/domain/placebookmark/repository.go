package placebookmark

import (
	"context"
	"database/sql"
	"time"
)

type Bookmark struct {
	UserID    string
	PlaceID   string
	CreatedAt time.Time
}

type Repository interface {
	Add(ctx context.Context, tx *sql.Tx, userID, placeID string) error
	Delete(ctx context.Context, tx *sql.Tx, userID, placeID string) error
	Exists(ctx context.Context, userID, placeID string) (bool, error)
	ListPlaceIDs(ctx context.Context, userID string) ([]string, error)
}
