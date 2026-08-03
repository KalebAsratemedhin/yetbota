package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"

	ctxYP "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
	authSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/auth"
)

func makeLoginEndpoint(svc authSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxYP.AppSession)
		ctxSess, ok := data.(*ctxYP.Context)
		if !ok {
			err := errors.New("error parsing AppSession")
			return err, nil
		}
		r, ok := request.(*authSvc.LoginRequest)
		if !ok {
			err := errors.New("error parse LoginRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message Login")

		respOK, respErr := svc.Login(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeRefreshEndpoint(svc authSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxYP.AppSession)
		ctxSess, ok := data.(*ctxYP.Context)
		if !ok {
			err := errors.New("error parsing AppSession")
			return err, nil
		}
		r, ok := request.(*authSvc.RefreshRequest)
		if !ok {
			err := errors.New("error parse RefreshRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message Refresh")

		respOK, respErr := svc.Refresh(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeLogoutEndpoint(svc authSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxYP.AppSession)
		ctxSess, ok := data.(*ctxYP.Context)
		if !ok {
			err := errors.New("error parsing AppSession")
			return err, nil
		}
		r, ok := request.(*authSvc.LogoutRequest)
		if !ok {
			err := errors.New("error parse LogoutRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message Logout")

		respOK, respErr := svc.Logout(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeAuthorizationEndpoint(svc authSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxYP.AppSession)
		ctxSess, ok := data.(*ctxYP.Context)
		if !ok {
			err := errors.New("error parsing AppSession")
			return err, nil
		}
		r, ok := request.(*authSvc.AuthorizationRequest)
		if !ok {
			err := errors.New("error parse AuthorizationRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message Authorization")

		respOK, respErr := svc.Authorization(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeChangePasswordEndpoint(svc authSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxYP.AppSession)
		ctxSess, ok := data.(*ctxYP.Context)
		if !ok {
			err := errors.New("error parsing AppSession")
			return err, nil
		}
		r, ok := request.(*authSvc.ChangePasswordRequest)
		if !ok {
			err := errors.New("error parse ChangePasswordRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message ChangePassword")

		respOK, respErr := svc.ChangePassword(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}
