package place

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	CategoryTip       = "tip"
	CategoryChange    = "change"
	CategoryImportant = "important"

	SourceAI   = "ai"
	SourceUser = "user"
)

type Place struct {
	ID               uuid.UUID
	Name             string
	Lat              float64
	Lng              float64
	Summary          string
	SummaryModel     string
	SummaryUpdatedAt *time.Time
	PostCount        int64
	Embedding        []float64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Entry struct {
	ID           uuid.UUID
	PlaceID      uuid.UUID
	Category     string
	Body         string
	Source       string
	SourcePostID *uuid.UUID
	AuthorID     *uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type PageCursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

type AttachCursor struct {
	AttachedAt time.Time
	PostID     uuid.UUID
}

type ListParams struct {
	Limit  int
	Cursor *PageCursor
	Query  string
}

type NearbyParams struct {
	Lat      float64
	Lng      float64
	RadiusM  float64
	Limit    int
}

type EntryListParams struct {
	Limit    int
	Cursor   *PageCursor
	Category string // empty = all
}

type PostIDPage struct {
	PostIDs    []uuid.UUID
	NextCursor *AttachCursor
}

type PlacePage struct {
	Places     []Place
	NextCursor *PageCursor
}

// RankedPlace is a place with absolute engagement for /places/top.
type RankedPlace struct {
	Place
	EngagementScore int64
	CoverURL        string
}

type EngagementCursor struct {
	Engagement int64
	ID         uuid.UUID
}

type RankedPlacePage struct {
	Places     []RankedPlace
	NextCursor *EngagementCursor
}

type EntryPage struct {
	Entries    []Entry
	NextCursor *PageCursor
}

type CreateInput struct {
	Name      string
	Lat       float64
	Lng       float64
	Embedding []float64
}

type AttachInput struct {
	PlaceID      uuid.UUID
	PostID       uuid.UUID
	AttachScore  float64
	AttachMethod string
	Embedding    []float64 // optional: update place centroid
}

type CreateAndAttachInput struct {
	Name         string
	Lat          float64
	Lng          float64
	PostID       uuid.UUID
	AttachScore  float64
	AttachMethod string
	Embedding    []float64
}

type AddEntryInput struct {
	PlaceID      uuid.UUID
	Category     string
	Body         string
	Source       string
	SourcePostID *uuid.UUID
	AuthorID     *uuid.UUID
}

type UpdateSummaryInput struct {
	PlaceID      uuid.UUID
	Summary      string
	SummaryModel string
}

type Repository interface {
	Create(ctx context.Context, in CreateInput) (Place, error)
	Get(ctx context.Context, id uuid.UUID) (Place, error)
	GetByPostID(ctx context.Context, postID uuid.UUID) (Place, error)
	List(ctx context.Context, p ListParams) (PlacePage, error)
	ListTopByEngagement(ctx context.Context, cursor *EngagementCursor, limit int) (RankedPlacePage, error)
	ListNearby(ctx context.Context, p NearbyParams) ([]Place, error)
	AttachPost(ctx context.Context, in AttachInput) (Place, error)
	CreateAndAttach(ctx context.Context, in CreateAndAttachInput) (Place, error)
	ListPostIDs(ctx context.Context, placeID uuid.UUID, limit int, cursor *AttachCursor) (PostIDPage, error)
	UpdateSummary(ctx context.Context, in UpdateSummaryInput) (Place, error)
	AddEntry(ctx context.Context, in AddEntryInput) (Entry, error)
	ListEntries(ctx context.Context, placeID uuid.UUID, p EntryListParams) (EntryPage, error)
}
