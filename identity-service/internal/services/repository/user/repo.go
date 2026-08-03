package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aarondl/sqlboiler/v4/boil"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	domainUser "github.com/beka-birhanu/yetbota/identity-service/internal/domain/user"
)

type repo struct {
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

func NewRepo(c *Config) (domainUser.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	return &repo{db: c.DB}, nil
}

func (r *repo) List(ctx context.Context, opts *domainUser.Options, pagination *domainUser.Pagination, sort *domainUser.SortOption) (dbmodels.UserSlice, error) {
	filterMods := buildQueryMods(opts)
	queryMods := append(filterMods, buildPaginationMods(pagination)...)
	queryMods = append(queryMods, buildSortMods(sort)...)

	users, err := dbmodels.Users(queryMods...).All(ctx, r.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.Users)
	}

	return users, nil
}

func (r *repo) Count(ctx context.Context, opts *domainUser.Options) (int64, error) {
	filterMods := buildFilterMods(opts)
	count, err := dbmodels.Users(filterMods...).Count(ctx, r.db)
	if err != nil {
		return 0, toddlerr.FromDBError(err, dbmodels.TableNames.Users)
	}
	return count, nil
}

func (r *repo) Read(ctx context.Context, id string, opts *domainUser.Options) (*dbmodels.User, error) {
	filterMods := buildQueryMods(opts)
	filterMods = append(filterMods, dbmodels.UserWhere.ID.EQ(id))
	u, err := dbmodels.Users(filterMods...).One(ctx, r.db)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.Users)
	}
	return u, nil
}

func (r *repo) ReadByUsername(ctx context.Context, username string, opts *domainUser.Options) (*dbmodels.User, error) {
	filterMods := buildQueryMods(opts)
	filterMods = append(filterMods, dbmodels.UserWhere.Username.EQ(username))
	u, err := dbmodels.Users(filterMods...).One(ctx, r.db)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.Users)
	}
	return u, nil
}

func (r *repo) Add(ctx context.Context, tx *sql.Tx, u *dbmodels.User) (*dbmodels.User, error) {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	if err := u.Insert(ctx, exec, boil.Infer()); err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.Users)
	}
	return u, nil
}

func (r *repo) Update(ctx context.Context, tx *sql.Tx, u *dbmodels.User, cols boil.Columns) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	if _, err := u.Update(ctx, exec, cols); err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Users)
	}

	return nil
}

func (r *repo) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	rowAff, err := dbmodels.Users(dbmodels.UserWhere.ID.EQ(id)).DeleteAll(ctx, exec)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Users)
	}

	if rowAff == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "User not found",
			ServiceMessage:    "user not found by id: " + id,
		}
	}

	return nil
}
