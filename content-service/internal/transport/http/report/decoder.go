package report

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	moderationSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/moderation"
)

func decodeReportCreateHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		ContentType string `json:"content_type"`
		ContentID   string `json:"content_id"`
		Reason      string `json:"reason"`
		Details     string `json:"details"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &moderationSvc.ReportRequest{
		ContentType: strings.ToUpper(strings.TrimSpace(in.ContentType)),
		ContentID:   strings.TrimSpace(in.ContentID),
		Reason:      strings.ToUpper(strings.TrimSpace(in.Reason)),
		Details:     strings.TrimSpace(in.Details),
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
