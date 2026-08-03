package notification

import (
	"context"
	"database/sql"

	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
)

type Repository interface {
	Add(ctx context.Context, tx *sql.Tx, notification *dbmodels.Notification) error
	List(ctx context.Context, filters *Filter, limit, page int32) ([]*dbmodels.Notification, error)
	Count(ctx context.Context, filters *Filter) (int64, error)
	Read(ctx context.Context, id string) (*dbmodels.Notification, error)
	ReadMany(ctx context.Context, ids []string) ([]*dbmodels.Notification, error)
	MarkAsRead(ctx context.Context, tx *sql.Tx, notifications dbmodels.NotificationSlice) error
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}

type Filter struct {
	UserID string
	Unread bool
}
