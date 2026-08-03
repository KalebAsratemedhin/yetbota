package photo

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainPhoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/photo"
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

func (r *repository) Add(ctx context.Context, tx *sql.Tx, u *dbmodels.Photo) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	if err := u.Insert(ctx, exec, boil.Infer()); err != nil {
		return err
	}
	return nil
}

func (r *repository) AddBulk(ctx context.Context, tx *sql.Tx, u dbmodels.PhotoSlice) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	for _, photo := range u {
		if err := photo.Insert(ctx, exec, boil.Infer()); err != nil {
			return toddlerr.FromDBError(err, dbmodels.TableNames.Photos)
		}
	}
	return nil
}

func (r *repository) Read(ctx context.Context, id string) (*dbmodels.Photo, error) {
	u, err := dbmodels.FindPhoto(ctx, r.db, id)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.Photos)
	}
	return u, nil
}

func (r *repository) Update(ctx context.Context, tx *sql.Tx, u *dbmodels.Photo) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	_, err := u.Update(ctx, exec, boil.Infer())
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Photos)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	p := &dbmodels.Photo{ID: id}
	_, err := p.Delete(ctx, exec)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Photos)
	}
	return nil
}
