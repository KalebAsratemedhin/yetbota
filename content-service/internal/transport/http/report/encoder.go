package report

import (
	"context"
	"encoding/json"
	"net/http"

	toddlerr "github.com/beka-birhanu/toddler/error"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	moderationSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/moderation"
	"github.com/beka-birhanu/yetbota/content-service/internal/transport/http/shared"
)

type reportData struct {
	CaseID      string `json:"case_id"`
	ReportCount int    `json:"report_count"`
	Status      string `json:"status"`
	AutoHidden  bool   `json:"auto_hidden"`
}

func setCtxResponse(ctx context.Context, env shared.Envelope) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.Response = env
}

func encodeReportCreateHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*moderationSvc.ReportResponse)
	if !ok || out == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	env := shared.Envelope{Success: true, Data: reportData{
		CaseID:      out.CaseID,
		ReportCount: out.ReportCount,
		Status:      out.Status,
		AutoHidden:  out.AutoHidden,
	}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}
