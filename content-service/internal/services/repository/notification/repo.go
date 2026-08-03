package notification

import (
	"context"
	"database/sql"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	dmn "github.com/beka-birhanu/yetbota/content-service/internal/domain/notification"
)

const entity = "notification"

type repository struct {
	db *sql.DB
}

type Config struct {
	DB *sql.DB `validate:"required"`
}

func (c *Config) Validate() error {
	return validator.Validate.Struct(c)
}

func NewRepo(c *Config) (dmn.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: c.DB}, nil
}

func (r *repository) List(ctx context.Context, filters *dmn.Filter, limit, page int32) ([]*dbmodels.Notification, error) {
	mods := buildQuery(filters)
	mods = append(mods,
		qm.Limit(int(limit)),
		qm.Offset(paginationOffset(page, limit)),
		qm.OrderBy(dbmodels.NotificationColumns.SentAt+" DESC"),
	)

	rows, err := dbmodels.Notifications(mods...).All(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, toddlerr.FromDBError(err, entity)
	}
	return rows, nil
}

func (r *repository) Count(ctx context.Context, filters *dmn.Filter) (int64, error) {
	count, err := dbmodels.Notifications(buildQuery(filters)...).Count(ctx, r.db)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, toddlerr.FromDBError(err, entity)
	}
	return count, nil
}

func (r *repository) Add(ctx context.Context, tx *sql.Tx, n *dbmodels.Notification) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}
	if err := n.Insert(ctx, exec, boil.Infer()); err != nil {
		return toddlerr.FromDBError(err, entity)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}
	rowsAff, err := dbmodels.Notifications(dbmodels.NotificationWhere.ID.EQ(id)).DeleteAll(ctx, exec)
	if err != nil {
		return toddlerr.FromDBError(err, entity)
	}
	if rowsAff == 0 {
		return toddlerr.FromDBError(sql.ErrNoRows, entity)
	}
	return nil
}

func (r *repository) Read(ctx context.Context, id string) (*dbmodels.Notification, error) {
	n, err := dbmodels.FindNotification(ctx, r.db, id)
	if err != nil {
		return nil, toddlerr.FromDBError(err, entity)
	}
	return n, nil
}

func (r *repository) ReadMany(ctx context.Context, ids []string) ([]*dbmodels.Notification, error) {
	rows, err := dbmodels.Notifications(dbmodels.NotificationWhere.ID.IN(ids)).All(ctx, r.db)
	if err != nil {
		return nil, toddlerr.FromDBError(err, entity)
	}
	return rows, nil
}

func (r *repository) MarkAsRead(ctx context.Context, tx *sql.Tx, notifications dbmodels.NotificationSlice) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}
	_, err := notifications.UpdateAll(ctx, exec, dbmodels.M{
		dbmodels.NotificationColumns.ReadAt: null.TimeFrom(time.Now()),
	})
	if err != nil {
		return toddlerr.FromDBError(err, entity)
	}
	return nil
}
