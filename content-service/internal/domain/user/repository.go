package user

import (
	"context"

	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
)

type Repository interface {
	Read(ctx context.Context, id string) (*dbmodels.User, error)
	Update(ctx context.Context, model *dbmodels.User) error
	GetDeviceTokens(ctx context.Context, userID string) ([]string, error)
}
