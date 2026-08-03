package placebookmark

import (
	"context"
	"database/sql"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainBookmark "github.com/beka-birhanu/yetbota/content-service/internal/domain/placebookmark"
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

func NewRepo(c *Config) (domainBookmark.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: c.DB}, nil
}

func (r *repository) Add(ctx context.Context, tx *sql.Tx, userID, placeID string) error {
	_, err := r.executor(tx).ExecContext(ctx, `
		INSERT INTO place_bookmarks (user_id, place_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, place_id) DO NOTHING
	`, userID, placeID)
	if err != nil {
		return toddlerr.FromDBError(err, "place_bookmarks")
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, tx *sql.Tx, userID, placeID string) error {
	_, err := r.executor(tx).ExecContext(ctx, `
		DELETE FROM place_bookmarks WHERE user_id = $1 AND place_id = $2
	`, userID, placeID)
	if err != nil {
		return toddlerr.FromDBError(err, "place_bookmarks")
	}
	return nil
}

func (r *repository) Exists(ctx context.Context, userID, placeID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM place_bookmarks WHERE user_id = $1 AND place_id = $2
		)
	`, userID, placeID).Scan(&exists)
	if err != nil {
		return false, toddlerr.FromDBError(err, "place_bookmarks")
	}
	return exists, nil
}

func (r *repository) ListPlaceIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT place_id FROM place_bookmarks WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, toddlerr.FromDBError(err, "place_bookmarks")
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, toddlerr.FromDBError(err, "place_bookmarks")
		}
		out = append(out, id)
	}
	return out, rows.Err()
}
