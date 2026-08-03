package contribution

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainContribution "github.com/beka-birhanu/yetbota/content-service/internal/domain/contribution"
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

func NewRepo(c *Config) (domainContribution.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: c.DB}, nil
}

const contributionSelectCols = `
	c.id, c.place_id, c.user_id, c.kind, c.content, c.parent_id, c.upvotes, c.downvotes, c.moderation_status,
	c.origin, c.source_post_id, c.accepted_at, c.accepted_by, c.created_at, c.updated_at
`

func scanContribution(scanner interface{ Scan(dest ...any) error }) (*domainContribution.Contribution, error) {
	c := &domainContribution.Contribution{}
	if err := scanner.Scan(
		&c.ID, &c.PlaceID, &c.UserID, &c.Kind, &c.Content, &c.ParentID, &c.Upvotes, &c.Downvotes, &c.ModerationStatus,
		&c.Origin, &c.SourcePostID, &c.AcceptedAt, &c.AcceptedBy, &c.CreatedAt, &c.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return c, nil
}

func (r *repository) Add(ctx context.Context, tx *sql.Tx, entity *domainContribution.Contribution) error {
	_, err := r.executor(tx).ExecContext(ctx, `
		INSERT INTO place_contributions (
			id, place_id, user_id, kind, content, parent_id, upvotes, downvotes,
			moderation_status, origin, source_post_id, accepted_at, accepted_by, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
	`, entity.ID, entity.PlaceID, entity.UserID, entity.Kind, entity.Content, entity.ParentID, entity.Upvotes, entity.Downvotes,
		entity.ModerationStatus, entity.Origin, entity.SourcePostID, entity.AcceptedAt, entity.AcceptedBy, entity.CreatedAt, entity.UpdatedAt)
	if err != nil {
		return toddlerr.FromDBError(err, "place_contributions")
	}
	return nil
}

func (r *repository) Read(ctx context.Context, id string) (*domainContribution.Contribution, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+contributionSelectCols+` FROM place_contributions c WHERE c.id = $1`, id)
	c, err := scanContribution(row)
	if err != nil {
		return nil, toddlerr.FromDBError(err, "place_contributions")
	}
	return c, nil
}

func (r *repository) List(ctx context.Context, opts *domainContribution.ListOptions) ([]*domainContribution.Contribution, error) {
	where, args, next, from := buildWhere(opts, 1)
	page, pageSize := pageInfo(opts)
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		%s
		ORDER BY c.created_at DESC
		LIMIT $%d OFFSET $%d
	`, contributionSelectCols, from, where, next, next+1)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, toddlerr.FromDBError(err, "place_contributions")
	}
	defer rows.Close()

	out := make([]*domainContribution.Contribution, 0)
	for rows.Next() {
		c, err := scanContribution(rows)
		if err != nil {
			return nil, toddlerr.FromDBError(err, "place_contributions")
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, toddlerr.FromDBError(err, "place_contributions")
	}
	return out, nil
}

func (r *repository) Count(ctx context.Context, opts *domainContribution.ListOptions) (int64, error) {
	where, args, _, from := buildWhere(opts, 1)
	query := fmt.Sprintf(`SELECT COUNT(1) FROM %s %s`, from, where)
	var count int64
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, toddlerr.FromDBError(err, "place_contributions")
	}
	return count, nil
}

func (r *repository) Update(ctx context.Context, tx *sql.Tx, entity *domainContribution.Contribution) error {
	res, err := r.executor(tx).ExecContext(ctx, `
		UPDATE place_contributions
		SET content = $1, moderation_status = $2, updated_at = $3
		WHERE id = $4
	`, entity.Content, entity.ModerationStatus, entity.UpdatedAt, entity.ID)
	if err != nil {
		return toddlerr.FromDBError(err, "place_contributions")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "Contribution not found",
			ServiceMessage:    fmt.Sprintf("contribution not found id: %s", entity.ID),
		}
	}
	return nil
}

func (r *repository) UpdateVoteCounts(ctx context.Context, tx *sql.Tx, id string, upDelta, downDelta, expectedUp, expectedDown int) error {
	res, err := r.executor(tx).ExecContext(ctx, `
		UPDATE place_contributions
		SET upvotes = upvotes + $1,
		    downvotes = downvotes + $2,
		    updated_at = now()
		WHERE id = $3
		  AND upvotes = $4
		  AND downvotes = $5
	`, upDelta, downDelta, id, expectedUp, expectedDown)
	if err != nil {
		return toddlerr.FromDBError(err, "place_contributions")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.Conflict,
			ServiceStatusCode: status.Conflict,
			PublicMessage:     "Please try again later.",
			ServiceMessage:    "contribution vote optimistic lock failed",
		}
	}
	return nil
}

func (r *repository) UpdateModerationStatus(ctx context.Context, tx *sql.Tx, id, modStatus string) error {
	_, err := r.executor(tx).ExecContext(ctx, `
		UPDATE place_contributions
		SET moderation_status = $1, updated_at = now()
		WHERE id = $2
	`, modStatus, id)
	if err != nil {
		return toddlerr.FromDBError(err, "place_contributions")
	}
	return nil
}

func (r *repository) ClearAcceptedAnswer(ctx context.Context, tx *sql.Tx, questionID string) error {
	_, err := r.executor(tx).ExecContext(ctx, `
		UPDATE place_contributions
		SET accepted_at = NULL, accepted_by = NULL, updated_at = now()
		WHERE parent_id = $1 AND kind = 'answer' AND accepted_at IS NOT NULL
	`, questionID)
	if err != nil {
		return toddlerr.FromDBError(err, "place_contributions")
	}
	return nil
}

func (r *repository) AcceptAnswer(ctx context.Context, tx *sql.Tx, answerID, acceptedBy string, acceptedAt time.Time) error {
	res, err := r.executor(tx).ExecContext(ctx, `
		UPDATE place_contributions
		SET accepted_at = $1, accepted_by = $2, updated_at = now()
		WHERE id = $3 AND kind = 'answer'
	`, acceptedAt, acceptedBy, answerID)
	if err != nil {
		return toddlerr.FromDBError(err, "place_contributions")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "Answer not found",
			ServiceMessage:    "answer not found for accept",
		}
	}
	return nil
}

func (r *repository) ExistsForSourcePost(ctx context.Context, sourcePostID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM place_contributions WHERE source_post_id = $1)
	`, sourcePostID).Scan(&exists)
	if err != nil {
		return false, toddlerr.FromDBError(err, "place_contributions")
	}
	return exists, nil
}

