package postvote

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/postvote"
	"github.com/lib/pq"
)

type repository struct {
	db *sql.DB
}

type Config struct {
	DB *sql.DB `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}
	return nil
}

func NewRepo(c *Config) (postvote.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: c.DB}, nil
}

// Exists implements [postvote.Repository].
func (r *repository) Exists(ctx context.Context, userID string, postID string) (bool, error) {
	exists, err := dbmodels.PostVoteExists(ctx, r.db, userID, postID)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Add implements [postvote.Repository].
func (r *repository) Add(ctx context.Context, tx *sql.Tx, entity *dbmodels.PostVote) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	if err := entity.Insert(ctx, exec, boil.Infer()); err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.PostVotes)
	}
	return nil
}

// Read implements [postvote.Repository].
func (r *repository) Read(ctx context.Context, userID string, postID string) (*dbmodels.PostVote, error) {
	v, err := dbmodels.FindPostVote(ctx, r.db, userID, postID)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.PostVotes)
	}
	return v, nil
}

// List implements [postvote.Repository].
func (r *repository) List(ctx context.Context, opts *postvote.ListOptions) (map[string]*dbmodels.PostVote, error) {
	if opts == nil || len(opts.PostIDs) == 0 || opts.UserID == "" {
		return nil, nil
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id, post_id, vote_type, created_at
		FROM post_votes
		WHERE user_id = $1 AND post_id = ANY($2)`,
		opts.UserID, pq.Array(opts.PostIDs),
	)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.PostVotes)
	}
	defer rows.Close()

	result := make(map[string]*dbmodels.PostVote, len(opts.PostIDs))
	for rows.Next() {
		v := &dbmodels.PostVote{}
		if err := rows.Scan(&v.UserID, &v.PostID, &v.VoteType, &v.CreatedAt); err != nil {
			return nil, toddlerr.FromDBError(err, dbmodels.TableNames.PostVotes)
		}
		result[v.PostID] = v
	}
	if err := rows.Err(); err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.PostVotes)
	}
	return result, nil
}

// UpdateCounts implements [postvote.Repository].
func (r *repository) UpdateCounts(ctx context.Context, tx *sql.Tx, id string, likesDelta int, dislikesDelta int, expectedLikes int, expectedDislikes int) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	result, err := exec.ExecContext(
		ctx,
		`
		UPDATE posts
		SET
			likes = likes + $1,
			dislikes = dislikes + $2
		WHERE
			id = $3
			AND likes = $4
			AND dislikes = $5
		`,
		likesDelta,
		dislikesDelta,
		id,
		expectedLikes,
		expectedDislikes,
	)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Posts)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.Conflict,
			ServiceStatusCode: status.Conflict,
			PublicMessage:     "Please try again later.",
			ServiceMessage:    "optimistic locking failed",
		}
	}

	return nil
}

// ListVotersByPostIDs implements [postvote.Repository].
func (r *repository) ListVotersByPostIDs(ctx context.Context, postIDs []string) (map[string][]string, error) {
	if len(postIDs) == 0 {
		return nil, nil
	}
	votes, err := dbmodels.PostVotes(
		dbmodels.PostVoteWhere.PostID.IN(postIDs),
	).All(ctx, r.db)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.PostVotes)
	}
	result := make(map[string][]string, len(postIDs))
	for _, v := range votes {
		result[v.PostID] = append(result[v.PostID], v.UserID)
	}
	return result, nil
}

// Update implements [postvote.Repository].
func (r *repository) Update(ctx context.Context, tx *sql.Tx, entity *dbmodels.PostVote) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}
	rowAff, err := entity.Update(ctx, exec, boil.Infer())
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.PostVotes)
	}
	if rowAff == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "post vote not found",
			ServiceMessage:    "post vote not found",
		}
	}
	return nil
}
