package userdevice

import (
	"context"
	"encoding/json"
	"net/http"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	ctxYB "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
	deviceSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/userdevice"
)

func decodeRegisterDeviceHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		DeviceID  string  `json:"device_id"`
		Token     string  `json:"token"`
		Oem       string  `json:"oem"`
		Device    string  `json:"device"`
		OS        string  `json:"os"`
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &deviceSvc.RegisterDeviceRequest{
		DeviceID:  in.DeviceID,
		Token:     in.Token,
		Oem:       in.Oem,
		Device:    in.Device,
		OS:        in.OS,
		Longitude: in.Longitude,
		Latitude:  in.Latitude,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func setCtxRequest(ctx context.Context, req any) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.SetRequest(req)
}

func badRequest(publicMsg string, err error) error {
	return &toddlerr.Error{
		PublicStatusCode:  status.BadRequest,
		ServiceStatusCode: status.BadRequest,
		PublicMessage:     publicMsg,
		ServiceMessage:    err.Error(),
	}
}
