package endpoint

import (
	"context"
	"strings"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"

	"github.com/yetbota/content-service/internal/domain/apperr"
	"github.com/yetbota/content-service/internal/domain/auth"
	domain "github.com/yetbota/content-service/internal/domain/qa"
	qauc "github.com/yetbota/content-service/internal/services/usecase/qa"
)

type QAService interface {
	CreateQuestion(ctx context.Context, in qauc.CreateQuestionInput) (domain.Question, error)
	UpdateQuestion(ctx context.Context, in qauc.UpdateQuestionInput) (domain.Question, error)
	GetQuestion(ctx context.Context, id uuid.UUID) (domain.Question, error)
	ListQuestions(ctx context.Context, in qauc.ListInput) (qauc.PageResult, error)
	CloseQuestion(ctx context.Context, id, actorID uuid.UUID) (domain.Question, error)
	SetQuestionVote(ctx context.Context, questionID, userID uuid.UUID, vote *domain.Vote) (domain.Question, error)
	AddAnswer(ctx context.Context, questionID, authorID uuid.UUID, body string) (domain.Answer, error)
	UpdateAnswer(ctx context.Context, in qauc.UpdateAnswerInput) (domain.Answer, error)
	ListAnswers(ctx context.Context, questionID uuid.UUID, in qauc.ListInput) (qauc.AnswerPageResult, error)
	SetAnswerVote(ctx context.Context, answerID, userID uuid.UUID, vote *domain.Vote) (domain.Answer, error)
}

type CreateQuestionRequest struct {
	Principal auth.Principal
	Body      string   `json:"body"`
	PostID    *string  `json:"post_id"`
	Lat       *float64 `json:"lat"`
	Lng       *float64 `json:"lng"`
}

type UpdateQuestionRequest struct {
	Principal   auth.Principal
	QuestionID  uuid.UUID
	Body        *string  `json:"body"`
	PostID      *string  `json:"post_id"` // omit unchanged; "" clears; uuid sets
	Lat         *float64 `json:"lat"`
	Lng         *float64 `json:"lng"`
	ClearPlace  bool     `json:"clear_place"`
}

type UpdateAnswerRequest struct {
	Principal auth.Principal
	AnswerID  uuid.UUID
	Body      *string `json:"body"`
}

type QuestionIDRequest struct {
	Principal  auth.Principal
	QuestionID uuid.UUID
}

type ListQuestionsRequest struct {
	Principal auth.Principal
	PostID    *uuid.UUID
	Limit     int
	Cursor    string
}

type QuestionVoteRequest struct {
	Principal  auth.Principal
	QuestionID uuid.UUID
	Vote       *string `json:"vote"`
}

type AddAnswerRequest struct {
	Principal  auth.Principal
	QuestionID uuid.UUID
	Body       string `json:"body"`
}

type ListAnswersRequest struct {
	Principal  auth.Principal
	QuestionID uuid.UUID
	Limit      int
	Cursor     string
}

type AnswerVoteRequest struct {
	Principal auth.Principal
	AnswerID  uuid.UUID
	Vote      *string `json:"vote"`
}

type QuestionJSON struct {
	ID            string   `json:"id"`
	AuthorID      string   `json:"author_id"`
	Body          string   `json:"body"`
	PostID        *string  `json:"post_id,omitempty"`
	Lat           *float64 `json:"lat,omitempty"`
	Lng           *float64 `json:"lng,omitempty"`
	UpvoteCount   int64    `json:"upvote_count"`
	DownvoteCount int64    `json:"downvote_count"`
	AnswerCount   int64    `json:"answer_count"`
	ClosedAt      *string  `json:"closed_at,omitempty"`
	ClosedBy      *string  `json:"closed_by,omitempty"`
	CreatedAt     string   `json:"created_at"`
}

