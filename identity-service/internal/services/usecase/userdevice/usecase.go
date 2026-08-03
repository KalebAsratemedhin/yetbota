package userdevice

import (
	"context"

	"github.com/beka-birhanu/yetbota/identity-service/drivers/utils"
	ctxYB "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
)

func (s *svc) RegisterDevice(ctx context.Context, ctxSess *ctxYB.Context, req *RegisterDeviceRequest) (*RegisterDeviceResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	d := newFromRegisterRequest(req, ctxSess)

	created, err := s.deviceRepo.Add(ctx, nil, d)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &RegisterDeviceResponse{Device: created}, nil
}
