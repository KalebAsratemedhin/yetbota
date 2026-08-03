package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"

	ctxYB "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
	deviceSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/userdevice"
)

func makeUserDeviceRegisterEndpoint(svc deviceSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxYB.AppSession)
		ctxSess, ok := data.(*ctxYB.Context)
		if !ok {
			err := errors.New("error parsing AppSession")
			return err, nil
		}
		r, ok := request.(*deviceSvc.RegisterDeviceRequest)
		if !ok {
			err := errors.New("error parse RegisterDeviceRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message UserDeviceRegister")

		respOK, respErr := svc.RegisterDevice(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}
