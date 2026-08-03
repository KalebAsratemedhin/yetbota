package comment

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	commentSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/comment"
	"github.com/beka-birhanu/yetbota/content-service/internal/transport/http/shared"
)

type commentDTO struct {
	ID        string    `json:"id"`
	Comment   string    `json:"comment"`
	Upvote    int       `json:"upvote"`
	Downvote  int       `json:"downvote"`
	UserID    string    `json:"user_id"`
	PostID    string    `json:"post_id"`
	IsAnswer  bool      `json:"is_answer"`
	CommentID string    `json:"comment_id,omitempty"`
	UserVote  *string   `json:"user_vote,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func setCtxResponse(ctx context.Context, env shared.Envelope) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.Response = env
}

func toCommentDTO(c *dbmodels.Comment) commentDTO {
	if c == nil {
		return commentDTO{}
	}
	out := commentDTO{
		ID:        c.ID,
		Comment:   c.Content,
		Upvote:    c.Upvote,
		Downvote:  c.Downvote,
		UserID:    c.UserID,
		PostID:    c.PostID,
		IsAnswer:  c.IsAnswer,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
	if c.CommentID.Valid {
		out.CommentID = c.CommentID.String
	}
	return out
}

type commentData struct {
	Comment commentDTO `json:"comment"`
}

func encodeCommentAddHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*commentSvc.AddResponse)
	if !ok || out == nil || out.Comment == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	env := shared.Envelope{Success: true, Data: commentData{Comment: toCommentDTO(out.Comment)}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeCommentReadHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*commentSvc.ReadResponse)
	if !ok || out == nil || out.Comment == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	env := shared.Envelope{Success: true, Data: commentData{Comment: toCommentDTO(out.Comment)}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

type commentsData struct {
	Comments []commentDTO `json:"comments"`
}

func encodeCommentListHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*commentSvc.ListResponse)
	if !ok || out == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	comments := make([]commentDTO, 0, len(out.Comments))
	for _, c := range out.Comments {
		dto := toCommentDTO(c)
		if v, ok := out.Interactions[c.ID]; ok && v != nil {
			vt := v.VoteType
			dto.UserVote = &vt
		}
		comments = append(comments, dto)
	}
	env := shared.Envelope{Success: true, Data: commentsData{Comments: comments}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

type commentVoteData struct {
	Upvote   int `json:"upvote"`
	Downvote int `json:"downvote"`
}

func encodeCommentVoteHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*commentSvc.VoteResponse)
	if !ok || out == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}
	env := shared.Envelope{Success: true, Data: commentVoteData{Upvote: out.Upvote, Downvote: out.Downvote}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodeCommentDeleteHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	env := shared.Envelope{Success: true}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}
