package notification

import (
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	dmn "github.com/beka-birhanu/yetbota/content-service/internal/domain/notification"
)

func buildQuery(filters *dmn.Filter) []qm.QueryMod {
	var mods []qm.QueryMod
	if filters.UserID != "" {
		mods = append(mods, dbmodels.NotificationWhere.UserID.EQ(filters.UserID))
	}
	if filters.Unread {
		mods = append(mods, dbmodels.NotificationWhere.ReadAt.IsNull())
	}
	return mods
}

func paginationOffset(page, length int32) int {
	return int((page - 1) * length)
}
