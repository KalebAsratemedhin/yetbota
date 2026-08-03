package commentvote

import (
	"context"
	"database/sql"

	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
)

type Repository interface {
	Exists(ctx context.Context, userID, commentID string) (bool, error)
	Read(ctx context.Context, userID, commentID string) (*dbmodels.CommentVote, error)
	Add(ctx context.Context, tx *sql.Tx, entity *dbmodels.CommentVote) error
	Update(ctx context.Context, tx *sql.Tx, entity *dbmodels.CommentVote) error
	UpdateCounts(ctx context.Context, tx *sql.Tx, id string, upvoteDelta, downvoteDelta, expectedUpvote, expectedDownvote int) error
	List(ctx context.Context, opts *ListOptions) (map[string]*dbmodels.CommentVote, error)
}

type ListOptions struct {
	UserID     string
	CommentIDs []string
}
