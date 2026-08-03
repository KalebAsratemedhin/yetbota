package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	notificationSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/notification"
	"github.com/beka-birhanu/yetbota/content-service/internal/transport/http/shared"
)

type notificationDTO struct {
	ID         string          `json:"id"`
	UserID     string          `json:"user_id"`
	Title      string          `json:"title"`
	Body       string          `json:"body"`
	Data       json.RawMessage `json:"data,omitempty"`
	Attachment string          `json:"attachment,omitempty"`
	SentAt     time.Time       `json:"sent_at"`
	ReadAt     *time.Time      `json:"read_at,omitempty"`
}

type paginationDTO struct {
	Page       int32 `json:"page"`
	Total      int32 `json:"total"`
	Length     int32 `json:"length"`
	TotalPages int32 `json:"total_pages"`
}

func setCtxResponse(ctx context.Context, env shared.Envelope) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.Response = env
}

func toNotificationDTO(n *dbmodels.Notification) notificationDTO {
	if n == nil {
		return notificationDTO{}
	}
	out := notificationDTO{
		ID:     n.ID,
		UserID: n.UserID,
		Title:  n.Title,
		Body:   n.Body,
		SentAt: n.SentAt,
	}
	if len(n.Data) > 0 {
		out.Data = json.RawMessage(n.Data)
	}
	if n.Attachment.Valid {
		out.Attachment = n.Attachment.String
	}
	if n.ReadAt.Valid {
		t := n.ReadAt.Time
		out.ReadAt = &t
	}
	return out
}

type notificationListData struct {
	Notifications []notificationDTO `json:"notifications"`
	Pagination    *paginationDTO    `json:"pagination,omitempty"`
}

func encodeNotificationListHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*notificationSvc.ListResponse)
	if !ok || out == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	notifications := make([]notificationDTO, 0, len(out.Notifications))
	for _, n := range out.Notifications {
		notifications = append(notifications, toNotificationDTO(n))
	}

	var pagination *paginationDTO
	if out.Pagination != nil {
		pagination = &paginationDTO{
			Page:       out.Pagination.Page,
			Total:      out.Pagination.Total,
			Length:     out.Pagination.Length,
			TotalPages: out.Pagination.TotalPages,
		}
	}

	env := shared.Envelope{Success: true, Data: notificationListData{Notifications: notifications, Pagination: pagination}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

type markAsReadData struct {
	Success []string `json:"success"`
	Failure []string `json:"failure"`
}

func encodeNotificationMarkAsReadHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*notificationSvc.MarkAsReadResponse)
	if !ok || out == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	data := markAsReadData{Success: out.Success, Failure: out.Failure}
	if data.Success == nil {
		data.Success = []string{}
	}
	if data.Failure == nil {
		data.Failure = []string{}
	}

	env := shared.Envelope{Success: true, Data: data}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeNotificationDeleteHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	env := shared.Envelope{Success: true}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}
