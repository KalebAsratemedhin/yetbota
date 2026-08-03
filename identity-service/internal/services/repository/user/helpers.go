package user

import (
	"fmt"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	domainUser "github.com/beka-birhanu/yetbota/identity-service/internal/domain/user"
)

func buildFilterMods(opts *domainUser.Options) []qm.QueryMod {
	var mods []qm.QueryMod
	if opts == nil {
		return mods
	}

	if opts.FirstName != "" {
		mods = append(mods, dbmodels.UserWhere.FirstName.EQ(opts.FirstName))
	}
	if opts.Surname != "" {
		mods = append(mods, dbmodels.UserWhere.LastName.EQ(opts.Surname))
	}
	if opts.Username != "" {
		mods = append(mods, dbmodels.UserWhere.Username.EQ(opts.Username))
	}
	if opts.Status != "" {
		mods = append(mods, dbmodels.UserWhere.Status.EQ(opts.Status))
	}
	if opts.Role != "" {
		mods = append(mods, dbmodels.UserWhere.Role.EQ(opts.Role))
	}
	if opts.Search != "" {
		pattern := "%" + opts.Search + "%"
		mods = append(mods, qm.Where("(username ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?)", pattern, pattern, pattern))
	}
	if opts.MinScore != nil {
		mods = append(mods, dbmodels.UserWhere.Score.GTE(*opts.MinScore))
	}
	if opts.MaxScore != nil {
		mods = append(mods, dbmodels.UserWhere.Score.LTE(*opts.MaxScore))
	}
	if opts.CreatedFrom != nil {
		mods = append(mods, dbmodels.UserWhere.CreatedAt.GTE(*opts.CreatedFrom))
	}
	if opts.CreatedTo != nil {
		mods = append(mods, dbmodels.UserWhere.CreatedAt.LTE(*opts.CreatedTo))
	}

	return mods
}

func buildLoadMods(opts *domainUser.Options) []qm.QueryMod {
	if opts == nil || !opts.LoadPhoto {
		return nil
	}
	return []qm.QueryMod{qm.Load(dbmodels.UserRels.ProfilePhoto)}
}

func buildQueryMods(opts *domainUser.Options) []qm.QueryMod {
	mods := buildFilterMods(opts)
	return append(mods, buildLoadMods(opts)...)
}

func buildPaginationMods(pagination *domainUser.Pagination) []qm.QueryMod {
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

func buildSortMods(sort *domainUser.SortOption) []qm.QueryMod {
	var mods []qm.QueryMod
	if sort == nil || sort.Field == "" || sort.Direction == "" {
		return append(mods, qm.OrderBy(fmt.Sprintf("%s %s", dbmodels.UserColumns.ID, "ASC")))
	}

	mods = append(mods, qm.OrderBy(fmt.Sprintf("%s %s", sort.Field, sort.Direction)))
	return mods
}
