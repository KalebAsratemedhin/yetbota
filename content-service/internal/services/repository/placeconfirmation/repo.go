package placeconfirmation

import (
	"context"
	"database/sql"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainConfirmation "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeconfirmation"
)

type repository struct {
	db *sql.DB
}

type dbExec interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (r *repository) executor(tx *sql.Tx) dbExec {
	if tx != nil {
		return tx
	}
	return r.db
}

type Config struct {
	DB *sql.DB `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewRepo(c *Config) (domainConfirmation.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: c.DB}, nil
}

func (r *repository) Upsert(ctx context.Context, tx *sql.Tx, entity *domainConfirmation.Confirmation) error {
	_, err := r.executor(tx).ExecContext(ctx, `
		INSERT INTO place_confirmations (id, place_id, user_id, note, confirmed_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (place_id, user_id) DO UPDATE
		SET note = EXCLUDED.note,
		    confirmed_at = EXCLUDED.confirmed_at
	`, entity.ID, entity.PlaceID, entity.UserID, entity.Note, entity.ConfirmedAt)
	if err != nil {
		return toddlerr.FromDBError(err, "place_confirmations")
	}
	return nil
}

func (r *repository) CountSince(ctx context.Context, placeID string, since time.Time) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(1) FROM place_confirmations
		WHERE place_id = $1 AND confirmed_at >= $2
	`, placeID, since).Scan(&count)
	if err != nil {
		return 0, toddlerr.FromDBError(err, "place_confirmations")
	}
	return count, nil
}
