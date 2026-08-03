package moderation

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainModeration "github.com/beka-birhanu/yetbota/content-service/internal/domain/moderation"
	moderationSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/moderation"
	"github.com/beka-birhanu/yetbota/content-service/internal/transport/http/shared"
)

type caseDTO struct {
	ID              string     `json:"id"`
	ContentType     string     `json:"content_type"`
	ContentID       string     `json:"content_id"`
	ReportCount     int        `json:"report_count"`
	Status          string     `json:"status"`
	Severity        int        `json:"severity"`
	AutoHidden      bool       `json:"auto_hidden"`
	FirstReportedAt time.Time  `json:"first_reported_at"`
	LastReportedAt  time.Time  `json:"last_reported_at"`
	ResolvedBy      string     `json:"resolved_by,omitempty"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	Resolution      string     `json:"resolution,omitempty"`
	Version         int        `json:"version"`
}

type previewDTO struct {
	ContentType      string `json:"content_type"`
	ContentID        string `json:"content_id"`
	AuthorID         string `json:"author_id,omitempty"`
	Title            string `json:"title,omitempty"`
	Snippet          string `json:"snippet,omitempty"`
	ModerationStatus string `json:"moderation_status,omitempty"`
	Missing          bool   `json:"missing,omitempty"`
}

type caseViewDTO struct {
	Case    caseDTO     `json:"case"`
	Preview *previewDTO `json:"preview,omitempty"`
}

type reportDTO struct {
	ID         string    `json:"id"`
	ReporterID string    `json:"reporter_id"`
	Reason     string    `json:"reason"`
	Details    string    `json:"details,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type contentDTO struct {
	ID               string    `json:"id"`
	ContentType      string    `json:"content_type"`
	AuthorID         string    `json:"author_id"`
	Title            string    `json:"title,omitempty"`
	Body             string    `json:"body"`
	IsQuestion       bool      `json:"is_question,omitempty"`
	IsAnswer         bool      `json:"is_answer,omitempty"`
	ModerationStatus string    `json:"moderation_status"`
	CreatedAt        time.Time `json:"created_at"`
}

func setCtxResponse(ctx context.Context, env shared.Envelope) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.Response = env
}

func toCaseDTO(c *domainModeration.ModerationCase) caseDTO {
	return caseDTO{
		ID:              c.ID,
		ContentType:     c.ContentType,
		ContentID:       c.ContentID,
		ReportCount:     c.ReportCount,
		Status:          c.Status,
		Severity:        c.Severity,
		AutoHidden:      c.AutoHidden,
		FirstReportedAt: c.FirstReportedAt,
		LastReportedAt:  c.LastReportedAt,
		ResolvedBy:      c.ResolvedBy,
		ResolvedAt:      c.ResolvedAt,
		Resolution:      c.Resolution,
		Version:         c.Version,
	}
}

type casesData struct {
	Cases    []caseViewDTO `json:"cases"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

func encodeListCasesHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*moderationSvc.ListCasesResponse)
	if !ok || out == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	views := make([]caseViewDTO, 0, len(out.Cases))
	for _, v := range out.Cases {
		view := caseViewDTO{Case: toCaseDTO(v.Case)}
		if v.Preview != nil {
			view.Preview = &previewDTO{
				ContentType:      v.Preview.ContentType,
				ContentID:        v.Preview.ContentID,
				AuthorID:         v.Preview.AuthorID,
				Title:            v.Preview.Title,
				Snippet:          v.Preview.Snippet,
				ModerationStatus: v.Preview.ModerationStatus,
				Missing:          v.Preview.Missing,
			}
		}
		views = append(views, view)
	}
	env := shared.Envelope{Success: true, Data: casesData{Cases: views, Total: out.Total, Page: out.Page, PageSize: out.PageSize}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

type caseDetailData struct {
	Case    caseDTO     `json:"case"`
	Reports []reportDTO `json:"reports"`
	Content *contentDTO `json:"content,omitempty"`
}

func encodeGetCaseHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*moderationSvc.GetCaseResponse)
	if !ok || out == nil || out.Case == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	reports := make([]reportDTO, 0, len(out.Reports))
	for _, rp := range out.Reports {
		reports = append(reports, reportDTO{ID: rp.ID, ReporterID: rp.ReporterID, Reason: rp.Reason, Details: rp.Details, CreatedAt: rp.CreatedAt})
	}
	data := caseDetailData{Case: toCaseDTO(out.Case), Reports: reports}
	if out.Post != nil {
		data.Content = &contentDTO{
			ID:               out.Post.ID,
			ContentType:      domainModeration.ContentTypePost,
			AuthorID:         out.Post.UserID,
			Title:            out.Post.Title,
			Body:             out.Post.Description,
			IsQuestion:       out.Post.IsQuestion,
			ModerationStatus: out.Post.ModerationStatus,
			CreatedAt:        out.Post.CreatedAt,
		}
	} else if out.Comment != nil {
		data.Content = &contentDTO{
			ID:               out.Comment.ID,
			ContentType:      domainModeration.ContentTypeComment,
			AuthorID:         out.Comment.UserID,
			Body:             out.Comment.Content,
			IsAnswer:         out.Comment.IsAnswer,
			ModerationStatus: out.Comment.ModerationStatus,
			CreatedAt:        out.Comment.CreatedAt,
		}
	}
	env := shared.Envelope{Success: true, Data: data}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

type actData struct {
	CaseID     string `json:"case_id"`
	Status     string `json:"status"`
	Resolution string `json:"resolution"`
}

func encodeActHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*moderationSvc.ActOnCaseResponse)
	if !ok || out == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	env := shared.Envelope{Success: true, Data: actData{CaseID: out.CaseID, Status: out.Status, Resolution: out.Resolution}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}
