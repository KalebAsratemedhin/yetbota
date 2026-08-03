package feed

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	feedSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/feed"
	postSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/post"
	"github.com/beka-birhanu/yetbota/content-service/internal/transport/http/shared"
)

type photoDTO struct {
	ID       string `json:"id"`
	PhotoURL string `json:"photo_url"`
	Position int    `json:"position"`
}

type coordDTO struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type postDTO struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Likes       int        `json:"likes"`
	Dislikes    int        `json:"dislikes"`
	Comments    int        `json:"comments"`
	UserID      string     `json:"user_id"`
	Tags        []string   `json:"tags"`
	IsQuestion  bool       `json:"is_question"`
	Photos      []photoDTO `json:"photos,omitempty"`
	Location    *coordDTO  `json:"location,omitempty"`
	Address     string     `json:"address,omitempty"`
	AttachedPostID string  `json:"attached_post_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func toDTO(p *dbmodels.Post, photos []*postSvc.OrderedPhoto) postDTO {
	if p == nil {
		return postDTO{}
	}
	var loc *coordDTO
	if p.Location.Valid && p.Location.Point != nil {
		coords := p.Location.Point.Coords()
		if len(coords) >= 2 {
			loc = &coordDTO{Latitude: coords[1], Longitude: coords[0]}
		}
	}
	phDTOs := make([]photoDTO, 0, len(photos))
	for _, ph := range photos {
		if ph == nil {
			continue
		}
		phDTOs = append(phDTOs, photoDTO{ID: ph.ID, PhotoURL: ph.URL, Position: ph.Position})
	}
	return postDTO{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		Likes:       p.Likes,
		Dislikes:    p.Dislikes,
		Comments:    p.CommentCount,
		UserID:      p.UserID,
		Tags:        []string(p.Tags),
		IsQuestion:  p.IsQuestion,
		Photos:      phDTOs,
		Location:    loc,
		Address:     p.Address.String,
		AttachedPostID: p.AttachedPostID.String,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func setCtxResponse(ctx context.Context, env shared.Envelope) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.Response = env
}

func encodeFeedMarkViewedHTTP(_ context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func encodeFeedGetHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*feedSvc.ListResponse)
	if !ok || out == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	posts := make([]postDTO, 0, len(out.Posts))
	for _, p := range out.Posts {
		posts = append(posts, toDTO(p, out.Photos[p.ID]))
	}

	env := shared.Envelope{
		Success: true,
		Data: map[string]any{
			"posts":       posts,
			"next_cursor": out.NextCursor,
		},
	}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}
