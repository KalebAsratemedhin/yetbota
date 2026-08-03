package moderation

import (
	"context"
	"database/sql"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainModeration "github.com/beka-birhanu/yetbota/content-service/internal/domain/moderation"
)

type banner struct {
	db *sql.DB
}

type BannerConfig struct {
	DB *sql.DB `validate:"required"`
}

func (c *BannerConfig) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewBanner(cfg *BannerConfig) (domainModeration.UserBanner, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &banner{db: cfg.DB}, nil
}

func (b *banner) Ban(ctx context.Context, userID, reason, actorID string) error {
	_, err := b.db.ExecContext(ctx, "UPDATE users SET status = $1, updated_at = now() WHERE id = $2", dbmodels.UserStatusBANNED, userID)
	return err
}
