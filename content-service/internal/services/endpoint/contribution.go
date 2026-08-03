package endpoint

import (
	"context"
	"errors"

	"github.com/go-kit/kit/endpoint"

	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	contributionSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/contribution"
)

func makeContributionAddEndpoint(svc contributionSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*contributionSvc.AddRequest)
		if !ok {
			err := errors.New("error parse AddRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message ContributionAdd")
		respOK, respErr := svc.Add(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeContributionAddAnswerEndpoint(svc contributionSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*contributionSvc.AddAnswerRequest)
		if !ok {
			err := errors.New("error parse AddAnswerRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message ContributionAddAnswer")
		respOK, respErr := svc.AddAnswer(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeContributionVoteEndpoint(svc contributionSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*contributionSvc.VoteRequest)
		if !ok {
			err := errors.New("error parse VoteRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message ContributionVote")
		respOK, respErr := svc.Vote(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeContributionListByPlaceEndpoint(svc contributionSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*contributionSvc.ListRequest)
		if !ok {
			err := errors.New("error parse ListRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message ContributionListByPlace")
		respOK, respErr := svc.ListByPlace(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeContributionAcceptAnswerEndpoint(svc contributionSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*contributionSvc.AcceptAnswerRequest)
		if !ok {
			err := errors.New("error parse AcceptAnswerRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message ContributionAcceptAnswer")
		respOK, respErr := svc.AcceptAnswer(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}

func makeContributionListOpenQuestionsEndpoint(svc contributionSvc.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		data := ctx.Value(ctxRP.AppSession)
		ctxSess, ok := data.(*ctxRP.Context)
		if !ok {
			return errors.New("error parsing AppSession"), nil
		}
		r, ok := request.(*contributionSvc.ListOpenQuestionsRequest)
		if !ok {
			err := errors.New("error parse ListOpenQuestionsRequest")
			ctxSess.SetErrorMessage(err.Error())
			ctxSess.Lv4()
			return nil, err
		}
		ctxSess.SetRequest(r)
		ctxSess.Lv1("Incoming message ContributionListOpenQuestions")
		respOK, respErr := svc.ListOpenQuestions(ctx, ctxSess, r)
		if respErr != nil {
			ctxSess.Lv4()
			return respErr, nil
		}
		ctxSess.Lv4()
		return respOK, nil
	}
}
