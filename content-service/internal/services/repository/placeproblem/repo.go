package placeproblem

import (
	"context"
	"database/sql"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainProblem "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeproblem"
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

func NewRepo(c *Config) (domainProblem.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: c.DB}, nil
}

func (r *repository) Add(ctx context.Context, tx *sql.Tx, entity *domainProblem.Report) error {
	_, err := r.executor(tx).ExecContext(ctx, `
		INSERT INTO place_problem_reports (
			id, place_id, user_id, kind, note, duplicate_of, status, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, entity.ID, entity.PlaceID, entity.UserID, entity.Kind, entity.Note, entity.DuplicateOf,
		entity.Status, entity.CreatedAt, entity.UpdatedAt)
	if err != nil {
		return toddlerr.FromDBError(err, "place_problem_reports")
	}
	return nil
}
