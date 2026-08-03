package userdevice

import (
	"context"
	"database/sql"

	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
)

type Options struct {
	UserIDs   []string
	DeviceIDs []string
}

type Pagination struct {
	Limit int
	Page  int
}

type SortField string

var (
	SortFieldDeviceID  SortField = SortField(dbmodels.UserDeviceColumns.DeviceID)
	SortFieldCreatedAt SortField = SortField(dbmodels.UserDeviceColumns.CreatedAt)
	SortFieldUpdatedAt SortField = SortField(dbmodels.UserDeviceColumns.UpdatedAt)
	SortFieldUserID    SortField = SortField(dbmodels.UserDeviceColumns.UserID)
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
	Add(ctx context.Context, tx *sql.Tx, d *dbmodels.UserDevice) (*dbmodels.UserDevice, error)
	Delete(ctx context.Context, tx *sql.Tx, id string) error
	List(ctx context.Context, opts *Options, pagination *Pagination, sort *SortOption) (dbmodels.UserDeviceSlice, error)
}
