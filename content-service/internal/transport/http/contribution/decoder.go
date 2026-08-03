package contribution

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	contributionSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/contribution"
)

func decodeContributionAddAnswerHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &contributionSvc.AddAnswerRequest{
		ContributionID: r.PathValue("id"),
		Content:        in.Content,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeContributionVoteHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		VoteType string `json:"vote_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &contributionSvc.VoteRequest{
		ContributionID: r.PathValue("id"),
		VoteType:       strings.ToLower(strings.TrimSpace(in.VoteType)),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeContributionAcceptAnswerHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		AnswerID string `json:"answer_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &contributionSvc.AcceptAnswerRequest{
		QuestionID: r.PathValue("id"),
		AnswerID:   strings.TrimSpace(in.AnswerID),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeOpenQuestionsListHTTP(ctx context.Context, r *http.Request) (any, error) {
	q := r.URL.Query()
	req := &contributionSvc.ListOpenQuestionsRequest{}
	if v := strings.TrimSpace(q.Get("page")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			req.Page = n
		}
	}
	if v := strings.TrimSpace(q.Get("page_size")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			req.PageSize = n
		}
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeOpenQuestionsNearHTTP(ctx context.Context, r *http.Request) (any, error) {
	q := r.URL.Query()
	req := &contributionSvc.ListOpenQuestionsRequest{}
	lat, err := strconv.ParseFloat(strings.TrimSpace(q.Get("near_lat")), 64)
	if err != nil {
		return nil, badRequest("near_lat is required", err)
	}
	lon, err := strconv.ParseFloat(strings.TrimSpace(q.Get("near_lon")), 64)
	if err != nil {
		return nil, badRequest("near_lon is required", err)
	}
	radius, err := strconv.ParseFloat(strings.TrimSpace(q.Get("radius_km")), 64)
	if err != nil {
		return nil, badRequest("radius_km is required", err)
	}
	req.NearLat = &lat
	req.NearLon = &lon
	req.RadiusKm = &radius
	if v := strings.TrimSpace(q.Get("page")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			req.Page = n
		}
	}
	if v := strings.TrimSpace(q.Get("page_size")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			req.PageSize = n
		}
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func setCtxRequest(ctx context.Context, req any) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.SetRequest(req)
}

func badRequest(publicMsg string, err error) error {
	return &toddlerr.Error{
		PublicStatusCode:  status.BadRequest,
		ServiceStatusCode: status.BadRequest,
		PublicMessage:     publicMsg,
		ServiceMessage:    err.Error(),
	}
}
