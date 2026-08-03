package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"

	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	placefeedSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/placefeed"
)

func makePlaceFeedListEndpoint(svc placefeedSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*placefeedSvc.ListRequest)
		if !ok {
			err := errors.New("error parse PlaceFeed ListRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message PlaceFeedList")
		respOK, respErr := svc.List(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}
