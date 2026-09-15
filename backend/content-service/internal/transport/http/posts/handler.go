package posts

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/yetbota/content-service/internal/domain/apperr"
	svcendpoint "github.com/yetbota/content-service/internal/services/endpoint"
	"github.com/yetbota/content-service/internal/transport/http/shared"
)

type Endpoints struct {
	SignMedia          endpoint.Endpoint
	Create             endpoint.Endpoint
	Update             endpoint.Endpoint
	Get                endpoint.Endpoint
	ListUser           endpoint.Endpoint
	SetReaction        endpoint.Endpoint
	AddComment         endpoint.Endpoint
	ListComments       endpoint.Endpoint
	ListReplies        endpoint.Endpoint
	SetCommentReaction endpoint.Endpoint
	AddBookmark        endpoint.Endpoint
	RemoveBookmark     endpoint.Endpoint
	ListBookmarks      endpoint.Endpoint
	Share              endpoint.Endpoint
	Search             endpoint.Endpoint
	ListHashtag        endpoint.Endpoint
	Feed               endpoint.Endpoint
	Related            endpoint.Endpoint
	TrendingPosts      endpoint.Endpoint
	TrendingCreators   endpoint.Endpoint
}

func Mount(mux *http.ServeMux, eps Endpoints, authn shared.TokenAuthenticator) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(shared.EncodeError),
	}
	auth := func(ep endpoint.Endpoint, dec kithttp.DecodeRequestFunc) http.Handler {
		return shared.RequireAuth(authn, kithttp.NewServer(ep, dec, shared.EncodeJSONResponse, opts...))
	}

	mux.Handle("POST /media/sign", auth(eps.SignMedia, decodeSignMedia))
	mux.Handle("POST /posts", auth(eps.Create, decodeCreate))
	mux.Handle("PATCH /posts/{id}", auth(eps.Update, decodeUpdate))
	mux.Handle("GET /posts/search", auth(eps.Search, decodeSearch))
	mux.Handle("GET /posts/trending", auth(eps.TrendingPosts, decodeTrending))
	mux.Handle("GET /posts/{id}", auth(eps.Get, decodePostID))
	mux.Handle("POST /posts/{id}/reaction", auth(eps.SetReaction, decodeReaction))
	mux.Handle("POST /posts/{id}/comments", auth(eps.AddComment, decodeComment))
	mux.Handle("GET /posts/{id}/comments", auth(eps.ListComments, decodePostList))
	mux.Handle("GET /comments/{id}/replies", auth(eps.ListReplies, decodeCommentList))
	mux.Handle("POST /comments/{id}/reaction", auth(eps.SetCommentReaction, decodeCommentReaction))
	mux.Handle("POST /posts/{id}/bookmark", auth(eps.AddBookmark, decodePostID))
	mux.Handle("DELETE /posts/{id}/bookmark", auth(eps.RemoveBookmark, decodePostID))
	mux.Handle("POST /posts/{id}/share", auth(eps.Share, decodePostID))
	mux.Handle("GET /users/{id}/posts", auth(eps.ListUser, decodeUserPosts))
	mux.Handle("GET /me/bookmarks", auth(eps.ListBookmarks, decodeMeList))
	mux.Handle("GET /hashtags/{tag}/posts", auth(eps.ListHashtag, decodeHashtag))
	mux.Handle("GET /feed", auth(eps.Feed, decodeFeed))
	mux.Handle("GET /posts/{id}/related", auth(eps.Related, decodeRelated))
	mux.Handle("GET /creators/trending", auth(eps.TrendingCreators, decodeTrendingCreators))
}

func decodeSignMedia(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var body struct {
		Kind string `json:"kind"`
	}
	if err := decodeJSON(r, &body); err != nil {
		return nil, err
	}
	return svcendpoint.SignMediaRequest{Principal: p, Kind: body.Kind}, nil
}

func decodeCreate(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	var req svcendpoint.CreatePostRequest
	if err := decodeJSON(r, &req); err != nil {
		return nil, err
	}
	req.Principal = p
	return req, nil
}

func decodeUpdate(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := svcendpoint.ParsePostID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	var req svcendpoint.UpdatePostRequest
	if err := decodeJSON(r, &req); err != nil {
		return nil, err
	}
	req.Principal = p
	req.PostID = id
	return req, nil
}

