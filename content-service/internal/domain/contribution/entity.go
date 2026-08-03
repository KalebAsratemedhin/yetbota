package contribution

import (
	"time"

	"github.com/aarondl/null/v8"
)

type Contribution struct {
	ID               string
	PlaceID          string
	UserID           string
	Kind             string
	Content          string
	ParentID         null.String
	Upvotes          int
	Downvotes        int
	ModerationStatus string
	Origin           string
	SourcePostID     null.String
	AcceptedAt       null.Time
	AcceptedBy       null.String
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
