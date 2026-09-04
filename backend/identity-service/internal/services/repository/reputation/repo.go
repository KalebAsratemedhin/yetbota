package reputation

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yetbota/identity-service/internal/db"
	"github.com/yetbota/identity-service/internal/domain/apperr"
	domain "github.com/yetbota/identity-service/internal/domain/reputation"
)

var _ domain.Repository = (*Repo)(nil)

type Repo struct {
	q *db.Queries
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{q: db.New(pool)}
}

func (r *Repo) AdjustScore(ctx context.Context, userID uuid.UUID, delta int64) (int64, error) {
	row, err := r.q.AddReputationScore(ctx, db.AddReputationScoreParams{
		ID: pgUUID(userID), ReputationScore: delta,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, apperr.ErrNotFound
		}
		return 0, apperr.Wrap(err, apperr.CodeInternal, "adjust reputation")
	}
	return row.ReputationScore, nil
}

func (r *Repo) SetScore(ctx context.Context, userID uuid.UUID, score int64) (int64, error) {
	if score < 0 {
		score = 0
	}
	row, err := r.q.SetReputationScore(ctx, db.SetReputationScoreParams{
		ID: pgUUID(userID), Score: score,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, apperr.ErrNotFound
		}
		return 0, apperr.Wrap(err, apperr.CodeInternal, "set reputation")
	}
	return row.ReputationScore, nil
}

func (r *Repo) InsertEvent(ctx context.Context, id uuid.UUID, userID uuid.UUID, actorID *uuid.UUID, eventType, subjectType string, subjectID *uuid.UUID, delta int32) error {
	_, err := r.q.InsertReputationEvent(ctx, db.InsertReputationEventParams{
		ID: pgUUID(id), UserID: pgUUID(userID), ActorID: optUUID(actorID),
		EventType: eventType, SubjectType: subjectType, SubjectID: optUUID(subjectID), Delta: delta,
	})
	if err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "insert reputation event")
	}
	return nil
}

func (r *Repo) SumPositiveDeltaToday(ctx context.Context, userID uuid.UUID) (int64, error) {
	n, err := r.q.SumReputationDeltaToday(ctx, pgUUID(userID))
	if err != nil {
		return 0, apperr.Wrap(err, apperr.CodeInternal, "sum daily reputation")
	}
	return n, nil
}

func (r *Repo) CountEventsByType(ctx context.Context, userID uuid.UUID, eventType string) (int64, error) {
	n, err := r.q.CountReputationEventsByType(ctx, db.CountReputationEventsByTypeParams{
		UserID: pgUUID(userID), EventType: eventType,
	})
	if err != nil {
		return 0, apperr.Wrap(err, apperr.CodeInternal, "count reputation events")
	}
	return n, nil
}

func (r *Repo) ListUserBadges(ctx context.Context, userID uuid.UUID) ([]domain.Badge, error) {
	rows, err := r.q.ListUserBadges(ctx, pgUUID(userID))
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternal, "list user badges")
	}
	out := make([]domain.Badge, 0, len(rows))
	for _, row := range rows {
		t := row.AwardedAt.Time
		out = append(out, domain.Badge{
			ID: row.ID, Name: row.Name, Description: row.Description, AwardedAt: &t,
		})
	}
	return out, nil
}

func (r *Repo) AwardBadge(ctx context.Context, userID uuid.UUID, badgeID string) (bool, error) {
	n, err := r.q.AwardBadge(ctx, db.AwardBadgeParams{UserID: pgUUID(userID), BadgeID: badgeID})
	if err != nil {
		return false, apperr.Wrap(err, apperr.CodeInternal, "award badge")
	}
	return n > 0, nil
}

func (r *Repo) HasBadge(ctx context.Context, userID uuid.UUID, badgeID string) (bool, error) {
	ok, err := r.q.HasBadge(ctx, db.HasBadgeParams{UserID: pgUUID(userID), BadgeID: badgeID})
	if err != nil {
		return false, apperr.Wrap(err, apperr.CodeInternal, "has badge")
	}
	return ok, nil
}

func (r *Repo) GetScore(ctx context.Context, userID uuid.UUID) (int64, error) {
	n, err := r.q.GetUserReputation(ctx, pgUUID(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, apperr.ErrNotFound
		}
		return 0, apperr.Wrap(err, apperr.CodeInternal, "get reputation")
	}
	return n, nil
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func optUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgUUID(*id)
}
