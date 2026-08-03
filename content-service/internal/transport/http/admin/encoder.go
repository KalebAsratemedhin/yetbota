package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	adminSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/admin"
	"github.com/beka-birhanu/yetbota/content-service/internal/transport/http/shared"
)

func setCtxResponse(ctx context.Context, env shared.Envelope) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.Response = env
}

func writeEnvelope(ctx context.Context, w http.ResponseWriter, data any) error {
	env := shared.Envelope{Success: true, Data: data}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func serverError(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusInternalServerError)
	return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
}

type statCardDTO struct {
	Value     int64  `json:"value"`
	ChangePct *int   `json:"change_pct,omitempty"`
	Direction string `json:"direction,omitempty"`
}

type overviewStatsData struct {
	TotalUsers     statCardDTO `json:"total_users"`
	TotalQuestions statCardDTO `json:"total_questions"`
	TotalLocations statCardDTO `json:"total_locations"`
}

func encodeOverviewStatsHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*adminSvc.OverviewStatsResponse)
	if !ok || out == nil || out.TotalUsers == nil || out.TotalQuestions == nil || out.TotalLocations == nil {
		return serverError(w)
	}
	data := overviewStatsData{
		TotalUsers:     statCardDTO{Value: out.TotalUsers.Value, ChangePct: out.TotalUsers.ChangePct, Direction: out.TotalUsers.Direction},
		TotalQuestions: statCardDTO{Value: out.TotalQuestions.Value, ChangePct: out.TotalQuestions.ChangePct, Direction: out.TotalQuestions.Direction},
		TotalLocations: statCardDTO{Value: out.TotalLocations.Value, ChangePct: out.TotalLocations.ChangePct, Direction: out.TotalLocations.Direction},
	}
	return writeEnvelope(ctx, w, data)
}

type growthPointDTO struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

type growthData struct {
	Metric     string           `json:"metric"`
	Range      string           `json:"range"`
	Total      int64            `json:"total"`
	DeltaLabel string           `json:"delta_label"`
	Points     []growthPointDTO `json:"points"`
}

func encodeOverviewGrowthHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*adminSvc.OverviewGrowthResponse)
	if !ok || out == nil {
		return serverError(w)
	}
	points := make([]growthPointDTO, 0, len(out.Points))
	for _, p := range out.Points {
		points = append(points, growthPointDTO{Label: p.Label, Value: p.Value})
	}
	return writeEnvelope(ctx, w, growthData{
		Metric:     out.Metric,
		Range:      out.Range,
		Total:      out.Total,
		DeltaLabel: out.DeltaLabel,
		Points:     points,
	})
}

type joinedCardDTO struct {
	Value     int64 `json:"value"`
	ChangePct *int  `json:"change_pct,omitempty"`
}

type reputationCardDTO struct {
	Value     int64 `json:"value"`
	Threshold int64 `json:"threshold"`
}

type userStatsData struct {
	NewlyJoinedToday   joinedCardDTO     `json:"newly_joined_today"`
	HighReputationUser reputationCardDTO `json:"high_reputation_users"`
}

func encodeUserStatsHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*adminSvc.UserStatsResponse)
	if !ok || out == nil {
		return serverError(w)
	}
	return writeEnvelope(ctx, w, userStatsData{
		NewlyJoinedToday:   joinedCardDTO{Value: out.NewlyJoinedToday, ChangePct: out.NewlyJoinedChange},
		HighReputationUser: reputationCardDTO{Value: out.HighReputation, Threshold: out.Threshold},
	})
}

type auditActorDTO struct {
	ID      string `json:"id"`
	Display string `json:"display"`
}

type auditEntryDTO struct {
	ID         string        `json:"id"`
	Timestamp  time.Time     `json:"timestamp"`
	Actor      auditActorDTO `json:"actor"`
	ActionType string        `json:"action_type"`
	Details    string        `json:"details,omitempty"`
}

type auditData struct {
	Entries  []auditEntryDTO `json:"entries"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

func encodeSystemAuditHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*adminSvc.SystemAuditResponse)
	if !ok || out == nil {
		return serverError(w)
	}
	entries := make([]auditEntryDTO, 0, len(out.Entries))
	for _, e := range out.Entries {
		display := e.ActorName
		if display == "" {
			display = e.ActorID
		}
		entries = append(entries, auditEntryDTO{
			ID:         e.ID,
			Timestamp:  e.Timestamp,
			Actor:      auditActorDTO{ID: e.ActorID, Display: display},
			ActionType: e.ActionType,
			Details:    e.Details,
		})
	}
	return writeEnvelope(ctx, w, auditData{Entries: entries, Total: out.Total, Page: out.Page, PageSize: out.PageSize})
}
