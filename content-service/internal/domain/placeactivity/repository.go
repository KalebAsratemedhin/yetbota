package placeactivity

import (
	"context"
	"database/sql"
	"time"

	"github.com/aarondl/null/v8"
)

const (
	EventContributionAdded   = "contribution_added"
	EventQuestionOpened      = "question_opened"
	EventAnswerAccepted      = "answer_accepted"
	EventPlaceConfirmed      = "place_confirmed"
	EventPlaceNeedsCheck     = "place_needs_check"
	EventPostLinked          = "post_linked"
	EventProblemReported     = "problem_reported"
)

type Event struct {
	ID        string
	PlaceID   string
	ActorID   null.String
	EventType string
	RefID     null.String
	Summary   string
	CreatedAt time.Time
}

type ListOptions struct {
	PlaceIDs []string
	Page     int
	PageSize int
}

type Repository interface {
	Add(ctx context.Context, tx *sql.Tx, event *Event) error
	List(ctx context.Context, opts *ListOptions) ([]*Event, error)
	Count(ctx context.Context, opts *ListOptions) (int64, error)
}
