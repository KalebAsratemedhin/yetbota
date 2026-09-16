package qa

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Vote string

const (
	VoteUp   Vote = "up"
	VoteDown Vote = "down"
)

type Question struct {
	ID            uuid.UUID
	AuthorID      uuid.UUID
	Body          string
	PostID        *uuid.UUID
	Lat           *float64
	Lng           *float64
	UpvoteCount   int64
	DownvoteCount int64
	AnswerCount   int64
	ClosedAt      *time.Time
	ClosedBy      *uuid.UUID
	CreatedAt     time.Time
}

func (q Question) Closed() bool {
	return q.ClosedAt != nil
}

type Answer struct {
	ID            uuid.UUID
	QuestionID    uuid.UUID
	AuthorID      uuid.UUID
	Body          string
	UpvoteCount   int64
	DownvoteCount int64
	CreatedAt     time.Time
}

type PageCursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

type ListParams struct {
	Limit  int
	Cursor *PageCursor
	PostID *uuid.UUID
}

type QuestionPage struct {
	Questions  []Question
	NextCursor *PageCursor
}

type AnswerPage struct {
	Answers    []Answer
	NextCursor *PageCursor
}

type CreateQuestionInput struct {
	AuthorID uuid.UUID
	Body     string
	PostID   *uuid.UUID
	Lat      *float64
	Lng      *float64
}

type CreateAnswerInput struct {
	QuestionID uuid.UUID
	AuthorID   uuid.UUID
	Body       string
}

type UpdateQuestionInput struct {
	ID     uuid.UUID
	Body   string
	PostID *uuid.UUID
	Lat    *float64
	Lng    *float64
}

type UpdateAnswerInput struct {
	ID   uuid.UUID
	Body string
}

type Repository interface {
	CreateQuestion(ctx context.Context, in CreateQuestionInput) (Question, error)
	UpdateQuestion(ctx context.Context, in UpdateQuestionInput) (Question, error)
	GetQuestion(ctx context.Context, id uuid.UUID) (Question, error)
	ListQuestions(ctx context.Context, p ListParams) (QuestionPage, error)
	CloseQuestion(ctx context.Context, id, closedBy uuid.UUID) (Question, error)
	SetQuestionVote(ctx context.Context, questionID, userID uuid.UUID, vote *Vote) (Question, error)
	CreateAnswer(ctx context.Context, in CreateAnswerInput) (Answer, error)
	UpdateAnswer(ctx context.Context, in UpdateAnswerInput) (Answer, error)
	GetAnswer(ctx context.Context, id uuid.UUID) (Answer, error)
	ListAnswers(ctx context.Context, questionID uuid.UUID, p ListParams) (AnswerPage, error)
	SetAnswerVote(ctx context.Context, answerID, userID uuid.UUID, vote *Vote) (Answer, error)
	PostExists(ctx context.Context, postID uuid.UUID) (bool, error)
}
