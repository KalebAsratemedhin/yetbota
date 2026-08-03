package userdevice

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	ctxYB "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
	deviceSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/userdevice"
	"github.com/beka-birhanu/yetbota/identity-service/internal/transport/http/shared"
)

type deviceDTO struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	DeviceID  string    `json:"device_id"`
	Oem       string    `json:"oem,omitempty"`
	Device    string    `json:"device,omitempty"`
	OS        string    `json:"os,omitempty"`
	Longitude float64   `json:"longitude,omitempty"`
	Latitude  float64   `json:"latitude,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toDeviceDTO(d *dbmodels.UserDevice) deviceDTO {
	if d == nil {
		return deviceDTO{}
	}
	return deviceDTO{
		ID:        d.ID,
		UserID:    d.UserID,
		DeviceID:  d.DeviceID,
		Oem:       d.Oem.String,
		Device:    d.Device.String,
		OS:        d.Os.String,
		Longitude: d.Long.Float64,
		Latitude:  d.Lat.Float64,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func setCtxResponse(ctx context.Context, env shared.Envelope) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.Response = env
}

type deviceData struct {
	Device deviceDTO `json:"device"`
}

func encodeRegisterDeviceHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*deviceSvc.RegisterDeviceResponse)
	if !ok || out == nil || out.Device == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{Success: true, Data: deviceData{Device: toDeviceDTO(out.Device)}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}