func decodePostID(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := svcendpoint.ParsePostID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.PostIDRequest{Principal: p, PostID: id}, nil
}

func decodeUserPosts(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := svcendpoint.ParseUserID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.UserPostsRequest{
		Principal: p, UserID: id, Limit: limit, Cursor: r.URL.Query().Get("cursor"),
	}, nil
}

func decodeReaction(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := svcendpoint.ParsePostID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	var body struct {
		Reaction *string `json:"reaction"`
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID", Message: "invalid JSON body"})
	}
	if len(bytes.TrimSpace(raw)) > 0 {
		if err := json.Unmarshal(raw, &body); err != nil {
			return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID", Message: "invalid JSON body"})
		}
	}
	return svcendpoint.ReactionRequest{Principal: p, PostID: id, Reaction: body.Reaction}, nil
}

func decodeComment(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := svcendpoint.ParsePostID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	var body struct {
		Body     string  `json:"body"`
		ParentID *string `json:"parent_id"`
	}
	if err := decodeJSON(r, &body); err != nil {
		return nil, err
	}
	return svcendpoint.CommentRequest{Principal: p, PostID: id, Body: body.Body, ParentID: body.ParentID}, nil
}

func decodeCommentList(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := svcendpoint.ParseCommentID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.CommentListRequest{
		Principal: p, CommentID: id, Limit: limit, Cursor: r.URL.Query().Get("cursor"),
	}, nil
}

func decodeCommentReaction(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := svcendpoint.ParseCommentID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	var body struct {
		Reaction *string `json:"reaction"`
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID", Message: "invalid JSON body"})
	}
	if len(bytes.TrimSpace(raw)) > 0 {
		if err := json.Unmarshal(raw, &body); err != nil {
			return nil, apperr.Validation(apperr.Field{Field: "body", Code: "INVALID", Message: "invalid JSON body"})
		}
	}
	return svcendpoint.CommentReactionRequest{Principal: p, CommentID: id, Reaction: body.Reaction}, nil
}

func decodePostList(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := svcendpoint.ParsePostID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.ListRequest{
		Principal: p, PostID: id, Limit: limit, Cursor: r.URL.Query().Get("cursor"),
	}, nil
}

func decodeMeList(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.ListRequest{
		Principal: p, Limit: limit, Cursor: r.URL.Query().Get("cursor"),
	}, nil
}

func decodeSearch(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.SearchRequest{
		Principal: p, Query: r.URL.Query().Get("q"), Limit: limit, Cursor: r.URL.Query().Get("cursor"),
	}, nil
}

func decodeHashtag(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.HashtagRequest{
		Principal: p, Tag: r.PathValue("tag"), Limit: limit, Cursor: r.URL.Query().Get("cursor"),
	}, nil
}

func decodeFeed(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	lat, err := svcendpoint.ParseOptionalCoord(r.URL.Query().Get("lat"), "lat")
	if err != nil {
		return nil, err
	}
	lng, err := svcendpoint.ParseOptionalCoord(r.URL.Query().Get("lng"), "lng")
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.FeedRequest{
		Principal: p, Lat: lat, Lng: lng, Limit: limit, Cursor: r.URL.Query().Get("cursor"),
	}, nil
}

func decodeRelated(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	id, err := svcendpoint.ParsePostID(r.PathValue("id"))
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.RelatedRequest{
		Principal: p,
		PostID:    id,
		Query:     r.URL.Query().Get("q"),
		Limit:     limit,
		Cursor:    r.URL.Query().Get("cursor"),
	}, nil
}

func decodeTrending(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.TrendingRequest{
		Principal: p, Limit: limit, Cursor: r.URL.Query().Get("cursor"),
	}, nil
}

func decodeTrendingCreators(ctx context.Context, r *http.Request) (interface{}, error) {
	p, err := shared.PrincipalFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	limit, err := svcendpoint.ParseListLimit(r.URL.Query().Get("limit"))
	if err != nil {
		return nil, err
	}
	return svcendpoint.TrendingRequest{Principal: p, Limit: limit}, nil
}

func decodeJSON(r *http.Request, dest interface{}) error {
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	if err := dec.Decode(dest); err != nil {
		return apperr.Validation(apperr.Field{Field: "body", Code: "INVALID", Message: "invalid JSON body"})
	}
	return nil
}
