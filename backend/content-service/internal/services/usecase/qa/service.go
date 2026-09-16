package qa

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/yetbota/content-service/internal/domain/apperr"
	domain "github.com/yetbota/content-service/internal/domain/qa"
)

// ContentEventPublisher notifies ai-service (thin ID events).
type ContentEventPublisher interface {
	PublishQuestionCreated(ctx context.Context, questionID uuid.UUID, postID *uuid.UUID) error
	PublishQuestionUpdated(ctx context.Context, questionID uuid.UUID, postID *uuid.UUID) error
	PublishAnswerCreated(ctx context.Context, answerID, questionID uuid.UUID) error
	PublishAnswerUpdated(ctx context.Context, answerID, questionID uuid.UUID) error
}

type Service struct {
	repo            domain.Repository
	events          ContentEventPublisher
	defaultPageSize int
	maxPageSize     int
}

func NewService(repo domain.Repository, events ContentEventPublisher, defaultPageSize, maxPageSize int) *Service {
	if defaultPageSize <= 0 {
		defaultPageSize = 20
	}
	if maxPageSize <= 0 {
		maxPageSize = 50
	}
	return &Service{repo: repo, events: events, defaultPageSize: defaultPageSize, maxPageSize: maxPageSize}
}

type CreateQuestionInput struct {
	AuthorID uuid.UUID
	Body     string
	PostID   *uuid.UUID
	Lat      *float64
	Lng      *float64
}

// UpdateQuestionInput is a PATCH: nil pointers mean leave unchanged.
// PostID: nil = unchanged; non-nil empty UUID clear via ClearPostID; use ClearPostID for unlink.
type UpdateQuestionInput struct {
	QuestionID  uuid.UUID
	ActorID     uuid.UUID
	Body        *string
	PostID      *uuid.UUID // set when updating link; ignored if ClearPostID
	ClearPostID bool
	Lat         *float64
	Lng         *float64
	ClearPlace  bool // clears lat/lng
}

type UpdateAnswerInput struct {
	AnswerID uuid.UUID
	ActorID  uuid.UUID
	Body     *string
}

type ListInput struct {
	Limit  int
	Cursor string
	PostID *uuid.UUID
}

type PageResult struct {
	Questions  []domain.Question
	NextCursor string
}

type AnswerPageResult struct {
	Answers    []domain.Answer
	NextCursor string
}

func (s *Service) CreateQuestion(ctx context.Context, in CreateQuestionInput) (domain.Question, error) {
	body := strings.TrimSpace(in.Body)
	var details []apperr.Field
	if body == "" {
		details = append(details, apperr.Required("body"))
	} else if utf8.RuneCountInString(body) > 10000 {
		details = append(details, apperr.Field{Field: "body", Code: "TOO_LONG", Message: "body must be at most 10000 characters"})
	}
	// Optional place location (same idea as post lat/lng — a chosen place, not only GPS).
	if (in.Lat == nil) != (in.Lng == nil) {
		details = append(details, apperr.Field{
			Field: "lat/lng", Code: "INVALID", Message: "lat and lng must both be provided or both omitted",
		})
	} else if in.Lat != nil && (*in.Lat < -90 || *in.Lat > 90 || *in.Lng < -180 || *in.Lng > 180) {
		details = append(details, apperr.Field{Field: "lat/lng", Code: "INVALID", Message: "lat/lng out of range"})
	}
	if len(details) > 0 {
		return domain.Question{}, apperr.Validation(details...)
	}
	if in.PostID != nil {
		ok, err := s.repo.PostExists(ctx, *in.PostID)
		if err != nil {
			return domain.Question{}, err
		}
		if !ok {
			return domain.Question{}, apperr.Validation(apperr.Field{
				Field: "post_id", Code: "NOT_FOUND", Message: "post not found",
			})
		}
	}
	q, err := s.repo.CreateQuestion(ctx, domain.CreateQuestionInput{
		AuthorID: in.AuthorID, Body: body,
		PostID: in.PostID, Lat: in.Lat, Lng: in.Lng,
	})
	if err != nil {
		return domain.Question{}, err
	}
	if s.events != nil {
		if err := s.events.PublishQuestionCreated(ctx, q.ID, q.PostID); err != nil {
			slog.Warn("question_created_publish_failed", "question_id", q.ID.String(), "err", err.Error())
		}
	}
	return q, nil
}

