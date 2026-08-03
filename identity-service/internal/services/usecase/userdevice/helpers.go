package userdevice

import (
	"github.com/aarondl/null/v8"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	ctxYB "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
)

func newFromRegisterRequest(req *RegisterDeviceRequest, ctxSess *ctxYB.Context) *dbmodels.UserDevice {
	d := &dbmodels.UserDevice{
		UserID:   ctxSess.UserSession.UserID,
		DeviceID: req.DeviceID,
	}
	if req.Token != "" {
		d.Token = null.StringFrom(req.Token)
	}
	if req.Oem != "" {
		d.Oem = null.StringFrom(req.Oem)
	}
	if req.Device != "" {
		d.Device = null.StringFrom(req.Device)
	}
	if req.OS != "" {
		d.Os = null.StringFrom(req.OS)
	}
	if req.Longitude != 0 {
		d.Long = null.Float64From(req.Longitude)
	}
	if req.Latitude != 0 {
		d.Lat = null.Float64From(req.Latitude)
	}
	return d
}
