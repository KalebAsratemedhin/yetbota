package qa

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yetbota/content-service/internal/db"
	"github.com/yetbota/content-service/internal/domain/apperr"
	domain "github.com/yetbota/content-service/internal/domain/qa"
)

var _ domain.Repository = (*Repo)(nil)

type Repo struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool, q: db.New(pool)}
}

func (r *Repo) PostExists(ctx context.Context, postID uuid.UUID) (bool, error) {
	_, err := r.q.GetPostByID(ctx, pgUUID(postID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, apperr.Wrap(err, apperr.CodeInternal, "get post")
	}
	return true, nil
}

func (r *Repo) CreateQuestion(ctx context.Context, in domain.CreateQuestionInput) (domain.Question, error) {
	params := db.CreateQuestionParams{
		ID: pgUUID(uuid.New()), AuthorID: pgUUID(in.AuthorID),
		Title: in.Title, Body: in.Body,
	}
	if in.PostID != nil {
		params.PostID = pgUUID(*in.PostID)
	}
	if in.Lat != nil && in.Lng != nil {
		params.Lat = pgtype.Float8{Float64: *in.Lat, Valid: true}
		params.Lng = pgtype.Float8{Float64: *in.Lng, Valid: true}
	}
	row, err := r.q.CreateQuestion(ctx, params)
	if err != nil {
		return domain.Question{}, apperr.Wrap(err, apperr.CodeInternal, "create question")
	}
	return questionFromDB(row)
}

func (r *Repo) GetQuestion(ctx context.Context, id uuid.UUID) (domain.Question, error) {
	row, err := r.q.GetQuestionByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Question{}, apperr.ErrNotFound
		}
		return domain.Question{}, apperr.Wrap(err, apperr.CodeInternal, "get question")
	}
	return questionFromDB(row)
}

func (r *Repo) ListQuestions(ctx context.Context, p domain.ListParams) (domain.QuestionPage, error) {
	params := db.ListQuestionsParams{
		CursorCreatedAt: cursorTime(p.Cursor),
		CursorID:        cursorID(p.Cursor),
		ResultLimit:     int32(p.Limit + 1),
	}
	if p.PostID != nil {
		params.PostID = pgUUID(*p.PostID)
	}
	rows, err := r.q.ListQuestions(ctx, params)
	if err != nil {
		return domain.QuestionPage{}, apperr.Wrap(err, apperr.CodeInternal, "list questions")
	}
	hasMore := len(rows) > p.Limit
	if hasMore {
		rows = rows[:p.Limit]
	}
	out := make([]domain.Question, 0, len(rows))
	for _, row := range rows {
		q, err := questionFromDB(row)
		if err != nil {
			return domain.QuestionPage{}, err
		}
		out = append(out, q)
	}
	page := domain.QuestionPage{Questions: out}
	if hasMore && len(out) > 0 {
		last := out[len(out)-1]
		page.NextCursor = &domain.PageCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	return page, nil
}

func (r *Repo) CloseQuestion(ctx context.Context, id, closedBy uuid.UUID) (domain.Question, error) {
	row, err := r.q.CloseQuestion(ctx, db.CloseQuestionParams{
		ID: pgUUID(id), ClosedBy: pgUUID(closedBy),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// already closed or missing — distinguish
			q, gerr := r.GetQuestion(ctx, id)
			if gerr != nil {
				return domain.Question{}, gerr
			}
			if q.Closed() {
				return domain.Question{}, apperr.Validation(apperr.Field{
					Field: "id", Code: "CLOSED", Message: "question is already closed",
				})
			}
			return domain.Question{}, apperr.ErrNotFound
		}
		return domain.Question{}, apperr.Wrap(err, apperr.CodeInternal, "close question")
	}
	return questionFromDB(row)
}

func (r *Repo) SetQuestionVote(ctx context.Context, questionID, userID uuid.UUID, vote *domain.Vote) (domain.Question, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Question{}, apperr.Wrap(err, apperr.CodeInternal, "begin question vote tx")
	}
	defer tx.Rollback(ctx)
	q := r.q.WithTx(tx)

	if _, err := q.GetQuestionByID(ctx, pgUUID(questionID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Question{}, apperr.ErrNotFound
		}
		return domain.Question{}, apperr.Wrap(err, apperr.CodeInternal, "get question")
	}

	prev, err := q.GetQuestionVote(ctx, db.GetQuestionVoteParams{
		QuestionID: pgUUID(questionID), UserID: pgUUID(userID),
	})
	hasPrev := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return domain.Question{}, apperr.Wrap(err, apperr.CodeInternal, "get question vote")
	}

	upDelta, downDelta := int64(0), int64(0)
	if hasPrev {
		switch prev.Vote {
		case string(domain.VoteUp):
			upDelta--
		case string(domain.VoteDown):
			downDelta--
		}
		if err := q.DeleteQuestionVote(ctx, db.DeleteQuestionVoteParams{
			QuestionID: pgUUID(questionID), UserID: pgUUID(userID),
		}); err != nil {
			return domain.Question{}, apperr.Wrap(err, apperr.CodeInternal, "delete question vote")
		}
	}
	if vote != nil {
		switch *vote {
		case domain.VoteUp:
			upDelta++
		case domain.VoteDown:
			downDelta++
		}
		if _, err := q.UpsertQuestionVote(ctx, db.UpsertQuestionVoteParams{
			QuestionID: pgUUID(questionID), UserID: pgUUID(userID), Vote: string(*vote),
		}); err != nil {
			return domain.Question{}, apperr.Wrap(err, apperr.CodeInternal, "upsert question vote")
		}
	}
	if upDelta != 0 {
		if _, err := q.AdjustQuestionUpvoteCount(ctx, db.AdjustQuestionUpvoteCountParams{
			ID: pgUUID(questionID), Delta: upDelta,
		}); err != nil {
			return domain.Question{}, apperr.Wrap(err, apperr.CodeInternal, "adjust question upvotes")
		}
	}
	if downDelta != 0 {
		if _, err := q.AdjustQuestionDownvoteCount(ctx, db.AdjustQuestionDownvoteCountParams{
			ID: pgUUID(questionID), Delta: downDelta,
		}); err != nil {
			return domain.Question{}, apperr.Wrap(err, apperr.CodeInternal, "adjust question downvotes")
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Question{}, apperr.Wrap(err, apperr.CodeInternal, "commit question vote")
	}
	return r.GetQuestion(ctx, questionID)
}

