package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	notificationSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/notification"
)

func decodeNotificationListHTTP(ctx context.Context, r *http.Request) (any, error) {
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	page, _ := strconv.Atoi(q.Get("page"))

	req := &notificationSvc.ListRequest{
		Unread: strings.EqualFold(q.Get("unread"), "true"),
		Limit:  int32(limit),
		Page:   int32(page),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeNotificationMarkAsReadHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &notificationSvc.MarkAsReadRequest{IDs: in.IDs}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeNotificationDeleteHTTP(ctx context.Context, r *http.Request) (any, error) {
	req := &notificationSvc.DeleteRequest{ID: r.PathValue("id")}
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
