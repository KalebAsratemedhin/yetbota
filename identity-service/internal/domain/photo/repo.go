package photo

import (
	"context"
	"database/sql"

	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
)

type Repository interface {
	Add(ctx context.Context, tx *sql.Tx, u *dbmodels.Photo) (*dbmodels.Photo, error)
	Read(ctx context.Context, id string) (*dbmodels.Photo, error)
}
