package placefeed

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	gkendpoint "github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	toddlerr "github.com/beka-birhanu/toddler/error"
	driverMW "github.com/beka-birhanu/yetbota/content-service/drivers/middleware"
	domainAuth "github.com/beka-birhanu/yetbota/content-service/internal/domain/auth"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainActivity "github.com/beka-birhanu/yetbota/content-service/internal/domain/placeactivity"
	"github.com/beka-birhanu/yetbota/content-service/internal/services/endpoint"
	placefeedSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/placefeed"
	"github.com/beka-birhanu/yetbota/content-service/internal/transport/http/shared"
)

type Config struct {
	E              *endpoint.Endpoints
	SessionManager domainAuth.SessionManager
	PathPrefix     string
}

func NewHandler(cfg *Config) (http.Handler, error) {
	mux := http.NewServeMux()
	if cfg != nil && cfg.E != nil && cfg.SessionManager != nil {
		auth := func(method, localPattern string) gkendpoint.Middleware {
			return driverMW.SelectAuthMiddleware(cfg.SessionManager, method+" "+cfg.PathPrefix+localPattern)
		}
		opts := shared.ServerOptions()
		mux.Handle("GET /", kithttp.NewServer(auth("GET", "/")(cfg.E.PlaceFeedList), decodePlaceFeedListHTTP, encodePlaceFeedListHTTP, opts...))
	}
	return mux, nil
}

func decodePlaceFeedListHTTP(ctx context.Context, r *http.Request) (any, error) {
	q := r.URL.Query()
	req := &placefeedSvc.ListRequest{}
	req.BookmarkedOnly = strings.EqualFold(strings.TrimSpace(q.Get("bookmarked_only")), "true")
	if v := strings.TrimSpace(q.Get("page")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			req.Page = n
		}
	}
	if v := strings.TrimSpace(q.Get("page_size")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			req.PageSize = n
		}
	}
	_ = req.Validate()
	if data := ctx.Value(ctxYB.AppSession); data != nil {
		if ctxSess, ok := data.(*ctxYB.Context); ok && ctxSess != nil {
			ctxSess.SetRequest(req)
		}
	}
	return req, nil
}

type eventDTO struct {
	ID        string    `json:"id"`
	PlaceID   string    `json:"place_id"`
	ActorID   string    `json:"actor_id,omitempty"`
	EventType string    `json:"event_type"`
	RefID     string    `json:"ref_id,omitempty"`
	Summary   string    `json:"summary"`
	CreatedAt time.Time `json:"created_at"`
}

func toEventDTO(e *domainActivity.Event) eventDTO {
	out := eventDTO{}
	if e == nil {
		return out
	}
	out = eventDTO{
		ID:        e.ID,
		PlaceID:   e.PlaceID,
		EventType: e.EventType,
		Summary:   e.Summary,
		CreatedAt: e.CreatedAt,
	}
	if e.ActorID.Valid {
		out.ActorID = e.ActorID.String
	}
	if e.RefID.Valid {
		out.RefID = e.RefID.String
	}
	return out
}

func encodePlaceFeedListHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*placefeedSvc.ListResponse)
	if !ok || out == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	items := make([]eventDTO, 0, len(out.Events))
	for _, e := range out.Events {
		items = append(items, toEventDTO(e))
	}
	env := shared.Envelope{
		Success: true,
		Data: map[string]any{
			"events":    items,
			"total":     out.Total,
			"page":      out.Page,
			"page_size": out.PageSize,
		},
	}
	if data := ctx.Value(ctxYB.AppSession); data != nil {
		if ctxSess, ok := data.(*ctxYB.Context); ok && ctxSess != nil {
			ctxSess.Response = env
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}
