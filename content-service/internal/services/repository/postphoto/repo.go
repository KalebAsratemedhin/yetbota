package postphoto

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/postphoto"
)

type repository struct {
	db *sql.DB
}

type Config struct {
	DB *sql.DB `validate:"required"`
}

func (c *Config) Validate() error {
	return nil
}

func NewRepo(c *Config) (postphoto.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	return &repository{db: c.DB}, nil
}

func (r *repository) AddBulk(ctx context.Context, tx *sql.Tx, entities dbmodels.PostPhotoSlice) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	for _, entity := range entities {
		if err := entity.Insert(ctx, exec, boil.Infer()); err != nil {
			return toddlerr.FromDBError(err, dbmodels.TableNames.PostPhotos)
		}
	}
	return nil
}

func (r *repository) List(ctx context.Context, opts *postphoto.Options, sort *postphoto.SortOptions) (dbmodels.PostPhotoSlice, error) {
	if opts == nil || len(opts.PostIDs) == 0 {
		return nil, nil
	}

	mods := buildQueryMods(opts)

	if sort != nil {
		mods = append(mods, qm.OrderBy(fmt.Sprintf("%s %s", string(sort.Field), string(sort.Direction))))
	}

	result, err := dbmodels.PostPhotos(mods...).All(ctx, r.db)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.PostPhotos)
	}
	return result, nil
}

func (r *repository) Add(ctx context.Context, tx *sql.Tx, entity *dbmodels.PostPhoto) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	if err := entity.Insert(ctx, exec, boil.Infer()); err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.PostPhotos)
	}
	return nil
}

func (r *repository) Read(ctx context.Context, id string, opts *postphoto.Options) (*dbmodels.PostPhoto, error) {
	mods := buildQueryMods(opts)
	mods = append(mods, dbmodels.PostPhotoWhere.ID.EQ(id))

	postphoto, err := dbmodels.PostPhotos(mods...).One(ctx, r.db)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.PostPhotos)
	}
	return postphoto, nil
}

func (r *repository) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	pp := &dbmodels.PostPhoto{ID: id}
	_, err := pp.Delete(ctx, exec)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.PostPhotos)
	}
	return nil
}

func (r *repository) Update(ctx context.Context, tx *sql.Tx, entity *dbmodels.PostPhoto) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	rowAff, err := entity.Update(ctx, exec, boil.Infer())
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.PostPhotos)
	}

	if rowAff == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "PostPhoto not found",
			ServiceMessage:    fmt.Sprintf("postphoto not found id: %s", entity.ID),
		}
	}
	return nil
}
