package post

import (
	"context"
	"database/sql"

	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
)

type ListSortField string

var (
	ListSortFieldCreatedAt = ListSortField(dbmodels.PostColumns.CreatedAt)
	ListSortFieldLikes     = ListSortField(dbmodels.PostColumns.Likes)
	ListSortFieldDislikes  = ListSortField(dbmodels.PostColumns.Dislikes)
	ListSortFieldComments  = ListSortField(dbmodels.PostColumns.CommentCount)
)

type ListSortDir string

const (
	ListSortDirAsc  ListSortDir = "ASC"
	ListSortDirDesc ListSortDir = "DESC"
)

type ListOptions struct {
	IDs            []string
	UserID         string
	Tags           []string
	IsQuestion     *bool
	AttachedPostID string
	Search         string
	NearLat        *float64
	NearLon        *float64
	RadiusKm       *float64
	SavedBy        *string
	SortField      ListSortField
	SortDir        ListSortDir
	Page           int
	PageSize       int
	OnlyVisible    bool
	NonVisibleOnly bool
}

type Repository interface {
	Add(ctx context.Context, tx *sql.Tx, entity *dbmodels.Post) error
	Read(ctx context.Context, id string) (*dbmodels.Post, error)
	Update(ctx context.Context, tx *sql.Tx, entity *dbmodels.Post) error
	List(ctx context.Context, opts *ListOptions) ([]*dbmodels.Post, error)
	Count(ctx context.Context, opts *ListOptions) (int64, error)
	UpdateCommentCount(ctx context.Context, tx *sql.Tx, postID string, delta int, expectedCount int) error
	UpdateModerationStatus(ctx context.Context, tx *sql.Tx, id, status string) error
}
