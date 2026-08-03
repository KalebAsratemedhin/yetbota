package post

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
)

type repo struct {
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

func NewRepo(cfg *Config) (domainPost.Repository, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &repo{db: cfg.DB}, nil
}

func (r *repo) Add(ctx context.Context, tx *sql.Tx, entity *dbmodels.Post) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	if err := entity.Insert(ctx, exec, boil.Infer()); err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Posts)
	}
	return nil
}

func (r *repo) Read(ctx context.Context, id string) (*dbmodels.Post, error) {
	post, err := dbmodels.FindPost(ctx, r.db, id)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.Posts)
	}
	return post, nil
}

func (r *repo) Update(ctx context.Context, tx *sql.Tx, entity *dbmodels.Post) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	rowAff, err := entity.Update(ctx, exec, boil.Infer())
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Posts)
	}

	if rowAff == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "Post not found",
			ServiceMessage:    fmt.Sprintf("post not found id: %s", entity.ID),
		}
	}

	return nil
}

func (r *repo) List(ctx context.Context, opts *domainPost.ListOptions) ([]*dbmodels.Post, error) {
	filterMods := FilterMods(opts)
	paginationMods := PaginationMods(opts)
	sortMods := SortMods(opts)

	allMods := append(filterMods, paginationMods...)
	if sortMods != nil {
		allMods = append(allMods, sortMods)
	}

	posts, err := dbmodels.Posts(allMods...).All(ctx, r.db)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.Posts)
	}

	return posts, nil
}

// UpdateCommentCount implements [post.Repository].
func (r *repo) UpdateCommentCount(ctx context.Context, tx *sql.Tx, postID string, delta int, expectedCount int) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	result, err := exec.ExecContext(ctx,
		`UPDATE posts SET comment_count = comment_count + $1 WHERE id = $2 AND comment_count = $3`,
		delta, postID, expectedCount,
	)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Posts)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.Conflict,
			ServiceStatusCode: status.Conflict,
			PublicMessage:     "Please try again later.",
			ServiceMessage:    "post comment count: optimistic locking failed",
		}
	}
	return nil
}

// UpdateModerationStatus implements [post.Repository].
func (r *repo) UpdateModerationStatus(ctx context.Context, tx *sql.Tx, id, modStatus string) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	_, err := exec.ExecContext(ctx,
		`UPDATE posts SET moderation_status = $1, deleted_at = CASE WHEN $1 = 'REMOVED' THEN now() ELSE NULL END WHERE id = $2`,
		modStatus, id,
	)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Posts)
	}
	return nil
}

// Count implements [post.Repository].
func (r *repo) Count(ctx context.Context, opts *domainPost.ListOptions) (int64, error) {
	filterMods := FilterMods(opts)

	count, err := dbmodels.Posts(filterMods...).Count(ctx, r.db)
	if err != nil {
		return 0, toddlerr.FromDBError(err, dbmodels.TableNames.Posts)
	}
	return count, nil
}
