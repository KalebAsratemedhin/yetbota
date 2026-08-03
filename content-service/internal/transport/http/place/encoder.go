package place

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainContribution "github.com/beka-birhanu/yetbota/content-service/internal/domain/contribution"
	domainPlace "github.com/beka-birhanu/yetbota/content-service/internal/domain/place"
	contributionSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/contribution"
	placeSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/place"
	"github.com/beka-birhanu/yetbota/content-service/internal/transport/http/shared"
)

type coordinateDTO struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type orderedPhotoDTO struct {
	ID       string `json:"id"`
	PhotoURL string `json:"photo_url"`
	Position int    `json:"position"`
}

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
	UserVote         *string    `json:"user_vote,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type questionThreadDTO struct {
	Question       contributionDTO   `json:"question"`
	Answers        []contributionDTO `json:"answers"`
	AcceptedAnswer *contributionDTO  `json:"accepted_answer,omitempty"`
	IsResolved     bool              `json:"is_resolved"`
}

type placeDTO struct {
	ID               string                         `json:"id"`
	Name             string                         `json:"name"`
	Description      string                         `json:"description"`
	Address          string                         `json:"address,omitempty"`
	Location         *coordinateDTO                 `json:"location,omitempty"`
	Status           string                         `json:"status"`
	ConfidenceScore  int16                          `json:"confidence_score"`
	LastActivityAt   *time.Time                     `json:"last_activity_at,omitempty"`
	LastConfirmedAt  *time.Time                     `json:"last_confirmed_at,omitempty"`
	NeedsCheck       bool                           `json:"needs_check"`
	IsBookmarked     bool                           `json:"is_bookmarked,omitempty"`
	SummaryText      string                         `json:"summary_text,omitempty"`
	SummaryGeneratedAt *time.Time                   `json:"summary_generated_at,omitempty"`
	SummaryCitations []any                          `json:"summary_citations,omitempty"`
	CreatedBy        string                         `json:"created_by"`
	Photos           []orderedPhotoDTO              `json:"photos,omitempty"`
	Contributions    map[string][]contributionDTO   `json:"contributions,omitempty"`
	Questions        []questionThreadDTO            `json:"questions,omitempty"`
	CreatedAt        time.Time                      `json:"created_at"`
	UpdatedAt        time.Time                      `json:"updated_at"`
}

func setCtxResponse(ctx context.Context, env shared.Envelope) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.Response = env
}

func toPlaceDTO(p *domainPlace.Place, photos []*placeSvc.OrderedPhoto) placeDTO {
	out := placeDTO{}
	if p == nil {
		return out
	}
	out = placeDTO{
		ID:              p.ID,
		Name:            p.Name,
		Description:     p.Description,
		Address:         p.Address.String,
		Status:          p.Status,
		ConfidenceScore: p.ConfidenceScore,
		NeedsCheck:      p.NeedsCheck,
		CreatedBy:       p.CreatedBy,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
	if p.LastActivityAt.Valid {
		t := p.LastActivityAt.Time
		out.LastActivityAt = &t
	}
	if p.LastConfirmedAt.Valid {
		t := p.LastConfirmedAt.Time
		out.LastConfirmedAt = &t
	}
	if p.SummaryText.Valid {
		out.SummaryText = p.SummaryText.String
	}
	if p.SummaryGeneratedAt.Valid {
		t := p.SummaryGeneratedAt.Time
		out.SummaryGeneratedAt = &t
	}
	if len(p.SummaryCitations) > 0 {
		var citations []any
		if err := json.Unmarshal(p.SummaryCitations, &citations); err == nil {
			out.SummaryCitations = citations
		}
	}
	if p.Location.Valid && p.Location.Point != nil {
		coords := p.Location.Point.Coords()
		if len(coords) >= 2 {
			out.Location = &coordinateDTO{Latitude: coords[1], Longitude: coords[0]}
		}
	}
	if len(photos) > 0 {
		out.Photos = make([]orderedPhotoDTO, 0, len(photos))
		for _, ph := range photos {
			if ph == nil {
				continue
			}
			out.Photos = append(out.Photos, orderedPhotoDTO{ID: ph.ID, PhotoURL: ph.URL, Position: ph.Position})
		}
	}
	return out
}

func toContributionDTO(c *domainContribution.Contribution, vote *string) contributionDTO {
	out := contributionDTO{}
	if c == nil {
		return out
	}
	out = contributionDTO{
		ID:               c.ID,
		PlaceID:          c.PlaceID,
		UserID:           c.UserID,
		Kind:             c.Kind,
		Content:          c.Content,
		Upvotes:          c.Upvotes,
		Downvotes:        c.Downvotes,
		ModerationStatus: c.ModerationStatus,
		Origin:           c.Origin,
		UserVote:         vote,
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

func toQuestionThreadDTO(t *placeSvc.QuestionThread) questionThreadDTO {
	out := questionThreadDTO{Answers: []contributionDTO{}}
	if t == nil {
		return out
	}
	if t.Question != nil {
		out.Question = toContributionDTO(t.Question.Contribution, t.Question.UserVote)
	}
	for _, a := range t.Answers {
		out.Answers = append(out.Answers, toContributionDTO(a.Contribution, a.UserVote))
	}
	if t.AcceptedAnswer != nil {
		dto := toContributionDTO(t.AcceptedAnswer.Contribution, t.AcceptedAnswer.UserVote)
		out.AcceptedAnswer = &dto
	}
	out.IsResolved = t.IsResolved
	return out
}

func encodePlaceAddHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*placeSvc.AddResponse)
	if !ok || out == nil || out.Place == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	env := shared.Envelope{Success: true, Data: map[string]any{"place": toPlaceDTO(out.Place, out.Photos)}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodePlaceReadHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*placeSvc.ReadResponse)
	if !ok || out == nil || out.Place == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	placeData := toPlaceDTO(out.Place, out.Photos)
	placeData.IsBookmarked = out.IsBookmarked
	grouped := make(map[string][]contributionDTO, len(out.ContributionsByKind))
	for kind, items := range out.ContributionsByKind {
		list := make([]contributionDTO, 0, len(items))
		for _, item := range items {
			list = append(list, toContributionDTO(item.Contribution, item.UserVote))
		}
		grouped[kind] = list
	}
	placeData.Contributions = grouped
	if len(out.Questions) > 0 {
		threads := make([]questionThreadDTO, 0, len(out.Questions))
		for _, q := range out.Questions {
			threads = append(threads, toQuestionThreadDTO(q))
		}
		placeData.Questions = threads
	}
	env := shared.Envelope{Success: true, Data: map[string]any{"place": placeData}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodePlaceBookmarkHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	env := shared.Envelope{Success: true, Data: map[string]any{}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodePlaceConfirmHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*placeSvc.ConfirmResponse)
	if !ok || out == nil || out.Place == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	env := shared.Envelope{Success: true, Data: map[string]any{"place": toPlaceDTO(out.Place, nil)}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodePlaceReportProblemHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*placeSvc.ReportProblemResponse)
	if !ok || out == nil || out.Report == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	data := map[string]any{
		"id":         out.Report.ID,
		"place_id":   out.Report.PlaceID,
		"kind":       out.Report.Kind,
		"status":     out.Report.Status,
		"created_at": out.Report.CreatedAt,
	}
	if out.Report.Note.Valid {
		data["note"] = out.Report.Note.String
	}
	if out.Report.DuplicateOf.Valid {
		data["duplicate_of"] = out.Report.DuplicateOf.String
	}
	env := shared.Envelope{Success: true, Data: map[string]any{"report": data}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodePlaceListHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*placeSvc.ListResponse)
	if !ok || out == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	items := make([]placeDTO, 0, len(out.Places))
	for _, p := range out.Places {
		items = append(items, toPlaceDTO(p, out.Photos[p.ID]))
	}
	env := shared.Envelope{
		Success: true,
		Data: map[string]any{
			"places":     items,
			"total":      out.Total,
			"page":       out.Page,
			"page_size":  out.PageSize,
		},
	}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodePlaceUpdateHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*placeSvc.UpdateResponse)
	if !ok || out == nil || out.Place == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	env := shared.Envelope{Success: true, Data: map[string]any{"place": toPlaceDTO(out.Place, nil)}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeContributionAddHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*contributionSvc.AddResponse)
	if !ok || out == nil || out.Contribution == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	env := shared.Envelope{Success: true, Data: map[string]any{"contribution": toContributionDTO(out.Contribution, nil)}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeContributionListHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*contributionSvc.ListResponse)
	if !ok || out == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	items := make([]contributionDTO, 0, len(out.Items))
	for _, item := range out.Items {
		var vote *string
		if item.UserVote != nil {
			vt := item.UserVote.VoteType
			vote = &vt
		}
		items = append(items, toContributionDTO(item.Contribution, vote))
	}
	env := shared.Envelope{
		Success: true,
		Data: map[string]any{
			"contributions": items,
			"total":         out.Total,
			"page":          out.Page,
			"page_size":     out.PageSize,
		},
	}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}
