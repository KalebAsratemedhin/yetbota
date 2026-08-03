package savedpost

import (
	"context"

	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
)

type Repository interface {
	Exists(ctx context.Context, userID string, postIDs []string) (map[string]bool, error)
	Add(ctx context.Context, entity *dbmodels.SavedPost) error
	Delete(ctx context.Context, entity *dbmodels.SavedPost) error
	Read(ctx context.Context, userID string, postID string) (*dbmodels.SavedPost, error)
}
