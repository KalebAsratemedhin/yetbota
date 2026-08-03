package moderation

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	moderationSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/moderation"
)

func decodeListCasesHTTP(ctx context.Context, r *http.Request) (any, error) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))

	req := &moderationSvc.ListCasesRequest{
		Status:      strings.ToUpper(strings.TrimSpace(q.Get("status"))),
		Reason:      strings.ToUpper(strings.TrimSpace(q.Get("reason"))),
		ContentType: strings.ToUpper(strings.TrimSpace(q.Get("content_type"))),
		Page:        page,
		PageSize:    pageSize,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeGetCaseHTTP(ctx context.Context, r *http.Request) (any, error) {
	req := &moderationSvc.GetCaseRequest{ID: r.PathValue("id")}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeActHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		Action    string `json:"action"`
		Note      string `json:"note"`
		BanReason string `json:"ban_reason"`
		Version   int    `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &moderationSvc.ActOnCaseRequest{
		CaseID:    r.PathValue("id"),
		Action:    strings.ToUpper(strings.TrimSpace(in.Action)),
		Note:      strings.TrimSpace(in.Note),
		BanReason: strings.TrimSpace(in.BanReason),
		Version:   in.Version,
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
