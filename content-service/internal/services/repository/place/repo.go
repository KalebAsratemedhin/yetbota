package place

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainPlace "github.com/beka-birhanu/yetbota/content-service/internal/domain/place"
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

func NewRepo(c *Config) (domainPlace.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: c.DB}, nil
}

const placeSelectCols = `
	id, name, description, address, location, created_by, status, duplicate_of,
	confidence_score, last_activity_at, last_confirmed_at, needs_check, summary_text,
	summary_generated_at, summary_citations, created_at, updated_at
`

func scanPlace(scanner interface{ Scan(dest ...any) error }) (*domainPlace.Place, error) {
	p := &domainPlace.Place{}
	if err := scanner.Scan(
		&p.ID, &p.Name, &p.Description, &p.Address, &p.Location, &p.CreatedBy, &p.Status, &p.DuplicateOf,
		&p.ConfidenceScore, &p.LastActivityAt, &p.LastConfirmedAt, &p.NeedsCheck, &p.SummaryText,
		&p.SummaryGeneratedAt, &p.SummaryCitations, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *repository) Add(ctx context.Context, tx *sql.Tx, entity *domainPlace.Place) error {
	citations := entity.SummaryCitations
	if len(citations) == 0 {
		citations = []byte("[]")
	}
	_, err := r.executor(tx).ExecContext(ctx, `
		INSERT INTO places (
			id, name, description, address, location, created_by, status, duplicate_of,
			confidence_score, last_activity_at, last_confirmed_at, needs_check, summary_text,
			summary_generated_at, summary_citations, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13,
			$14, $15, $16, $17
		)
	`, entity.ID, entity.Name, entity.Description, entity.Address, entity.Location, entity.CreatedBy, entity.Status, entity.DuplicateOf,
		entity.ConfidenceScore, entity.LastActivityAt, entity.LastConfirmedAt, entity.NeedsCheck, entity.SummaryText,
		entity.SummaryGeneratedAt, citations, entity.CreatedAt, entity.UpdatedAt)
	if err != nil {
		return toddlerr.FromDBError(err, "places")
	}
	return nil
}

func (r *repository) Read(ctx context.Context, id string) (*domainPlace.Place, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+placeSelectCols+` FROM places WHERE id = $1`, id)
	p, err := scanPlace(row)
	if err != nil {
		return nil, toddlerr.FromDBError(err, "places")
	}
	return p, nil
}

func (r *repository) Update(ctx context.Context, tx *sql.Tx, entity *domainPlace.Place) error {
	citations := entity.SummaryCitations
	if len(citations) == 0 {
		citations = []byte("[]")
	}
	res, err := r.executor(tx).ExecContext(ctx, `
		UPDATE places
		SET name = $1,
		    description = $2,
		    address = $3,
		    location = $4,
		    status = $5,
		    duplicate_of = $6,
		    confidence_score = $7,
		    last_activity_at = $8,
		    last_confirmed_at = $9,
		    needs_check = $10,
		    summary_text = $11,
		    summary_generated_at = $12,
		    summary_citations = $13,
		    updated_at = $14
		WHERE id = $15
	`, entity.Name, entity.Description, entity.Address, entity.Location, entity.Status, entity.DuplicateOf,
		entity.ConfidenceScore, entity.LastActivityAt, entity.LastConfirmedAt, entity.NeedsCheck, entity.SummaryText,
		entity.SummaryGeneratedAt, citations, entity.UpdatedAt, entity.ID)
	if err != nil {
		return toddlerr.FromDBError(err, "places")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "Place not found",
			ServiceMessage:    fmt.Sprintf("place not found id: %s", entity.ID),
		}
	}
	return nil
}

func (r *repository) List(ctx context.Context, opts *domainPlace.ListOptions) ([]*domainPlace.Place, error) {
	where, args, nextArg := buildWhereClause(opts, 1)
	page, pageSize := listPaging(opts)
	orderBy := "created_at DESC"
	if opts != nil && opts.NearLat != nil && opts.NearLon != nil {
		orderBy = fmt.Sprintf(
			"ST_Distance(location::geography, ST_SetSRID(ST_MakePoint(%f, %f), 4326)::geography) ASC",
			*opts.NearLon, *opts.NearLat,
		)
	}
	query := fmt.Sprintf(`
		SELECT %s
		FROM places
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, placeSelectCols, where, orderBy, nextArg, nextArg+1)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, toddlerr.FromDBError(err, "places")
	}
	defer rows.Close()

	out := make([]*domainPlace.Place, 0)
	for rows.Next() {
		p, err := scanPlace(rows)
		if err != nil {
			return nil, toddlerr.FromDBError(err, "places")
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, toddlerr.FromDBError(err, "places")
	}
	return out, nil
}

func (r *repository) Count(ctx context.Context, opts *domainPlace.ListOptions) (int64, error) {
	where, args, _ := buildWhereClause(opts, 1)
	query := fmt.Sprintf(`SELECT COUNT(1) FROM places %s`, where)
	var count int64
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, toddlerr.FromDBError(err, "places")
	}
	return count, nil
}

func (r *repository) UpdateStatus(ctx context.Context, tx *sql.Tx, id, modStatus string) error {
	_, err := r.executor(tx).ExecContext(ctx, `UPDATE places SET status = $1, updated_at = now() WHERE id = $2`, modStatus, id)
	if err != nil {
		return toddlerr.FromDBError(err, "places")
	}
	return nil
}

func (r *repository) RecordConfirmation(ctx context.Context, tx *sql.Tx, id string, confirmedAt time.Time, confidence int16) error {
	_, err := r.executor(tx).ExecContext(ctx, `
		UPDATE places
		SET last_confirmed_at = $1,
		    needs_check = false,
		    confidence_score = $2,
		    updated_at = now()
		WHERE id = $3
	`, confirmedAt, confidence, id)
	if err != nil {
		return toddlerr.FromDBError(err, "places")
	}
	return nil
}

func (r *repository) RecordContributionActivity(ctx context.Context, tx *sql.Tx, id string, at time.Time) error {
	_, err := r.executor(tx).ExecContext(ctx, `
		UPDATE places
		SET last_activity_at = $1,
		    needs_check = false,
		    updated_at = now()
		WHERE id = $2
	`, at, id)
	if err != nil {
		return toddlerr.FromDBError(err, "places")
	}
	return nil
}

func (r *repository) MarkStale(ctx context.Context, decayDays int) (int64, error) {
	if decayDays <= 0 {
		decayDays = 90
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE places
		SET needs_check = true,
		    updated_at = now()
		WHERE status = 'VISIBLE'
		  AND needs_check = false
		  AND COALESCE(last_confirmed_at, last_activity_at, created_at) < now() - ($1::text || ' days')::interval
	`, decayDays)
	if err != nil {
		return 0, toddlerr.FromDBError(err, "places")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return ra, nil
}

func (r *repository) UpdateSummary(ctx context.Context, tx *sql.Tx, id, summaryText string, generatedAt time.Time, citationsJSON []byte) error {
	if len(citationsJSON) == 0 {
		citationsJSON = []byte("[]")
	}
	res, err := r.executor(tx).ExecContext(ctx, `
		UPDATE places
		SET summary_text = $1,
		    summary_generated_at = $2,
		    summary_citations = $3,
		    updated_at = now()
		WHERE id = $4
	`, summaryText, generatedAt, citationsJSON, id)
	if err != nil {
		return toddlerr.FromDBError(err, "places")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "Place not found",
			ServiceMessage:    fmt.Sprintf("place not found id: %s", id),
		}
	}
	return nil
}
