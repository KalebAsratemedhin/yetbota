package comment

import (
	"context"
	"database/sql"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainComment "github.com/beka-birhanu/yetbota/content-service/internal/domain/comment"
)

type repository struct {
	db *sql.DB
}

type Config struct {
	DB *sql.DB `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewRepo(cfg *Config) (domainComment.Repository, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: cfg.DB}, nil
}

func (r *repository) Add(ctx context.Context, tx *sql.Tx, entity *dbmodels.Comment) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}
	if err := entity.Insert(ctx, exec, boil.Infer()); err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Comments)
	}
	return nil
}

func (r *repository) Read(ctx context.Context, id string) (*dbmodels.Comment, error) {
	c, err := dbmodels.FindComment(ctx, r.db, id)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.Comments)
	}
	return c, nil
}

func (r *repository) List(ctx context.Context, opts *domainComment.Options) (dbmodels.CommentSlice, error) {
	mods := filterMods(opts)
	mods = append(mods, qm.OrderBy(dbmodels.CommentColumns.CreatedAt+" DESC"))
	mods = append(mods, paginationMods(opts)...)

	result, err := dbmodels.Comments(mods...).All(ctx, r.db)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.Comments)
	}
	return result, nil
}

func (r *repository) Count(ctx context.Context, opts *domainComment.Options) (int64, error) {
	count, err := dbmodels.Comments(filterMods(opts)...).Count(ctx, r.db)
	if err != nil {
		return 0, toddlerr.FromDBError(err, dbmodels.TableNames.Comments)
	}
	return count, nil
}

func filterMods(opts *domainComment.Options) []qm.QueryMod {
	var mods []qm.QueryMod
	if opts == nil {
		return mods
	}
	if opts.PostID != "" {
		mods = append(mods, dbmodels.CommentWhere.PostID.EQ(opts.PostID))
	}
	if opts.CommentID != "" {
		mods = append(mods, dbmodels.CommentWhere.CommentID.EQ(null.StringFrom(opts.CommentID)))
	}
	if opts.OnlyVisible {
		mods = append(mods, dbmodels.CommentWhere.ModerationStatus.EQ(constants.ModerationStatusVisible))
	}
	return mods
}

func paginationMods(opts *domainComment.Options) []qm.QueryMod {
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

func (r *repository) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}
	c := &dbmodels.Comment{ID: id}
	_, err := c.Delete(ctx, exec)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Comments)
	}
	return nil
}

func (r *repository) UpdateModerationStatus(ctx context.Context, tx *sql.Tx, id, modStatus string) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}
	_, err := exec.ExecContext(ctx,
		`UPDATE comments SET moderation_status = $1, deleted_at = CASE WHEN $1 = 'REMOVED' THEN now() ELSE NULL END WHERE id = $2`,
		modStatus, id,
	)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Comments)
	}
	return nil
}
