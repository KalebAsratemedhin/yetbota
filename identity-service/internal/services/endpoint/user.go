package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"

	ctxRP "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
	userSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/user"
)

func makeUserListEndpoint(svc userSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			err := errors.New("error parsing AppSession")
			return err, nil
		}
		r, ok := request.(*userSvc.ListRequest)
		if !ok {
			err := errors.New("error parse ListRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message UserList")

		respOK, respErr := svc.List(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeUserReadEndpoint(svc userSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			err := errors.New("error parsing AppSession")
			return err, nil
		}
		r, ok := request.(*userSvc.ReadRequest)
		if !ok {
			err := errors.New("error parse ReadRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message UserRead")

		respOK, respErr := svc.Read(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeUserReadPublicEndpoint(svc userSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			err := errors.New("error parsing AppSession")
			return err, nil
		}
		r, ok := request.(*userSvc.ReadPublicRequest)
		if !ok {
			err := errors.New("error parse ReadPublicRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message UserReadPublic")

		respOK, respErr := svc.ReadPublic(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeUserUpdateEndpoint(svc userSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			err := errors.New("error parsing AppSession")
			return err, nil
		}
		r, ok := request.(*userSvc.UpdateRequest)
		if !ok {
			err := errors.New("error parse UpdateRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message UserUpdate")

		respOK, respErr := svc.Update(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeUserUpdateSelfEndpoint(svc userSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			err := errors.New("error parsing AppSession")
			return err, nil
		}
		r, ok := request.(*userSvc.UpdateSelfRequest)
		if !ok {
			err := errors.New("error parse UpdateSelfRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}

		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message UserUpdateSelf")

		respOK, respErr := svc.UpdateSelf(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeUserRegisterEndpoint(svc userSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			err := errors.New("error parsing AppSession")
			return err, nil
		}
		r, ok := request.(*userSvc.RegisterRequest)
		if !ok {
			err := errors.New("error parse RegisterRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message UserRegister")

		respOK, respErr := svc.Register(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeUserDeleteEndpoint(svc userSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			err := errors.New("error parsing AppSession")
			return err, nil
		}
		r, ok := request.(*userSvc.DeleteRequest)
		if !ok {
			err := errors.New("error parse DeleteRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message UserDelete")

		respOK, respErr := svc.Delete(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeUserDeleteSelfEndpoint(svc userSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			err := errors.New("error parsing AppSession")
			return err, nil
		}
		r, ok := request.(*userSvc.DeleteSelfRequest)
		if !ok {
			err := errors.New("error parse DeleteSelfRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}

		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message UserDeleteSelf")

		respOK, respErr := svc.DeleteSelf(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeUserUploadProfileEndpoint(svc userSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			err := errors.New("error parsing AppSession")
			return err, nil
		}
		r, ok := request.(*userSvc.UploadProfileRequest)
		if !ok {
			err := errors.New("error parse UploadProfileRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}

		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message UserUploadProfile")

		respOK, respErr := svc.UploadProfile(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeUserFollowEndpoint(svc userSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			err := errors.New("error parsing AppSession")
			return err, nil
		}
		r, ok := request.(*userSvc.FollowRequest)
		if !ok {
			err := errors.New("error parse FollowRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message UserFollow")

		respOK, respErr := svc.Follow(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeUserUnfollowEndpoint(svc userSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			err := errors.New("error parsing AppSession")
			return err, nil
		}
		r, ok := request.(*userSvc.UnfollowRequest)
		if !ok {
			err := errors.New("error parse UnfollowRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message UserUnfollow")

		respOK, respErr := svc.Unfollow(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}
