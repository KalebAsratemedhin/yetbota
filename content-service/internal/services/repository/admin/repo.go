package admin

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainAdmin "github.com/beka-birhanu/yetbota/content-service/internal/domain/admin"
)

const (
	tableUsers            = "users"
	tablePosts            = "posts"
	tableModerationAction = "moderation_actions"
)

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

func NewRepo(cfg *Config) (domainAdmin.Repository, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: cfg.DB}, nil
}

func (r *repository) CountUsers(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx, "SELECT count(*) FROM users").Scan(&n)
	if err != nil {
		return 0, toddlerr.FromDBError(err, tableUsers)
	}
	return n, nil
}

func (r *repository) CountQuestions(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx,
		"SELECT count(*) FROM posts WHERE is_question = true AND deleted_at IS NULL").Scan(&n)
	if err != nil {
		return 0, toddlerr.FromDBError(err, tablePosts)
	}
	return n, nil
}

func (r *repository) CountPostsWithLocation(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx,
		"SELECT count(*) FROM posts WHERE deleted_at IS NULL AND location IS NOT NULL").Scan(&n)
	if err != nil {
		return 0, toddlerr.FromDBError(err, tablePosts)
	}
	return n, nil
}

func (r *repository) UserPeriodCount(ctx context.Context, window time.Duration) (*domainAdmin.PeriodCount, error) {
	var pc domainAdmin.PeriodCount
	err := r.db.QueryRowContext(ctx,
		`SELECT
		   count(*) FILTER (WHERE created_at >= now() - $1::interval),
		   count(*) FILTER (WHERE created_at >= now() - ($1::interval * 2) AND created_at < now() - $1::interval)
		 FROM users`,
		intervalArg(window)).Scan(&pc.Current, &pc.Previous)
	if err != nil {
		return nil, toddlerr.FromDBError(err, tableUsers)
	}
	return &pc, nil
}

func (r *repository) CountUsersBefore(ctx context.Context, t time.Time) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx, "SELECT count(*) FROM users WHERE created_at < $1", t).Scan(&n)
	if err != nil {
		return 0, toddlerr.FromDBError(err, tableUsers)
	}
	return n, nil
}

func (r *repository) CountUsersSince(ctx context.Context, since time.Time) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx, "SELECT count(*) FROM users WHERE created_at >= $1", since).Scan(&n)
	if err != nil {
		return 0, toddlerr.FromDBError(err, tableUsers)
	}
	return n, nil
}

func (r *repository) HighReputationCount(ctx context.Context, threshold int64) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx, "SELECT count(*) FROM users WHERE score >= $1", threshold).Scan(&n)
	if err != nil {
		return 0, toddlerr.FromDBError(err, tableUsers)
	}
	return n, nil
}

func (r *repository) UserGrowthDaily(ctx context.Context, since time.Time) ([]domainAdmin.GrowthBucket, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT date_trunc('day', created_at) AS day, count(*)
		 FROM users WHERE created_at >= $1
		 GROUP BY day ORDER BY day`, since)
	if err != nil {
		return nil, toddlerr.FromDBError(err, tableUsers)
	}
	defer rows.Close()

	var buckets []domainAdmin.GrowthBucket
	for rows.Next() {
		var b domainAdmin.GrowthBucket
		if err := rows.Scan(&b.Day, &b.Count); err != nil {
			return nil, toddlerr.FromDBError(err, tableUsers)
		}
		buckets = append(buckets, b)
	}
	return buckets, rows.Err()
}

func (r *repository) auditFilter(opts *domainAdmin.AuditOptions) (string, []any) {
	clauses := []string{}
	args := []any{}
	i := 1
	if opts.ActionType != "" {
		clauses = append(clauses, fmt.Sprintf("ma.action = $%d", i))
		args = append(args, opts.ActionType)
		i++
	}
	if opts.Actor != "" {
		clauses = append(clauses, fmt.Sprintf("(u.username ILIKE $%d OR ma.admin_id::text ILIKE $%d)", i, i))
		args = append(args, "%"+opts.Actor+"%")
		i++
	}
	if opts.From != nil {
		clauses = append(clauses, fmt.Sprintf("ma.created_at >= $%d", i))
		args = append(args, *opts.From)
		i++
	}
	if opts.To != nil {
		clauses = append(clauses, fmt.Sprintf("ma.created_at <= $%d", i))
		args = append(args, *opts.To)
		i++
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (r *repository) ListAudit(ctx context.Context, opts *domainAdmin.AuditOptions) ([]*domainAdmin.AuditEntry, error) {
	where, args := r.auditFilter(opts)
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
		`SELECT ma.id, ma.created_at, ma.admin_id, coalesce(u.username, ''), ma.action, coalesce(ma.note, '')
		 FROM moderation_actions ma LEFT JOIN users u ON u.id = ma.admin_id%s
		 ORDER BY ma.created_at DESC LIMIT $%d OFFSET $%d`,
		where, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, toddlerr.FromDBError(err, tableModerationAction)
	}
	defer rows.Close()

	var entries []*domainAdmin.AuditEntry
	for rows.Next() {
		var e domainAdmin.AuditEntry
		if err := rows.Scan(&e.ID, &e.Timestamp, &e.ActorID, &e.ActorName, &e.ActionType, &e.Details); err != nil {
			return nil, toddlerr.FromDBError(err, tableModerationAction)
		}
		entries = append(entries, &e)
	}
	return entries, rows.Err()
}

func (r *repository) CountAudit(ctx context.Context, opts *domainAdmin.AuditOptions) (int64, error) {
	where, args := r.auditFilter(opts)
	var n int64
	err := r.db.QueryRowContext(ctx,
		"SELECT count(*) FROM moderation_actions ma LEFT JOIN users u ON u.id = ma.admin_id"+where, args...).Scan(&n)
	if err != nil {
		return 0, toddlerr.FromDBError(err, tableModerationAction)
	}
	return n, nil
}

func intervalArg(window time.Duration) string {
	secs := int64(window.Seconds())
	if secs <= 0 {
		secs = 1
	}
	return fmt.Sprintf("%d seconds", secs)
}
