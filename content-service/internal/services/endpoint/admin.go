package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"

	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	adminSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/admin"
)

func sessFromCtx(ctx context.Context) (*ctxRP.Context, bool) {
	data := ctx.Value(ctxRP.AppSession)
	ctxSess, ok := data.(*ctxRP.Context)
	return ctxSess, ok
}

func makeAdminOverviewStatsEndpoint(svc adminSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		ctxSess, ok := sessFromCtx(ctx)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		ctxSess.Lv1("Incoming message AdminOverviewStats")
		respOK, respErr := svc.OverviewStats(ctx, ctxSess)
		ctxSess.Lv4()
		if respErr != nil {
			return respErr, nil
		}
		return respOK, nil
	}
}

func makeAdminOverviewGrowthEndpoint(svc adminSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		ctxSess, ok := sessFromCtx(ctx)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*adminSvc.OverviewGrowthRequest)
		if !ok {
			err := errors.New("error parse OverviewGrowthRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message AdminOverviewGrowth")
		respOK, respErr := svc.OverviewGrowth(ctx, ctxSess, r)
		ctxSess.Lv4()
		if respErr != nil {
			return respErr, nil
		}
		return respOK, nil
	}
}

func makeAdminUserStatsEndpoint(svc adminSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		ctxSess, ok := sessFromCtx(ctx)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		ctxSess.Lv1("Incoming message AdminUserStats")
		respOK, respErr := svc.UserStats(ctx, ctxSess)
		ctxSess.Lv4()
		if respErr != nil {
			return respErr, nil
		}
		return respOK, nil
	}
}

func makeAdminSystemAuditEndpoint(svc adminSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		ctxSess, ok := sessFromCtx(ctx)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*adminSvc.SystemAuditRequest)
		if !ok {
			err := errors.New("error parse SystemAuditRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message AdminSystemAudit")
		respOK, respErr := svc.SystemAudit(ctx, ctxSess, r)
		ctxSess.Lv4()
		if respErr != nil {
			return respErr, nil
		}
		return respOK, nil
	}
}