func (r *Repo) CreateAnswer(ctx context.Context, in domain.CreateAnswerInput) (domain.Answer, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Answer{}, apperr.Wrap(err, apperr.CodeInternal, "begin answer tx")
	}
	defer tx.Rollback(ctx)
	q := r.q.WithTx(tx)

	qr, err := q.GetQuestionByID(ctx, pgUUID(in.QuestionID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Answer{}, apperr.ErrNotFound
		}
		return domain.Answer{}, apperr.Wrap(err, apperr.CodeInternal, "get question")
	}
	if qr.ClosedAt.Valid {
		return domain.Answer{}, apperr.Validation(apperr.Field{
			Field: "question_id", Code: "CLOSED", Message: "cannot answer a closed question",
		})
	}

	row, err := q.CreateAnswer(ctx, db.CreateAnswerParams{
		ID: pgUUID(uuid.New()), QuestionID: pgUUID(in.QuestionID),
		AuthorID: pgUUID(in.AuthorID), Body: in.Body,
	})
	if err != nil {
		return domain.Answer{}, apperr.Wrap(err, apperr.CodeInternal, "create answer")
	}
	if _, err := q.AdjustQuestionAnswerCount(ctx, db.AdjustQuestionAnswerCountParams{
		ID: pgUUID(in.QuestionID), Delta: 1,
	}); err != nil {
		return domain.Answer{}, apperr.Wrap(err, apperr.CodeInternal, "adjust answer count")
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Answer{}, apperr.Wrap(err, apperr.CodeInternal, "commit answer")
	}
	return answerFromDB(row)
}

func (r *Repo) GetAnswer(ctx context.Context, id uuid.UUID) (domain.Answer, error) {
	row, err := r.q.GetAnswerByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Answer{}, apperr.ErrNotFound
		}
		return domain.Answer{}, apperr.Wrap(err, apperr.CodeInternal, "get answer")
	}
	return answerFromDB(row)
}

