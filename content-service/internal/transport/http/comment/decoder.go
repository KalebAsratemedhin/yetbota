package comment

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	commentSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/comment"
)

func decodeCommentAddHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		PostID    string `json:"post_id"`
		Comment   string `json:"comment"`
		IsAnswer  bool   `json:"is_answer"`
		CommentID string `json:"comment_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &commentSvc.AddRequest{
		PostID:    in.PostID,
		Comment:   in.Comment,
		IsAnswer:  in.IsAnswer,
		CommentID: in.CommentID,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeCommentReadHTTP(ctx context.Context, r *http.Request) (any, error) {
	req := &commentSvc.ReadRequest{ID: r.PathValue("id")}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeCommentListHTTP(ctx context.Context, r *http.Request) (any, error) {
	q := r.URL.Query()
	req := &commentSvc.ListRequest{
		PostID:    strings.TrimSpace(q.Get("post_id")),
		CommentID: strings.TrimSpace(q.Get("comment_id")),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeCommentDeleteHTTP(ctx context.Context, r *http.Request) (any, error) {
	req := &commentSvc.DeleteRequest{ID: r.PathValue("id")}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeCommentVoteHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		VoteType string `json:"vote_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &commentSvc.VoteRequest{
		CommentID: r.PathValue("id"),
		VoteType:  strings.ToLower(strings.TrimSpace(in.VoteType)),
	}
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
