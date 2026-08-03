package savedpost

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/savedpost"
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

func NewRepo(c *Config) (savedpost.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: c.DB}, nil
}

// Exists implements [savedpost.Repository].
func (r *repository) Exists(ctx context.Context, userID string, postIDs []string) (map[string]bool, error) {
	if len(postIDs) == 0 || userID == "" {
		return nil, nil
	}
	saved, err := dbmodels.SavedPosts(
		dbmodels.SavedPostWhere.UserID.EQ(userID),
		dbmodels.SavedPostWhere.PostID.IN(postIDs),
	).All(ctx, r.db)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.SavedPosts)
	}
	result := make(map[string]bool, len(saved))
	for _, sp := range saved {
		result[sp.PostID] = true
	}
	return result, nil
}

// Add implements [savedpost.Repository]. Idempotent — saving an already-saved post is a no-op.
func (r *repository) Add(ctx context.Context, entity *dbmodels.SavedPost) error {
	if err := entity.Upsert(ctx, r.db, false, nil, boil.Infer(), boil.Infer()); err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.SavedPosts)
	}
	return nil
}

// Delete implements [savedpost.Repository].
func (r *repository) Delete(ctx context.Context, entity *dbmodels.SavedPost) error {
	if _, err := entity.Delete(ctx, r.db); err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.SavedPosts)
	}
	return nil
}

// Read implements [savedpost.Repository].
func (r *repository) Read(ctx context.Context, userID, postID string) (*dbmodels.SavedPost, error) {
	sp, err := dbmodels.FindSavedPost(ctx, r.db, userID, postID)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.SavedPosts)
	}
	return sp, nil
}
