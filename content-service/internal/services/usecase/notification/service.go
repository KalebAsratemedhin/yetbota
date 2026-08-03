package notification

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	dmn "github.com/beka-birhanu/yetbota/content-service/internal/domain/notification"
)

type Service interface {
	List(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error)
	MarkAsRead(ctx context.Context, ctxSess *ctxRP.Context, req *MarkAsReadRequest) (*MarkAsReadResponse, error)
	Delete(ctx context.Context, ctxSess *ctxRP.Context, req *DeleteRequest) error
}

type Config struct {
	NotificationRepo dmn.Repository `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type svc struct {
	notificationRepo dmn.Repository
}

func NewService(cfg *Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &svc{notificationRepo: cfg.NotificationRepo}, nil
}
