package placeactivity

import (
	"context"
	"database/sql"
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainActivity "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeactivity"
	"github.com/lib/pq"
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

func NewRepo(c *Config) (domainActivity.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: c.DB}, nil
}

func (r *repository) Add(ctx context.Context, tx *sql.Tx, event *domainActivity.Event) error {
	_, err := r.executor(tx).ExecContext(ctx, `
		INSERT INTO place_activity_events (id, place_id, actor_id, event_type, ref_id, summary, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, event.ID, event.PlaceID, event.ActorID, event.EventType, event.RefID, event.Summary, event.CreatedAt)
	if err != nil {
		return toddlerr.FromDBError(err, "place_activity_events")
	}
	return nil
}

func (r *repository) List(ctx context.Context, opts *domainActivity.ListOptions) ([]*domainActivity.Event, error) {
	page, pageSize := 1, 20
	if opts != nil {
		if opts.Page > 0 {
			page = opts.Page
		}
		if opts.PageSize > 0 && opts.PageSize <= 100 {
			pageSize = opts.PageSize
		}
	}
	args := []any{}
	where := ""
	next := 1
	if opts != nil && len(opts.PlaceIDs) > 0 {
		where = fmt.Sprintf("WHERE place_id = ANY($%d)", next)
		args = append(args, pq.Array(opts.PlaceIDs))
		next++
	}
	query := fmt.Sprintf(`
		SELECT id, place_id, actor_id, event_type, ref_id, summary, created_at
		FROM place_activity_events
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, next, next+1)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, toddlerr.FromDBError(err, "place_activity_events")
	}
	defer rows.Close()

	out := make([]*domainActivity.Event, 0)
	for rows.Next() {
		e := &domainActivity.Event{}
		if err := rows.Scan(&e.ID, &e.PlaceID, &e.ActorID, &e.EventType, &e.RefID, &e.Summary, &e.CreatedAt); err != nil {
			return nil, toddlerr.FromDBError(err, "place_activity_events")
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *repository) Count(ctx context.Context, opts *domainActivity.ListOptions) (int64, error) {
	args := []any{}
	where := ""
	if opts != nil && len(opts.PlaceIDs) > 0 {
		where = "WHERE place_id = ANY($1)"
		args = append(args, pq.Array(opts.PlaceIDs))
	}
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM place_activity_events `+where, args...).Scan(&count)
	if err != nil {
		return 0, toddlerr.FromDBError(err, "place_activity_events")
	}
	return count, nil
}
