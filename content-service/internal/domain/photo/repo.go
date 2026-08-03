package photo

import (
	"context"
	"database/sql"

	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
)

type Repository interface {
	Add(ctx context.Context, tx *sql.Tx, u *dbmodels.Photo) error
	AddBulk(ctx context.Context, tx *sql.Tx, u dbmodels.PhotoSlice) error
	Read(ctx context.Context, id string) (*dbmodels.Photo, error)
	Update(ctx context.Context, tx *sql.Tx, u *dbmodels.Photo) error
	Delete(ctx context.Context, tx *sql.Tx, id string) error
}
