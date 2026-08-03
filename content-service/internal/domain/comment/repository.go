package comment

import (
	"context"
	"database/sql"
	"errors"

	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
)

var ErrConflict = errors.New("optimistic lock conflict")

type Repository interface {
	Add(ctx context.Context, tx *sql.Tx, entity *dbmodels.Comment) error
	Read(ctx context.Context, id string) (*dbmodels.Comment, error)
	List(ctx context.Context, opts *Options) (dbmodels.CommentSlice, error)
	Count(ctx context.Context, opts *Options) (int64, error)
	Delete(ctx context.Context, tx *sql.Tx, id string) error
	UpdateModerationStatus(ctx context.Context, tx *sql.Tx, id, status string) error
}

type Options struct {
	PostID    string
	CommentID string
	Page      int
	PageSize  int
	OnlyVisible bool
}
