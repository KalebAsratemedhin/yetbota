package endpoint

import (
	"context"

	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	feedSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/feed"
	"github.com/go-kit/kit/endpoint"
)

func makeFeedGetEndpoint(svc feedSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		ctxSess := ctx.Value(ctxRP.AppSession).(*ctxRP.Context)
		r := request.(*feedSvc.ListRequest)
		ctxSess.Lv1("Incoming message FeedGet")
		resp, err := svc.List(ctx, ctxSess, r)
		if err != nil {
			ctxSess.Lv4()
			return err, nil
		}
		ctxSess.Lv4()
		return resp, nil
	}
}

func makeFeedMarkViewedEndpoint(svc feedSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		ctxSess := ctx.Value(ctxRP.AppSession).(*ctxRP.Context)
		r := request.(*feedSvc.MarkViewedRequest)
		ctxSess.Lv1("Incoming message FeedMarkViewed")
		if err := svc.MarkViewed(ctx, ctxSess, r); err != nil {
			ctxSess.Lv4()
			return err, nil
		}
		ctxSess.Lv4()
		return nil, nil
	}
}
