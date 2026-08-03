package userdevice

import (
	"context"
	"database/sql"
	"errors"

	"github.com/aarondl/sqlboiler/v4/boil"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	domainUserDevice "github.com/beka-birhanu/yetbota/identity-service/internal/domain/userdevice"
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

func NewRepo(c *Config) (domainUserDevice.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	return &repo{db: c.DB}, nil
}

func (r *repo) Add(ctx context.Context, tx *sql.Tx, d *dbmodels.UserDevice) (*dbmodels.UserDevice, error) {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	if err := d.Insert(ctx, exec, boil.Infer()); err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.UserDevices)
	}
	return d, nil
}

func (r *repo) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	rowsAff, err := dbmodels.UserDevices(dbmodels.UserDeviceWhere.ID.EQ(id)).DeleteAll(ctx, exec)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.UserDevices)
	}

	if rowsAff == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "user device not found",
			ServiceMessage:    "user device not found by id: " + id,
		}
	}

	return nil
}

func (r *repo) List(ctx context.Context, opts *domainUserDevice.Options, pagination *domainUserDevice.Pagination, sort *domainUserDevice.SortOption) (dbmodels.UserDeviceSlice, error) {
	filterMods := buildQueryMods(opts)
	queryMods := append(filterMods, buildPaginationMods(pagination)...)
	queryMods = append(queryMods, buildSortMods(sort)...)

	devices, err := dbmodels.UserDevices(queryMods...).All(ctx, r.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.UserDevices)
	}

	return devices, nil
}

