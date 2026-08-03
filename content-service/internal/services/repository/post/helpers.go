package post

import (
	"fmt"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	"github.com/lib/pq"
)

func FilterMods(opts *domainPost.ListOptions) []qm.QueryMod {
	var filterMods []qm.QueryMod
	if opts == nil {
		return filterMods
	}

	if len(opts.IDs) > 0 {
		filterMods = append(filterMods, dbmodels.PostWhere.ID.IN(opts.IDs))
	}
	if opts.UserID != "" {
		filterMods = append(filterMods, dbmodels.PostWhere.UserID.EQ(opts.UserID))
	}
	if opts.Search != "" {
		pat := fmt.Sprintf("%%%s%%", opts.Search)
		filterMods = append(filterMods, qm.Expr(
			dbmodels.PostWhere.Title.ILIKE(pat),
			qm.Or2(dbmodels.PostWhere.Description.ILIKE(pat)),
		))
	}
	if opts.IsQuestion != nil {
		filterMods = append(filterMods, dbmodels.PostWhere.IsQuestion.EQ(*opts.IsQuestion))
	}
	if opts.AttachedPostID != "" {
		filterMods = append(filterMods, dbmodels.PostWhere.AttachedPostID.EQ(null.StringFrom(opts.AttachedPostID)))
	}
	if opts.OnlyVisible {
		filterMods = append(filterMods, dbmodels.PostWhere.ModerationStatus.EQ(constants.ModerationStatusVisible))
	} else if opts.NonVisibleOnly {
		filterMods = append(filterMods, dbmodels.PostWhere.ModerationStatus.NEQ(constants.ModerationStatusVisible))
	}
	if len(opts.Tags) > 0 {
		filterMods = append(filterMods, qm.Where("tags && ?", pq.Array(opts.Tags)))
	}
	if opts.SavedBy != nil {
		filterMods = append(filterMods, qm.InnerJoin("saved_posts ON saved_posts.post_id = posts.id AND saved_posts.user_id = ?", *opts.SavedBy))
	}

	if opts.NearLat != nil && opts.NearLon != nil && opts.RadiusKm != nil {
		filterMods = append(filterMods, qm.Where(
			"location IS NOT NULL AND ST_DWithin(location::geography, ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography, ?)",
			*opts.NearLon, *opts.NearLat, *opts.RadiusKm*1000,
		))
	}

	return filterMods
}

func SortMods(opts *domainPost.ListOptions) qm.QueryMod {
	if opts == nil {
		return nil
	}

	if opts.SortField == "" || opts.SortDir == "" {
		return nil
	}

	return qm.OrderBy(fmt.Sprintf("%s %s", opts.SortField, opts.SortDir))
}

func PaginationMods(opts *domainPost.ListOptions) []qm.QueryMod {
	if opts == nil {
		return nil
	}
	pageSize := opts.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	page := opts.Page
	if page <= 0 {
		page = 1
	}

	return []qm.QueryMod{
		qm.Limit(pageSize),
		qm.Offset((page - 1) * pageSize),
	}
}