func (s *Service) UpdateQuestion(ctx context.Context, in UpdateQuestionInput) (domain.Question, error) {
	cur, err := s.repo.GetQuestion(ctx, in.QuestionID)
	if err != nil {
		return domain.Question{}, err
	}
	if cur.AuthorID != in.ActorID {
		return domain.Question{}, apperr.Validation(apperr.Field{
			Field: "id", Code: "FORBIDDEN", Message: "only the author can edit this question",
		})
	}
	if cur.Closed() {
		return domain.Question{}, apperr.Validation(apperr.Field{
			Field: "id", Code: "CLOSED", Message: "closed questions cannot be edited",
		})
	}

	body := cur.Body
	postID := cur.PostID
	lat, lng := cur.Lat, cur.Lng
	changed := false

	if in.Body != nil {
		changed = true
		body = strings.TrimSpace(*in.Body)
	}
	if in.ClearPostID {
		changed = true
		postID = nil
	} else if in.PostID != nil {
		changed = true
		postID = in.PostID
	}
	if in.ClearPlace {
		changed = true
		lat, lng = nil, nil
	} else if (in.Lat == nil) != (in.Lng == nil) {
		return domain.Question{}, apperr.Validation(apperr.Field{
			Field: "lat/lng", Code: "INVALID", Message: "lat and lng must both be provided or both omitted",
		})
	} else if in.Lat != nil {
		changed = true
		lat, lng = in.Lat, in.Lng
	}
	if !changed {
		return domain.Question{}, apperr.Validation(apperr.Field{
			Field: "body", Code: "REQUIRED", Message: "provide at least one field to update",
		})
	}

	var details []apperr.Field
	if body == "" {
		details = append(details, apperr.Required("body"))
	} else if utf8.RuneCountInString(body) > 10000 {
		details = append(details, apperr.Field{Field: "body", Code: "TOO_LONG", Message: "body must be at most 10000 characters"})
	}
	if lat != nil && (*lat < -90 || *lat > 90 || *lng < -180 || *lng > 180) {
		details = append(details, apperr.Field{Field: "lat/lng", Code: "INVALID", Message: "lat/lng out of range"})
	}
	if len(details) > 0 {
		return domain.Question{}, apperr.Validation(details...)
	}
	if postID != nil {
		ok, err := s.repo.PostExists(ctx, *postID)
		if err != nil {
			return domain.Question{}, err
		}
		if !ok {
			return domain.Question{}, apperr.Validation(apperr.Field{
				Field: "post_id", Code: "NOT_FOUND", Message: "post not found",
			})
		}
	}

	q, err := s.repo.UpdateQuestion(ctx, domain.UpdateQuestionInput{
		ID: in.QuestionID, Body: body, PostID: postID, Lat: lat, Lng: lng,
	})
	if err != nil {
		return domain.Question{}, err
	}
	if s.events != nil {
		if err := s.events.PublishQuestionUpdated(ctx, q.ID, q.PostID); err != nil {
			slog.Warn("question_updated_publish_failed", "question_id", q.ID.String(), "err", err.Error())
		}
	}
	return q, nil
}

func (s *Service) GetQuestion(ctx context.Context, id uuid.UUID) (domain.Question, error) {
	return s.repo.GetQuestion(ctx, id)
}

func (s *Service) ListQuestions(ctx context.Context, in ListInput) (PageResult, error) {
	p, err := s.listParams(in.Limit, in.Cursor)
	if err != nil {
		return PageResult{}, err
	}
	p.PostID = in.PostID
	page, err := s.repo.ListQuestions(ctx, p)
	if err != nil {
		return PageResult{}, err
	}
	out := PageResult{Questions: page.Questions}
	if page.NextCursor != nil {
		out.NextCursor = encodeCursor(*page.NextCursor)
	}
	return out, nil
}

func (s *Service) CloseQuestion(ctx context.Context, id, actorID uuid.UUID) (domain.Question, error) {
	q, err := s.repo.GetQuestion(ctx, id)
	if err != nil {
		return domain.Question{}, err
	}
	if q.AuthorID != actorID {
		return domain.Question{}, apperr.Validation(apperr.Field{
			Field: "id", Code: "FORBIDDEN", Message: "only the author can close this question",
		})
	}
	if q.Closed() {
		return domain.Question{}, apperr.Validation(apperr.Field{
			Field: "id", Code: "CLOSED", Message: "question is already closed",
		})
	}
	return s.repo.CloseQuestion(ctx, id, actorID)
}

func (s *Service) SetQuestionVote(ctx context.Context, questionID, userID uuid.UUID, vote *domain.Vote) (domain.Question, error) {
	if err := validateVote(vote); err != nil {
		return domain.Question{}, err
	}
	q, err := s.repo.GetQuestion(ctx, questionID)
	if err != nil {
		return domain.Question{}, err
	}
	if q.AuthorID == userID {
		return domain.Question{}, apperr.Validation(apperr.Field{
			Field: "vote", Code: "FORBIDDEN", Message: "you can’t vote on your own question",
		})
	}
	return s.repo.SetQuestionVote(ctx, questionID, userID, vote)
}

