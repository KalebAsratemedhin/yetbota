package placeproblem

import (
	"context"
	"database/sql"
	"time"

	"github.com/aarondl/null/v8"
)

const (
	KindPermanentlyClosed = "PERMANENTLY_CLOSED"
	KindDoesNotExist      = "DOES_NOT_EXIST"
	KindDuplicate         = "DUPLICATE"
	KindWrongLocation     = "WRONG_LOCATION"

	StatusPending   = "PENDING"
	StatusResolved  = "RESOLVED"
	StatusDismissed = "DISMISSED"
)

type Report struct {
	ID          string
	PlaceID     string
	UserID      string
	Kind        string
	Note        null.String
	DuplicateOf null.String
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Repository interface {
	Add(ctx context.Context, tx *sql.Tx, entity *Report) error
}