func buildWhere(opts *domainContribution.ListOptions, start int) (string, []any, int, string) {
	from := "place_contributions c"
	if opts == nil {
		return "", nil, start, from
	}
	clauses := make([]string, 0, 8)
	args := make([]any, 0, 8)
	next := start
	needsPlaceJoin := opts.NearLat != nil && opts.NearLon != nil && opts.RadiusKm != nil
	if needsPlaceJoin {
		from = "place_contributions c JOIN places p ON p.id = c.place_id"
	}

	if opts.PlaceID != "" {
		clauses = append(clauses, fmt.Sprintf("c.place_id = $%d", next))
		args = append(args, opts.PlaceID)
		next++
	}
	if opts.Kind != "" {
		clauses = append(clauses, fmt.Sprintf("c.kind = $%d", next))
		args = append(args, opts.Kind)
		next++
	}
	if opts.ParentID != nil {
		if *opts.ParentID == "" {
			clauses = append(clauses, "c.parent_id IS NULL")
		} else {
			clauses = append(clauses, fmt.Sprintf("c.parent_id = $%d", next))
			args = append(args, *opts.ParentID)
			next++
		}
	}
	if opts.OnlyVisible {
		clauses = append(clauses, fmt.Sprintf("c.moderation_status = $%d", next))
		args = append(args, constants.ModerationStatusVisible)
		next++
	}
	if opts.NonAnswerTop {
		clauses = append(clauses, "(c.parent_id IS NULL OR c.kind = 'question')")
	}
	if opts.SourcePostID != "" {
		clauses = append(clauses, fmt.Sprintf("c.source_post_id = $%d", next))
		args = append(args, opts.SourcePostID)
		next++
	}
	if opts.OpenOnly {
		clauses = append(clauses, `c.kind = 'question'`)
		clauses = append(clauses, `
			NOT EXISTS (
				SELECT 1 FROM place_contributions a
				WHERE a.parent_id = c.id AND a.kind = 'answer' AND a.accepted_at IS NOT NULL
			)`)
	}
	if needsPlaceJoin {
		clauses = append(clauses, fmt.Sprintf(
			"p.location IS NOT NULL AND ST_DWithin(p.location::geography, ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography, $%d)",
			next, next+1, next+2,
		))
		args = append(args, *opts.NearLon, *opts.NearLat, *opts.RadiusKm*1000)
		next += 3
	}

	if len(clauses) == 0 {
		return "", args, next, from
	}
	return "WHERE " + strings.Join(clauses, " AND "), args, next, from
}

func pageInfo(opts *domainContribution.ListOptions) (int, int) {
	page := 1
	pageSize := 20
	if opts == nil {
		return page, pageSize
	}
	if opts.Page > 0 {
		page = opts.Page
	}
	if opts.PageSize > 0 && opts.PageSize <= 100 {
		pageSize = opts.PageSize
	}
	return page, pageSize
}
