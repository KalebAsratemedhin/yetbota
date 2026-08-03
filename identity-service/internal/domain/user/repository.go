package user

import (
	"context"
	"database/sql"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
)

type Options struct {
	FirstName   string
	Surname     string
	Username    string
	Status      string `validate:"omitempty,oneof=ACTIVE INACTIVE BANNED"`
	Role        string `validate:"omitempty,oneof=USER ADMIN"`
	Search      string
	MinScore    *int64
	MaxScore    *int64
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	LoadPhoto   bool
}

type Pagination struct {
	Limit int
	Page  int
}

type SortField string

var (
	SortFieldRating        SortField = SortField(dbmodels.UserColumns.Score)
	SortFieldFollowers     SortField = SortField(dbmodels.UserColumns.Followers)
	SortFieldFollowing     SortField = SortField(dbmodels.UserColumns.Following)
	SortFieldContributions SortField = SortField(dbmodels.UserColumns.Contributions)
	SortFieldCreatedAt     SortField = SortField(dbmodels.UserColumns.CreatedAt)
	SortFieldUpdatedAt     SortField = SortField(dbmodels.UserColumns.UpdatedAt)
)

type SortDirection string

const (
	SortDirectionAsc  SortDirection = "ASC"
	SortDirectionDesc SortDirection = "DESC"
)

type SortOption struct {
	Field     SortField
	Direction SortDirection
}

type Repository interface {
	List(ctx context.Context, opts *Options, pagination *Pagination, sort *SortOption) (dbmodels.UserSlice, error)
	Count(ctx context.Context, opts *Options) (int64, error)
	Add(ctx context.Context, tx *sql.Tx, u *dbmodels.User) (*dbmodels.User, error)
	Read(ctx context.Context, id string, opts *Options) (*dbmodels.User, error)
	ReadByUsername(ctx context.Context, username string, opts *Options) (*dbmodels.User, error)
	Update(ctx context.Context, tx *sql.Tx, u *dbmodels.User, cols boil.Columns) error
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}