func (r *Repo) ListAnswers(ctx context.Context, questionID uuid.UUID, p domain.ListParams) (domain.AnswerPage, error) {
	if _, err := r.GetQuestion(ctx, questionID); err != nil {
		return domain.AnswerPage{}, err
	}
	rows, err := r.q.ListAnswersByQuestion(ctx, db.ListAnswersByQuestionParams{
		QuestionID:      pgUUID(questionID),
		CursorCreatedAt: cursorTime(p.Cursor),
		CursorID:        cursorID(p.Cursor),
		ResultLimit:     int32(p.Limit + 1),
	})
	if err != nil {
		return domain.AnswerPage{}, apperr.Wrap(err, apperr.CodeInternal, "list answers")
	}
	hasMore := len(rows) > p.Limit
	if hasMore {
		rows = rows[:p.Limit]
	}
	out := make([]domain.Answer, 0, len(rows))
	for _, row := range rows {
		a, err := answerFromDB(row)
		if err != nil {
			return domain.AnswerPage{}, err
		}
		out = append(out, a)
	}
	page := domain.AnswerPage{Answers: out}
	if hasMore && len(out) > 0 {
		last := out[len(out)-1]
		page.NextCursor = &domain.PageCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	return page, nil
}

func (r *Repo) SetAnswerVote(ctx context.Context, answerID, userID uuid.UUID, vote *domain.Vote) (domain.Answer, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Answer{}, apperr.Wrap(err, apperr.CodeInternal, "begin answer vote tx")
	}
	defer tx.Rollback(ctx)
	q := r.q.WithTx(tx)

	if _, err := q.GetAnswerByID(ctx, pgUUID(answerID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Answer{}, apperr.ErrNotFound
		}
		return domain.Answer{}, apperr.Wrap(err, apperr.CodeInternal, "get answer")
	}

	prev, err := q.GetAnswerVote(ctx, db.GetAnswerVoteParams{
		AnswerID: pgUUID(answerID), UserID: pgUUID(userID),
	})
	hasPrev := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return domain.Answer{}, apperr.Wrap(err, apperr.CodeInternal, "get answer vote")
	}

	upDelta, downDelta := int64(0), int64(0)
	if hasPrev {
		switch prev.Vote {
		case string(domain.VoteUp):
			upDelta--
		case string(domain.VoteDown):
			downDelta--
		}
		if err := q.DeleteAnswerVote(ctx, db.DeleteAnswerVoteParams{
			AnswerID: pgUUID(answerID), UserID: pgUUID(userID),
		}); err != nil {
			return domain.Answer{}, apperr.Wrap(err, apperr.CodeInternal, "delete answer vote")
		}
	}
	if vote != nil {
		switch *vote {
		case domain.VoteUp:
			upDelta++
		case domain.VoteDown:
			downDelta++
		}
		if _, err := q.UpsertAnswerVote(ctx, db.UpsertAnswerVoteParams{
			AnswerID: pgUUID(answerID), UserID: pgUUID(userID), Vote: string(*vote),
		}); err != nil {
			return domain.Answer{}, apperr.Wrap(err, apperr.CodeInternal, "upsert answer vote")
		}
	}
	if upDelta != 0 {
		if _, err := q.AdjustAnswerUpvoteCount(ctx, db.AdjustAnswerUpvoteCountParams{
			ID: pgUUID(answerID), Delta: upDelta,
		}); err != nil {
			return domain.Answer{}, apperr.Wrap(err, apperr.CodeInternal, "adjust answer upvotes")
		}
	}
	if downDelta != 0 {
		if _, err := q.AdjustAnswerDownvoteCount(ctx, db.AdjustAnswerDownvoteCountParams{
			ID: pgUUID(answerID), Delta: downDelta,
		}); err != nil {
			return domain.Answer{}, apperr.Wrap(err, apperr.CodeInternal, "adjust answer downvotes")
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Answer{}, apperr.Wrap(err, apperr.CodeInternal, "commit answer vote")
	}
	return r.GetAnswer(ctx, answerID)
}

func questionFromDB(row db.Question) (domain.Question, error) {
	id, err := uuidFromPg(row.ID)
	if err != nil {
		return domain.Question{}, err
	}
	aid, err := uuidFromPg(row.AuthorID)
	if err != nil {
		return domain.Question{}, err
	}
	q := domain.Question{
		ID: id, AuthorID: aid, Title: row.Title, Body: row.Body,
		UpvoteCount: row.UpvoteCount, DownvoteCount: row.DownvoteCount,
		AnswerCount: row.AnswerCount, CreatedAt: row.CreatedAt.Time,
	}
	if row.PostID.Valid {
		pid, err := uuidFromPg(row.PostID)
		if err != nil {
			return domain.Question{}, err
		}
		q.PostID = &pid
	}
	if row.Lat.Valid && row.Lng.Valid {
		lat, lng := row.Lat.Float64, row.Lng.Float64
		q.Lat, q.Lng = &lat, &lng
	}
	if row.ClosedAt.Valid {
		t := row.ClosedAt.Time
		q.ClosedAt = &t
	}
	if row.ClosedBy.Valid {
		cb, err := uuidFromPg(row.ClosedBy)
		if err != nil {
			return domain.Question{}, err
		}
		q.ClosedBy = &cb
	}
	return q, nil
}

func answerFromDB(row db.Answer) (domain.Answer, error) {
	id, err := uuidFromPg(row.ID)
	if err != nil {
		return domain.Answer{}, err
	}
	qid, err := uuidFromPg(row.QuestionID)
	if err != nil {
		return domain.Answer{}, err
	}
	aid, err := uuidFromPg(row.AuthorID)
	if err != nil {
		return domain.Answer{}, err
	}
	return domain.Answer{
		ID: id, QuestionID: qid, AuthorID: aid, Body: row.Body,
		UpvoteCount: row.UpvoteCount, DownvoteCount: row.DownvoteCount,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func uuidFromPg(v pgtype.UUID) (uuid.UUID, error) {
	if !v.Valid {
		return uuid.Nil, apperr.New(apperr.CodeInternal, "invalid uuid")
	}
	return uuid.UUID(v.Bytes), nil
}

func cursorTime(c *domain.PageCursor) pgtype.Timestamptz {
	if c == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: c.CreatedAt, Valid: true}
}

func cursorID(c *domain.PageCursor) pgtype.UUID {
	if c == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgUUID(c.ID)
}