func (s *Service) AddAnswer(ctx context.Context, questionID, authorID uuid.UUID, body string) (domain.Answer, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return domain.Answer{}, apperr.Validation(apperr.Required("body"))
	}
	if utf8.RuneCountInString(body) > 10000 {
		return domain.Answer{}, apperr.Validation(apperr.Field{
			Field: "body", Code: "TOO_LONG", Message: "body must be at most 10000 characters",
		})
	}
	a, err := s.repo.CreateAnswer(ctx, domain.CreateAnswerInput{
		QuestionID: questionID, AuthorID: authorID, Body: body,
	})
	if err != nil {
		return domain.Answer{}, err
	}
	if s.events != nil {
		if err := s.events.PublishAnswerCreated(ctx, a.ID, questionID); err != nil {
			slog.Warn("answer_created_publish_failed", "answer_id", a.ID.String(), "err", err.Error())
		}
	}
	return a, nil
}

func (s *Service) UpdateAnswer(ctx context.Context, in UpdateAnswerInput) (domain.Answer, error) {
	cur, err := s.repo.GetAnswer(ctx, in.AnswerID)
	if err != nil {
		return domain.Answer{}, err
	}
	if cur.AuthorID != in.ActorID {
		return domain.Answer{}, apperr.Validation(apperr.Field{
			Field: "id", Code: "FORBIDDEN", Message: "only the author can edit this answer",
		})
	}
	if in.Body == nil {
		return domain.Answer{}, apperr.Validation(apperr.Field{
			Field: "body", Code: "REQUIRED", Message: "provide body to update",
		})
	}
	body := strings.TrimSpace(*in.Body)
	if body == "" {
		return domain.Answer{}, apperr.Validation(apperr.Required("body"))
	}
	if utf8.RuneCountInString(body) > 10000 {
		return domain.Answer{}, apperr.Validation(apperr.Field{
			Field: "body", Code: "TOO_LONG", Message: "body must be at most 10000 characters",
		})
	}
	a, err := s.repo.UpdateAnswer(ctx, domain.UpdateAnswerInput{ID: in.AnswerID, Body: body})
	if err != nil {
		return domain.Answer{}, err
	}
	if s.events != nil {
		if err := s.events.PublishAnswerUpdated(ctx, a.ID, a.QuestionID); err != nil {
			slog.Warn("answer_updated_publish_failed", "answer_id", a.ID.String(), "err", err.Error())
		}
	}
	return a, nil
}

func (s *Service) ListAnswers(ctx context.Context, questionID uuid.UUID, in ListInput) (AnswerPageResult, error) {
	p, err := s.listParams(in.Limit, in.Cursor)
	if err != nil {
		return AnswerPageResult{}, err
	}
	page, err := s.repo.ListAnswers(ctx, questionID, p)
	if err != nil {
		return AnswerPageResult{}, err
	}
	out := AnswerPageResult{Answers: page.Answers}
	if page.NextCursor != nil {
		out.NextCursor = encodeCursor(*page.NextCursor)
	}
	return out, nil
}

func (s *Service) SetAnswerVote(ctx context.Context, answerID, userID uuid.UUID, vote *domain.Vote) (domain.Answer, error) {
	if err := validateVote(vote); err != nil {
		return domain.Answer{}, err
	}
	a, err := s.repo.GetAnswer(ctx, answerID)
	if err != nil {
		return domain.Answer{}, err
	}
	if a.AuthorID == userID {
		return domain.Answer{}, apperr.Validation(apperr.Field{
			Field: "vote", Code: "FORBIDDEN", Message: "you can’t vote on your own answer",
		})
	}
	return s.repo.SetAnswerVote(ctx, answerID, userID, vote)
}

func validateVote(vote *domain.Vote) error {
	if vote == nil {
		return nil
	}
	switch *vote {
	case domain.VoteUp, domain.VoteDown:
		return nil
	default:
		return apperr.Validation(apperr.Field{
			Field: "vote", Code: "INVALID", Message: "vote must be up, down, or null",
		})
	}
}

func (s *Service) listParams(limit int, cursor string) (domain.ListParams, error) {
	if limit <= 0 {
		limit = s.defaultPageSize
	}
	if limit > s.maxPageSize {
		limit = s.maxPageSize
	}
	p := domain.ListParams{Limit: limit}
	if strings.TrimSpace(cursor) == "" {
		return p, nil
	}
	c, err := decodeCursor(cursor)
	if err != nil {
		return domain.ListParams{}, err
	}
	p.Cursor = &c
	return p, nil
}

func encodeCursor(c domain.PageCursor) string {
	raw := fmt.Sprintf("%d|%s", c.CreatedAt.UTC().UnixNano(), c.ID.String())
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(raw string) (domain.PageCursor, error) {
	b, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(raw))
	if err != nil {
		return domain.PageCursor{}, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	parts := strings.Split(string(b), "|")
	if len(parts) != 2 {
		return domain.PageCursor{}, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	ns, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return domain.PageCursor{}, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return domain.PageCursor{}, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	return domain.PageCursor{CreatedAt: time.Unix(0, ns).UTC(), ID: id}, nil
}
