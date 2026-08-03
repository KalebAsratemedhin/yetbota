package photo

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	domainPhoto "github.com/beka-birhanu/yetbota/identity-service/internal/domain/photo"
)

type repository struct {
	db *sql.DB
}

type Config struct {
	DB *sql.DB `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}

	return nil
}

func NewRepo(c *Config) (domainPhoto.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	return &repository{db: c.DB}, nil
}

// Add implements [photo.Repository].
func (r *repository) Add(ctx context.Context, tx *sql.Tx, u *dbmodels.Photo) (*dbmodels.Photo, error) {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	if err := u.Insert(ctx, exec, boil.Infer()); err != nil {
		return nil, err
	}
	return u, nil
}

// Read implements [photo.Repository].
func (r *repository) Read(ctx context.Context, id string) (*dbmodels.Photo, error) {
	u, err := dbmodels.FindPhoto(ctx, r.db, id)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.Photos)
	}
	return u, nil
}