type QuestionsPageJSON struct {
	Questions  []QuestionJSON `json:"questions"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

type AnswerJSON struct {
	ID            string `json:"id"`
	QuestionID    string `json:"question_id"`
	AuthorID      string `json:"author_id"`
	Body          string `json:"body"`
	UpvoteCount   int64  `json:"upvote_count"`
	DownvoteCount int64  `json:"downvote_count"`
	CreatedAt     string `json:"created_at"`
}

type AnswersPageJSON struct {
	Answers    []AnswerJSON `json:"answers"`
	NextCursor string       `json:"next_cursor,omitempty"`
}

func MakeCreateQuestionEndpoint(svc QAService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CreateQuestionRequest)
		authorID, err := ParseUserID(req.Principal.UserID)
		if err != nil {
			return nil, err
		}
		var postID *uuid.UUID
		if req.PostID != nil && strings.TrimSpace(*req.PostID) != "" {
			id, err := uuid.Parse(strings.TrimSpace(*req.PostID))
			if err != nil {
				return nil, apperr.Validation(apperr.Field{
					Field: "post_id", Code: "INVALID", Message: "post_id must be a UUID",
				})
			}
			postID = &id
		}
		q, err := svc.CreateQuestion(ctx, qauc.CreateQuestionInput{
			AuthorID: authorID, Body: req.Body,
			PostID: postID, Lat: req.Lat, Lng: req.Lng,
		})
		if err != nil {
			return nil, err
		}
		return toQuestionJSON(q), nil
	}
}

func MakeUpdateQuestionEndpoint(svc QAService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdateQuestionRequest)
		actorID, err := ParseUserID(req.Principal.UserID)
		if err != nil {
			return nil, err
		}
		in := qauc.UpdateQuestionInput{
			QuestionID: req.QuestionID, ActorID: actorID,
			Body: req.Body, Lat: req.Lat, Lng: req.Lng, ClearPlace: req.ClearPlace,
		}
		if req.PostID != nil {
			raw := strings.TrimSpace(*req.PostID)
			if raw == "" {
				in.ClearPostID = true
			} else {
				id, err := uuid.Parse(raw)
				if err != nil {
					return nil, apperr.Validation(apperr.Field{
						Field: "post_id", Code: "INVALID", Message: "post_id must be a UUID",
					})
				}
				in.PostID = &id
			}
		}
		q, err := svc.UpdateQuestion(ctx, in)
		if err != nil {
			return nil, err
		}
		return toQuestionJSON(q), nil
	}
}

func MakeGetQuestionEndpoint(svc QAService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(QuestionIDRequest)
		q, err := svc.GetQuestion(ctx, req.QuestionID)
		if err != nil {
			return nil, err
		}
		return toQuestionJSON(q), nil
	}
}

func MakeListQuestionsEndpoint(svc QAService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ListQuestionsRequest)
		page, err := svc.ListQuestions(ctx, qauc.ListInput{
			Limit: req.Limit, Cursor: req.Cursor, PostID: req.PostID,
		})
		if err != nil {
			return nil, err
		}
		items := make([]QuestionJSON, 0, len(page.Questions))
		for _, q := range page.Questions {
			items = append(items, toQuestionJSON(q))
		}
		return QuestionsPageJSON{Questions: items, NextCursor: page.NextCursor}, nil
	}
}

func MakeCloseQuestionEndpoint(svc QAService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(QuestionIDRequest)
		actorID, err := ParseUserID(req.Principal.UserID)
		if err != nil {
			return nil, err
		}
		q, err := svc.CloseQuestion(ctx, req.QuestionID, actorID)
		if err != nil {
			return nil, err
		}
		return toQuestionJSON(q), nil
	}
}

func MakeQuestionVoteEndpoint(svc QAService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(QuestionVoteRequest)
		userID, err := ParseUserID(req.Principal.UserID)
		if err != nil {
			return nil, err
		}
		vote, err := parseVote(req.Vote)
		if err != nil {
			return nil, err
		}
		q, err := svc.SetQuestionVote(ctx, req.QuestionID, userID, vote)
		if err != nil {
			return nil, err
		}
		return toQuestionJSON(q), nil
	}
}

func MakeAddAnswerEndpoint(svc QAService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AddAnswerRequest)
		authorID, err := ParseUserID(req.Principal.UserID)
		if err != nil {
			return nil, err
		}
		a, err := svc.AddAnswer(ctx, req.QuestionID, authorID, req.Body)
		if err != nil {
			return nil, err
		}
		return toAnswerJSON(a), nil
	}
}

func MakeUpdateAnswerEndpoint(svc QAService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdateAnswerRequest)
		actorID, err := ParseUserID(req.Principal.UserID)
		if err != nil {
			return nil, err
		}
		a, err := svc.UpdateAnswer(ctx, qauc.UpdateAnswerInput{
			AnswerID: req.AnswerID, ActorID: actorID, Body: req.Body,
		})
		if err != nil {
			return nil, err
		}
		return toAnswerJSON(a), nil
	}
}

func MakeListAnswersEndpoint(svc QAService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ListAnswersRequest)
		page, err := svc.ListAnswers(ctx, req.QuestionID, qauc.ListInput{
			Limit: req.Limit, Cursor: req.Cursor,
		})
		if err != nil {
			return nil, err
		}
		items := make([]AnswerJSON, 0, len(page.Answers))
		for _, a := range page.Answers {
			items = append(items, toAnswerJSON(a))
		}
		return AnswersPageJSON{Answers: items, NextCursor: page.NextCursor}, nil
	}
}

func MakeAnswerVoteEndpoint(svc QAService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AnswerVoteRequest)
		userID, err := ParseUserID(req.Principal.UserID)
		if err != nil {
			return nil, err
		}
		vote, err := parseVote(req.Vote)
		if err != nil {
			return nil, err
		}
		a, err := svc.SetAnswerVote(ctx, req.AnswerID, userID, vote)
		if err != nil {
			return nil, err
		}
		return toAnswerJSON(a), nil
	}
}

func parseVote(raw *string) (*domain.Vote, error) {
	if raw == nil {
		return nil, nil
	}
	v := domain.Vote(strings.TrimSpace(*raw))
	switch v {
	case domain.VoteUp, domain.VoteDown:
		return &v, nil
	case "":
		return nil, nil
	default:
		return nil, apperr.Validation(apperr.Field{
			Field: "vote", Code: "INVALID", Message: "vote must be up, down, or null",
		})
	}
}

func toQuestionJSON(q domain.Question) QuestionJSON {
	out := QuestionJSON{
		ID: q.ID.String(), AuthorID: q.AuthorID.String(),
		Body: q.Body,
		UpvoteCount: q.UpvoteCount, DownvoteCount: q.DownvoteCount,
		AnswerCount: q.AnswerCount,
		CreatedAt:   q.CreatedAt.UTC().Format(time.RFC3339Nano),
		Lat:         q.Lat, Lng: q.Lng,
	}
	if q.PostID != nil {
		s := q.PostID.String()
		out.PostID = &s
	}
	if q.ClosedAt != nil {
		s := q.ClosedAt.UTC().Format(time.RFC3339Nano)
		out.ClosedAt = &s
	}
	if q.ClosedBy != nil {
		s := q.ClosedBy.String()
		out.ClosedBy = &s
	}
	return out
}

func toAnswerJSON(a domain.Answer) AnswerJSON {
	return AnswerJSON{
		ID: a.ID.String(), QuestionID: a.QuestionID.String(), AuthorID: a.AuthorID.String(),
		Body: a.Body, UpvoteCount: a.UpvoteCount, DownvoteCount: a.DownvoteCount,
		CreatedAt: a.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}
