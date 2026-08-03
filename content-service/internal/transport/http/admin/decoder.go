package admin

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	adminSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/admin"
)

func decodeNoParams(ctx context.Context, r *http.Request) (any, error) {
	return nil, nil
}

func decodeOverviewGrowthHTTP(ctx context.Context, r *http.Request) (any, error) {
	q := r.URL.Query()
	req := &adminSvc.OverviewGrowthRequest{
		Metric: strings.ToLower(strings.TrimSpace(q.Get("metric"))),
		Range:  strings.ToLower(strings.TrimSpace(q.Get("range"))),
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeSystemAuditHTTP(ctx context.Context, r *http.Request) (any, error) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	req := &adminSvc.SystemAuditRequest{
		ActionType: strings.ToUpper(strings.TrimSpace(q.Get("action_type"))),
		Actor:      strings.TrimSpace(q.Get("actor")),
		From:       strings.TrimSpace(q.Get("from")),
		To:         strings.TrimSpace(q.Get("to")),
		Page:       page,
		PageSize:   pageSize,
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
