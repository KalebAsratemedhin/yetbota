package contributionvote

import (
	"context"
	"database/sql"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainContributionVote "github.com/beka-birhanu/yetbota/content-service/internal/domain/contributionvote"
	"github.com/google/uuid"
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

func NewRepo(c *Config) (domainContributionVote.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: c.DB}, nil
}

func (r *repository) Exists(ctx context.Context, userID, contributionID string) (bool, error) {
	var exists bool
	if err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM contribution_votes WHERE user_id = $1 AND contribution_id = $2
		)
	`, userID, contributionID).Scan(&exists); err != nil {
		return false, toddlerr.FromDBError(err, "contribution_votes")
	}
	return exists, nil
}

func (r *repository) Read(ctx context.Context, userID, contributionID string) (*domainContributionVote.Vote, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, contribution_id, user_id, vote_type, created_at
		FROM contribution_votes
		WHERE user_id = $1 AND contribution_id = $2
	`, userID, contributionID)
	v := &domainContributionVote.Vote{}
	if err := row.Scan(&v.ID, &v.ContributionID, &v.UserID, &v.VoteType, &v.CreatedAt); err != nil {
		return nil, toddlerr.FromDBError(err, "contribution_votes")
	}
	return v, nil
}

func (r *repository) Add(ctx context.Context, tx *sql.Tx, entity *domainContributionVote.Vote) error {
	if entity.ID == "" {
		entity.ID = uuid.NewString()
	}
	_, err := r.executor(tx).ExecContext(ctx, `
		INSERT INTO contribution_votes (id, contribution_id, user_id, vote_type, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, entity.ID, entity.ContributionID, entity.UserID, entity.VoteType, entity.CreatedAt)
	if err != nil {
		return toddlerr.FromDBError(err, "contribution_votes")
	}
	return nil
}

func (r *repository) Update(ctx context.Context, tx *sql.Tx, entity *domainContributionVote.Vote) error {
	res, err := r.executor(tx).ExecContext(ctx, `
		UPDATE contribution_votes
		SET vote_type = $1
		WHERE contribution_id = $2 AND user_id = $3
	`, entity.VoteType, entity.ContributionID, entity.UserID)
	if err != nil {
		return toddlerr.FromDBError(err, "contribution_votes")
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "Contribution vote not found",
			ServiceMessage:    "contribution vote not found",
		}
	}
	return nil
}

func (r *repository) List(ctx context.Context, opts *domainContributionVote.ListOptions) (map[string]*domainContributionVote.Vote, error) {
	if opts == nil || opts.UserID == "" || len(opts.ContributionIDs) == 0 {
		return nil, nil
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, contribution_id, user_id, vote_type, created_at
		FROM contribution_votes
		WHERE user_id = $1
		  AND contribution_id = ANY($2)
	`, opts.UserID, pq.Array(opts.ContributionIDs))
	if err != nil {
		return nil, toddlerr.FromDBError(err, "contribution_votes")
	}
	defer rows.Close()

	result := make(map[string]*domainContributionVote.Vote, len(opts.ContributionIDs))
	for rows.Next() {
		v := &domainContributionVote.Vote{}
		if err := rows.Scan(&v.ID, &v.ContributionID, &v.UserID, &v.VoteType, &v.CreatedAt); err != nil {
			return nil, toddlerr.FromDBError(err, "contribution_votes")
		}
		result[v.ContributionID] = v
	}
	if err := rows.Err(); err != nil {
		return nil, toddlerr.FromDBError(err, "contribution_votes")
	}
	return result, nil
}
