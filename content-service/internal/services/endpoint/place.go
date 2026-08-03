package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"

	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	placeSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/place"
)

func makePlaceAddEndpoint(svc placeSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*placeSvc.AddRequest)
		if !ok {
			err := errors.New("error parse AddRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message PlaceAdd")
		respOK, respErr := svc.Add(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makePlaceReadEndpoint(svc placeSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*placeSvc.ReadRequest)
		if !ok {
			err := errors.New("error parse ReadRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message PlaceRead")
		respOK, respErr := svc.Read(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makePlaceListEndpoint(svc placeSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*placeSvc.ListRequest)
		if !ok {
			err := errors.New("error parse ListRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message PlaceList")
		respOK, respErr := svc.List(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makePlaceUpdateEndpoint(svc placeSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*placeSvc.UpdateRequest)
		if !ok {
			err := errors.New("error parse UpdateRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message PlaceUpdate")
		respOK, respErr := svc.Update(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makePlaceBookmarkEndpoint(svc placeSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*placeSvc.BookmarkRequest)
		if !ok {
			err := errors.New("error parse BookmarkRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message PlaceBookmark")
		respOK, respErr := svc.Bookmark(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makePlaceUnbookmarkEndpoint(svc placeSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*placeSvc.UnbookmarkRequest)
		if !ok {
			err := errors.New("error parse UnbookmarkRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message PlaceUnbookmark")
		respOK, respErr := svc.Unbookmark(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makePlaceConfirmEndpoint(svc placeSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*placeSvc.ConfirmRequest)
		if !ok {
			err := errors.New("error parse ConfirmRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message PlaceConfirm")
		respOK, respErr := svc.Confirm(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makePlaceReportProblemEndpoint(svc placeSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*placeSvc.ReportProblemRequest)
		if !ok {
			err := errors.New("error parse ReportProblemRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message PlaceReportProblem")
		respOK, respErr := svc.ReportProblem(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}
