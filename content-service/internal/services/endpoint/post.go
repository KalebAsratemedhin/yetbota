package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"

	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	postSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/post"
)

func makePostAddEndpoint(svc postSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*postSvc.AddRequest)
		if !ok {
			err := errors.New("error parse AddRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message PostAdd")

		respOK, respErr := svc.Add(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makePostReadEndpoint(svc postSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*postSvc.ReadRequest)
		if !ok {
			err := errors.New("error parse ReadRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message PostRead")

		respOK, respErr := svc.Read(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makePostUpdateEndpoint(svc postSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*postSvc.UpdateRequest)
		if !ok {
			err := errors.New("error parse UpdateRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message PostUpdate")

		respOK, respErr := svc.Update(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makePostVoteEndpoint(svc postSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*postSvc.PostVoteRequest)
		if !ok {
			err := errors.New("error parse PostVoteRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message PostVote")

		respOK, respErr := svc.Vote(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}


func makePostSaveEndpoint(svc postSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*postSvc.SaveRequest)
		if !ok {
			err := errors.New("error parse SaveRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message PostSave")

		respOK, respErr := svc.Save(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makePostUnsaveEndpoint(svc postSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*postSvc.UnsaveRequest)
		if !ok {
			err := errors.New("error parse UnsaveRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message PostUnsave")

		respOK, respErr := svc.Unsave(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}


func makePostListEndpoint(svc postSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*postSvc.ListRequest)
		if !ok {
			err := errors.New("error parse ListRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message PostList")

		respOK, respErr := svc.List(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}
