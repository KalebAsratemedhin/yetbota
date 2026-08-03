package placephoto

import (
	"context"
	"database/sql"
	"time"

	"github.com/aarondl/null/v8"
)

type Photo struct {
	ID        string
	URL       string
	URLMobile null.String
	URLWeb    null.String
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PlacePhoto struct {
	PlaceID   string
	PhotoID   string
	Position  int
	PhotoData *Photo
}

type Options struct {
	PlaceIDs   []string
	LoadPhoto  bool
	PositionEq *int
}

type SortOptions struct {
	Field     SortField
	Direction SortDirection
}

type (
	SortField     string
	SortDirection string
)

const (
	SortDirectionAsc  SortDirection = "ASC"
	SortDirectionDesc SortDirection = "DESC"
)

const (
	SortFieldPlaceID  SortField = "place_id"
	SortFieldPhotoID  SortField = "photo_id"
	SortFieldPosition SortField = "position"
)

type Repository interface {
	Add(ctx context.Context, tx *sql.Tx, entity *PlacePhoto) error
	AddBulk(ctx context.Context, tx *sql.Tx, entities []*PlacePhoto) error
	Update(ctx context.Context, tx *sql.Tx, entity *PlacePhoto) error
	Delete(ctx context.Context, tx *sql.Tx, placeID string, position int) error
	List(ctx context.Context, opts *Options, sort *SortOptions) ([]*PlacePhoto, error)
}
