package moderation

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainModeration "github.com/beka-birhanu/yetbota/content-service/internal/domain/moderation"
)

const (
	tableReports          = "reports"
	tableModerationCases  = "moderation_cases"
	tableModerationAction = "moderation_actions"
)

const caseColumns = "id, content_type, content_id, report_count, status, severity, auto_hidden, first_reported_at, last_reported_at, resolved_by, resolved_at, resolution, version"

type repository struct {
	db *sql.DB
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

func NewRepo(cfg *Config) (domainModeration.Repository, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: cfg.DB}, nil
}

func (r *repository) exec(tx *sql.Tx) boil.ContextExecutor {
	if tx != nil {
		return tx
	}
	return r.db
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCase(s rowScanner) (*domainModeration.ModerationCase, error) {
	var (
		c          domainModeration.ModerationCase
		resolvedBy sql.NullString
		resolvedAt sql.NullTime
		resolution sql.NullString
	)
	if err := s.Scan(
		&c.ID, &c.ContentType, &c.ContentID, &c.ReportCount, &c.Status, &c.Severity,
		&c.AutoHidden, &c.FirstReportedAt, &c.LastReportedAt,
		&resolvedBy, &resolvedAt, &resolution, &c.Version,
	); err != nil {
		return nil, err
	}
	c.ResolvedBy = resolvedBy.String
	if resolvedAt.Valid {
		c.ResolvedAt = &resolvedAt.Time
	}
	c.Resolution = resolution.String
	return &c, nil
}

func (r *repository) CreateReport(ctx context.Context, tx *sql.Tx, rep *domainModeration.Report) (bool, error) {
	var details any
	if rep.Details != "" {
		details = rep.Details
	}
	var id string
	err := r.exec(tx).QueryRowContext(ctx,
		`INSERT INTO reports (id, content_type, content_id, reporter_id, reason, details, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, now())
		 ON CONFLICT (reporter_id, content_type, content_id) DO NOTHING
		 RETURNING id`,
		rep.ID, rep.ContentType, rep.ContentID, rep.ReporterID, rep.Reason, details,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, toddlerr.FromDBError(err, tableReports)
	}
	return true, nil
}

func (r *repository) UpsertCaseOnReport(ctx context.Context, tx *sql.Tx, id, contentType, contentID string, reportedAt time.Time) (*domainModeration.ModerationCase, error) {
	row := r.exec(tx).QueryRowContext(ctx,
		`INSERT INTO moderation_cases (id, content_type, content_id, report_count, status, severity, first_reported_at, last_reported_at)
		 VALUES ($1, $2, $3, 1, 'PENDING', 1, $4, $4)
		 ON CONFLICT (content_type, content_id) DO UPDATE
		 SET report_count = moderation_cases.report_count + 1,
		     severity = moderation_cases.report_count + 1,
		     last_reported_at = $4,
		     status = CASE WHEN moderation_cases.status = 'REJECTED' THEN 'PENDING' ELSE moderation_cases.status END
		 RETURNING `+caseColumns,
		id, contentType, contentID, reportedAt,
	)
	c, err := scanCase(row)
	if err != nil {
		return nil, toddlerr.FromDBError(err, tableModerationCases)
	}
	return c, nil
}

func (r *repository) MarkCaseAutoHidden(ctx context.Context, tx *sql.Tx, caseID string) error {
	_, err := r.exec(tx).ExecContext(ctx,
		`UPDATE moderation_cases SET auto_hidden = true WHERE id = $1`, caseID)
	if err != nil {
		return toddlerr.FromDBError(err, tableModerationCases)
	}
	return nil
}

func (r *repository) caseFilter(opts *domainModeration.CaseListOptions) (string, []any) {
	clauses := []string{}
	args := []any{}
	i := 1
	if opts.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = $%d", i))
		args = append(args, opts.Status)
		i++
	}
	if opts.ContentType != "" {
		clauses = append(clauses, fmt.Sprintf("content_type = $%d", i))
		args = append(args, opts.ContentType)
		i++
	}
	if opts.Reason != "" {
		clauses = append(clauses, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM reports rp WHERE rp.content_type = moderation_cases.content_type AND rp.content_id = moderation_cases.content_id AND rp.reason = $%d)", i))
		args = append(args, opts.Reason)
		i++
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (r *repository) ListCases(ctx context.Context, opts *domainModeration.CaseListOptions) ([]*domainModeration.ModerationCase, error) {
	where, args := r.caseFilter(opts)
	pageSize := opts.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	page := opts.Page
	if page <= 0 {
		page = 1
	}
	args = append(args, pageSize, (page-1)*pageSize)
	query := fmt.Sprintf(
		"SELECT %s FROM moderation_cases%s ORDER BY severity DESC, last_reported_at DESC LIMIT $%d OFFSET $%d",
		caseColumns, where, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, toddlerr.FromDBError(err, tableModerationCases)
	}
	defer rows.Close()

	var cases []*domainModeration.ModerationCase
	for rows.Next() {
		c, err := scanCase(rows)
		if err != nil {
			return nil, toddlerr.FromDBError(err, tableModerationCases)
		}
		cases = append(cases, c)
	}
	return cases, rows.Err()
}

func (r *repository) CountCases(ctx context.Context, opts *domainModeration.CaseListOptions) (int64, error) {
	where, args := r.caseFilter(opts)
	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT count(*) FROM moderation_cases"+where, args...).Scan(&count)
	if err != nil {
		return 0, toddlerr.FromDBError(err, tableModerationCases)
	}
	return count, nil
}

func (r *repository) GetCase(ctx context.Context, id string) (*domainModeration.ModerationCase, error) {
	row := r.db.QueryRowContext(ctx, "SELECT "+caseColumns+" FROM moderation_cases WHERE id = $1", id)
	c, err := scanCase(row)
	if err != nil {
		return nil, toddlerr.FromDBError(err, tableModerationCases)
	}
	return c, nil
}

func (r *repository) ListReports(ctx context.Context, contentType, contentID string) ([]*domainModeration.Report, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, content_type, content_id, reporter_id, reason, coalesce(details, ''), created_at
		 FROM reports WHERE content_type = $1 AND content_id = $2 ORDER BY created_at DESC`,
		contentType, contentID)
	if err != nil {
		return nil, toddlerr.FromDBError(err, tableReports)
	}
	defer rows.Close()

	var reports []*domainModeration.Report
	for rows.Next() {
		var rep domainModeration.Report
		if err := rows.Scan(&rep.ID, &rep.ContentType, &rep.ContentID, &rep.ReporterID, &rep.Reason, &rep.Details, &rep.CreatedAt); err != nil {
			return nil, toddlerr.FromDBError(err, tableReports)
		}
		reports = append(reports, &rep)
	}
	return reports, rows.Err()
}

func (r *repository) ResolveCase(ctx context.Context, tx *sql.Tx, caseID string, expectedVersion int, status, resolution, resolvedBy string) (bool, error) {
	res, err := r.exec(tx).ExecContext(ctx,
		`UPDATE moderation_cases
		 SET status = $1, resolution = $2, resolved_by = $3, resolved_at = now(), version = version + 1
		 WHERE id = $4 AND version = $5`,
		status, resolution, resolvedBy, caseID, expectedVersion)
	if err != nil {
		return false, toddlerr.FromDBError(err, tableModerationCases)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, toddlerr.FromDBError(err, tableModerationCases)
	}
	return affected > 0, nil
}

func (r *repository) CreateAction(ctx context.Context, tx *sql.Tx, a *domainModeration.ModerationAction) error {
	var note any
	if a.Note != "" {
		note = a.Note
	}
	_, err := r.exec(tx).ExecContext(ctx,
		`INSERT INTO moderation_actions (id, case_id, admin_id, action, target_type, target_id, note, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, now())`,
		a.ID, a.CaseID, a.AdminID, a.Action, a.TargetType, a.TargetID, note)
	if err != nil {
		return toddlerr.FromDBError(err, tableModerationAction)
	}
	return nil
}
