package userdevice

import (
	"fmt"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	domainUserDevice "github.com/beka-birhanu/yetbota/identity-service/internal/domain/userdevice"
)

func buildQueryMods(opts *domainUserDevice.Options) []qm.QueryMod {
	var mods []qm.QueryMod
	if opts == nil {
		return mods
	}

	if len(opts.UserIDs) > 0 {
		mods = append(mods, dbmodels.UserDeviceWhere.UserID.IN(opts.UserIDs))
	}

	if len(opts.DeviceIDs) > 0 {
		mods = append(mods, dbmodels.UserDeviceWhere.DeviceID.IN(opts.DeviceIDs))
	}

	return mods
}

func buildPaginationMods(pagination *domainUserDevice.Pagination) []qm.QueryMod {
	var mods []qm.QueryMod
	if pagination == nil {
		return mods
	}

	if pagination.Limit > 0 && pagination.Page > 0 {
		offset := (pagination.Page - 1) * pagination.Limit

		mods = append(mods, qm.Limit(pagination.Limit))
		mods = append(mods, qm.Offset(offset))
	}

	return mods
}

func buildSortMods(sort *domainUserDevice.SortOption) []qm.QueryMod {
	var mods []qm.QueryMod
	if sort == nil || sort.Field == "" || sort.Direction == "" {
		return append(mods, qm.OrderBy(fmt.Sprintf("%s %s", dbmodels.UserDeviceColumns.CreatedAt, "DESC")))
	}

	mods = append(mods, qm.OrderBy(fmt.Sprintf("%s %s", sort.Field, sort.Direction)))
	return mods
}
