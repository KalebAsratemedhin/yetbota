package post

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	postSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/post"
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

type postDTO struct {
	ID               string            `json:"id"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Likes            int               `json:"likes"`
	Dislikes         int               `json:"dislikes"`
	Comments         int               `json:"comments"`
	UserID           string            `json:"user_id"`
	Tags             []string          `json:"tags"`
	IsQuestion       bool              `json:"is_question"`
	Photos           []orderedPhotoDTO `json:"photos,omitempty"`
	Location         *coordinateDTO    `json:"location,omitempty"`
	Address          string            `json:"address,omitempty"`
	AttachedPostID   string            `json:"attached_post_id,omitempty"`
	AttachedPost     *postDTO          `json:"attached_post,omitempty"`
	Interaction      *string           `json:"interaction,omitempty"`
	FollowingAuthor  *bool             `json:"following_author,omitempty"`
	Saved            *bool             `json:"saved,omitempty"`
	ModerationStatus string            `json:"moderation_status"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

func setCtxResponse(ctx context.Context, env shared.Envelope) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.Response = env
}

func toPostDTO(p *dbmodels.Post, photos []*postSvc.OrderedPhoto) postDTO {
	if p == nil {
		return postDTO{}
	}

	var loc *coordinateDTO
	if p.Location.Valid && p.Location.Point != nil {
		coords := p.Location.Point.Coords()
		if len(coords) >= 2 {
			loc = &coordinateDTO{Latitude: coords[1], Longitude: coords[0]}
		}
	}

	var outPhotos []orderedPhotoDTO
	if len(photos) > 0 {
		outPhotos = make([]orderedPhotoDTO, 0, len(photos))
		for _, ph := range photos {
			if ph == nil {
				continue
			}
			outPhotos = append(outPhotos, orderedPhotoDTO{
				ID:       ph.ID,
				PhotoURL: ph.URL,
				Position: ph.Position,
			})
		}
	}

	return postDTO{
		ID:               p.ID,
		Title:            p.Title,
		Description:      p.Description,
		Likes:            p.Likes,
		Dislikes:         p.Dislikes,
		Comments:         p.CommentCount,
		UserID:           p.UserID,
		Tags:             []string(p.Tags),
		IsQuestion:       p.IsQuestion,
		Photos:           outPhotos,
		Location:         loc,
		Address:          p.Address.String,
		AttachedPostID:   p.AttachedPostID.String,
		ModerationStatus: p.ModerationStatus,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}

type postData struct {
	Post postDTO `json:"post"`
}

func encodePostAddHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*postSvc.AddResponse)
	if !ok || out == nil || out.Post == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{Success: true, Data: postData{Post: toPostDTO(out.Post, out.Photos)}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodePostReadHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*postSvc.ReadResponse)
	if !ok || out == nil || out.Post == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	dto := toPostDTO(out.Post, out.Photos)
	if out.AttachedPost != nil {
		attached := toPostDTO(out.AttachedPost, out.AttachedPostPhotos)
		dto.AttachedPost = &attached
	}
	if out.Vote != nil {
		vt := out.Vote.VoteType
		dto.Interaction = &vt
	}
	if out.FollowingAuthor != nil {
		dto.FollowingAuthor = out.FollowingAuthor
	}
	if out.Saved != nil {
		dto.Saved = out.Saved
	}

	env := shared.Envelope{Success: true, Data: postData{Post: dto}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodePostUpdateHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*postSvc.UpdateResponse)
	if !ok || out == nil || out.Post == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{Success: true, Data: postData{Post: toPostDTO(out.Post, nil)}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

type voteData struct {
	Likes    int `json:"likes"`
	Dislikes int `json:"dislikes"`
}

func encodePostVoteHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*postSvc.PostVoteResponse)
	if !ok || out == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{Success: true, Data: voteData{Likes: out.Likes, Dislikes: out.Dislikes}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodePostSaveHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	if _, ok := resp.(*postSvc.SaveResponse); !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	env := shared.Envelope{Success: true}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodePostUnsaveHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	if _, ok := resp.(*postSvc.UnsaveResponse); !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	env := shared.Envelope{Success: true}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

type listData struct {
	Posts    []postDTO `json:"posts"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}

func encodePostListHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*postSvc.ListResponse)
	if !ok || out == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	posts := make([]postDTO, 0, len(out.Posts))
	for _, p := range out.Posts {
		dto := toPostDTO(p, out.Photos[p.ID])
		if v, ok := out.Votes[p.ID]; ok && v != nil {
			vt := v.VoteType
			dto.Interaction = &vt
		}
		if fa, ok := out.FollowingAuthor[p.UserID]; ok {
			b := fa
			dto.FollowingAuthor = &b
		}
		if sv, ok := out.Saved[p.ID]; ok {
			b := sv
			dto.Saved = &b
		}
		posts = append(posts, dto)
	}

	env := shared.Envelope{
		Success: true,
		Data: listData{
			Posts:    posts,
			Total:    out.Total,
			Page:     out.Page,
			PageSize: out.PageSize,
		},
	}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}
