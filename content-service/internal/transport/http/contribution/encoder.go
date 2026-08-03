package contribution

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainContribution "github.com/beka-birhanu/yetbota/content-service/internal/domain/contribution"
	contributionSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/contribution"
	"github.com/beka-birhanu/yetbota/content-service/internal/transport/http/shared"
)

type contributionDTO struct {
	ID               string     `json:"id"`
	PlaceID          string     `json:"place_id"`
	UserID           string     `json:"user_id"`
	Kind             string     `json:"kind"`
	Content          string     `json:"content"`
	ParentID         string     `json:"parent_id,omitempty"`
	Upvotes          int        `json:"upvotes"`
	Downvotes        int        `json:"downvotes"`
	ModerationStatus string     `json:"moderation_status"`
	Origin           string     `json:"origin"`
	SourcePostID     string     `json:"source_post_id,omitempty"`
	AcceptedAt       *time.Time `json:"accepted_at,omitempty"`
	AcceptedBy       string     `json:"accepted_by,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func setCtxResponse(ctx context.Context, env shared.Envelope) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.Response = env
}

func toContributionDTO(c *domainContribution.Contribution) contributionDTO {
	if c == nil {
		return contributionDTO{}
	}
	out := contributionDTO{
		ID:               c.ID,
		PlaceID:          c.PlaceID,
		UserID:           c.UserID,
		Kind:             c.Kind,
		Content:          c.Content,
		Upvotes:          c.Upvotes,
		Downvotes:        c.Downvotes,
		ModerationStatus: c.ModerationStatus,
		Origin:           c.Origin,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
	}
	if c.ParentID.Valid {
		out.ParentID = c.ParentID.String
	}
	if c.SourcePostID.Valid {
		out.SourcePostID = c.SourcePostID.String
	}
	if c.AcceptedAt.Valid {
		t := c.AcceptedAt.Time
		out.AcceptedAt = &t
	}
	if c.AcceptedBy.Valid {
		out.AcceptedBy = c.AcceptedBy.String
	}
	return out
}

func encodeContributionAddAnswerHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*contributionSvc.AddResponse)
	if !ok || out == nil || out.Contribution == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	env := shared.Envelope{Success: true, Data: map[string]any{"contribution": toContributionDTO(out.Contribution)}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeContributionVoteHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*contributionSvc.VoteResponse)
	if !ok || out == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	env := shared.Envelope{Success: true, Data: map[string]any{"upvotes": out.Upvotes, "downvotes": out.Downvotes}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeContributionAcceptAnswerHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*contributionSvc.AcceptAnswerResponse)
	if !ok || out == nil || out.Answer == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	env := shared.Envelope{Success: true, Data: map[string]any{"answer": toContributionDTO(out.Answer)}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeOpenQuestionsListHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*contributionSvc.ListOpenQuestionsResponse)
	if !ok || out == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	items := make([]contributionDTO, 0, len(out.Questions))
	for _, q := range out.Questions {
		items = append(items, toContributionDTO(q))
	}
	env := shared.Envelope{
		Success: true,
		Data: map[string]any{
			"questions": items,
			"total":     out.Total,
			"page":      out.Page,
			"page_size": out.PageSize,
		},
	}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}
