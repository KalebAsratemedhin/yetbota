package userdevice

import (
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	toddlerr "github.com/beka-birhanu/toddler/error"
)

type RegisterDeviceRequest struct {
	DeviceID  string  `validate:"required"`
	Token     string  `mask:"true"`
	Oem       string
	Device    string
	OS        string
	Longitude float64
	Latitude  float64
}

func (r *RegisterDeviceRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type RegisterDeviceResponse struct {
	Device *dbmodels.UserDevice
}
