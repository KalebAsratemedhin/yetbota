package postphoto

import (
	"context"
	"database/sql"

	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
)

type Repository interface {
	Add(ctx context.Context, tx *sql.Tx, entity *dbmodels.PostPhoto) error
	AddBulk(ctx context.Context, tx *sql.Tx, entities dbmodels.PostPhotoSlice) error
	Read(ctx context.Context, id string, opts *Options) (*dbmodels.PostPhoto, error)
	Update(ctx context.Context, tx *sql.Tx, entity *dbmodels.PostPhoto) error
	Delete(ctx context.Context, tx *sql.Tx, id string) error
	List(ctx context.Context, opts *Options, sort *SortOptions) (dbmodels.PostPhotoSlice, error)
}

type Options struct {
	LoadPhoto bool
	PostIDs   []string
}

type SortOptions struct {
	Field     SortField
	Direction SortDirection
}

type (
	SortField     string
	SortDirection string
)

const (
	SortDirectionAsc  SortDirection = "ASC"
	SortDirectionDesc SortDirection = "DESC"
)

var (
	SortFieldID       SortField = SortField(dbmodels.PostPhotoColumns.ID)
	SortFieldPostID   SortField = SortField(dbmodels.PostPhotoColumns.PostID)
	SortFieldPosition SortField = SortField(dbmodels.PostPhotoColumns.Position)
)
