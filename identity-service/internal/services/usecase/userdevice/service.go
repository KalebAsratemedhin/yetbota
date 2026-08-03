package userdevice

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	ctxYB "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
	domainDevice "github.com/beka-birhanu/yetbota/identity-service/internal/domain/userdevice"
)

type Service interface {
	RegisterDevice(ctx context.Context, ctxSess *ctxYB.Context, req *RegisterDeviceRequest) (*RegisterDeviceResponse, error)
}

type Config struct {
	DeviceRepo domainDevice.Repository `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type svc struct {
	deviceRepo domainDevice.Repository
}

func NewService(cfg *Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &svc{deviceRepo: cfg.DeviceRepo}, nil
}
