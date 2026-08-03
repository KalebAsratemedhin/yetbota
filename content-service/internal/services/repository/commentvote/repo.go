package commentvote

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/commentvote"
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

func NewRepo(c *Config) (commentvote.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: c.DB}, nil
}

// Exists implements [commentvote.Repository].
func (r *repository) Exists(ctx context.Context, userID string, commentID string) (bool, error) {
	exists, err := dbmodels.CommentVoteExists(ctx, r.db, userID, commentID)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Read implements [commentvote.Repository].
func (r *repository) Read(ctx context.Context, userID string, commentID string) (*dbmodels.CommentVote, error) {
	v, err := dbmodels.FindCommentVote(ctx, r.db, userID, commentID)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.CommentVotes)
	}
	return v, nil
}

// List implements [commentvote.Repository].
func (r *repository) List(ctx context.Context, opts *commentvote.ListOptions) (map[string]*dbmodels.CommentVote, error) {
	if opts == nil || len(opts.CommentIDs) == 0 {
		return nil, nil
	}
	votes, err := dbmodels.CommentVotes(
		dbmodels.CommentVoteWhere.UserID.EQ(opts.UserID),
		dbmodels.CommentVoteWhere.CommentID.IN(opts.CommentIDs),
	).All(ctx, r.db)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.CommentVotes)
	}
	result := make(map[string]*dbmodels.CommentVote, len(votes))
	for _, v := range votes {
		result[v.CommentID] = v
	}
	return result, nil
}

// Add implements [commentvote.Repository].
func (r *repository) Add(ctx context.Context, tx *sql.Tx, entity *dbmodels.CommentVote) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	if err := entity.Insert(ctx, exec, boil.Infer()); err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.CommentVotes)
	}
	return nil
}

// Update implements [commentvote.Repository].
func (r *repository) Update(ctx context.Context, tx *sql.Tx, entity *dbmodels.CommentVote) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}
	rowAff, err := entity.Update(ctx, exec, boil.Infer())
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.CommentVotes)
	}
	if rowAff == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "comment vote not found",
			ServiceMessage:    "comment vote not found",
		}
	}
	return nil
}

// UpdateCounts implements [commentvote.Repository].
func (r *repository) UpdateCounts(ctx context.Context, tx *sql.Tx, id string, upvoteDelta int, downvoteDelta int, expectedUpvote int, expectedDownvote int) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	result, err := exec.ExecContext(
		ctx,
		`
		UPDATE comments
		SET
			upvote = upvote + $1,
			downvote = downvote + $2
		WHERE
			id = $3
			AND upvote = $4
			AND downvote = $5
		`,
		upvoteDelta,
		downvoteDelta,
		id,
		expectedUpvote,
		expectedDownvote,
	)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Comments)
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

