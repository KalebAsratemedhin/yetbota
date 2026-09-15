package endpoint

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"

	"github.com/yetbota/content-service/internal/domain/apperr"
	"github.com/yetbota/content-service/internal/domain/auth"
	domain "github.com/yetbota/content-service/internal/domain/post"
	postuc "github.com/yetbota/content-service/internal/services/usecase/post"
)

type PostService interface {
	SignMedia(ctx context.Context, in postuc.SignMediaInput) (postuc.SignMediaResult, error)
	Create(ctx context.Context, in postuc.CreatePostInput) (domain.Post, error)
	Update(ctx context.Context, in postuc.UpdatePostInput) (domain.Post, error)
	Get(ctx context.Context, id uuid.UUID) (domain.Post, error)
	ListByAuthor(ctx context.Context, authorID uuid.UUID, in postuc.ListInput) (postuc.PageResult, error)
	SetReaction(ctx context.Context, postID, userID uuid.UUID, reaction *domain.Reaction) (domain.Post, error)
	AddComment(ctx context.Context, postID, authorID uuid.UUID, body string, parentID *uuid.UUID) (domain.Comment, error)
	ListComments(ctx context.Context, postID uuid.UUID, in postuc.ListInput) (postuc.CommentPageResult, error)
	ListReplies(ctx context.Context, parentID uuid.UUID, in postuc.ListInput) (postuc.CommentPageResult, error)
	SetCommentReaction(ctx context.Context, commentID, userID uuid.UUID, reaction *domain.Reaction) (domain.Comment, error)
	AddBookmark(ctx context.Context, userID, postID uuid.UUID) (domain.Post, error)
	RemoveBookmark(ctx context.Context, userID, postID uuid.UUID) (domain.Post, error)
	ListBookmarks(ctx context.Context, userID uuid.UUID, in postuc.ListInput) (postuc.PageResult, error)
	Share(ctx context.Context, postID, actorID uuid.UUID) (postuc.ShareResult, error)
	Search(ctx context.Context, query string, in postuc.ListInput) (postuc.PageResult, error)
	ListByHashtag(ctx context.Context, tag string, in postuc.ListInput) (postuc.PageResult, error)
	Feed(ctx context.Context, in postuc.FeedInput) (postuc.PageResult, error)
	Related(ctx context.Context, in postuc.RelatedInput) (postuc.RelatedResult, error)
	TrendingPosts(ctx context.Context, in postuc.TrendingListInput) (postuc.PageResult, error)
	TrendingCreators(ctx context.Context, in postuc.TrendingListInput) (postuc.TrendingCreatorsResult, error)
}

type SignMediaRequest struct {
	Principal auth.Principal
	Kind      string `json:"kind"`
}

type CreatePostRequest struct {
	Principal auth.Principal
	Title     string            `json:"title"`
	Caption   string            `json:"caption"`
	Lat       *float64          `json:"lat"`
	Lng       *float64          `json:"lng"`
	MediaMode string            `json:"media_mode"`
	Media     []CreateMediaJSON `json:"media"`
	Hashtags  []string          `json:"hashtags"`
}

type UpdatePostRequest struct {
	Principal auth.Principal
	PostID    uuid.UUID
	Title     *string            `json:"title"`
	Caption   *string            `json:"caption"`
	Lat       *float64           `json:"lat"`
	Lng       *float64           `json:"lng"`
	MediaMode *string            `json:"media_mode"`
	Media     *[]CreateMediaJSON `json:"media"`
	Hashtags  *[]string          `json:"hashtags"`
}

type CreateMediaJSON struct {
	Kind               string `json:"kind"`
	CloudinaryPublicID string `json:"cloudinary_public_id"`
	URL                string `json:"url"`
	SortOrder          int    `json:"sort_order"`
	DurationMs         int    `json:"duration_ms"`
	Width              int    `json:"width"`
	Height             int    `json:"height"`
	Mime               string `json:"mime"`
}

type PostIDRequest struct {
	Principal auth.Principal
	PostID    uuid.UUID
}

type UserPostsRequest struct {
	Principal auth.Principal
	UserID    uuid.UUID
	Limit     int
	Cursor    string
}

type ReactionRequest struct {
	Principal auth.Principal
	PostID    uuid.UUID
	Reaction  *string `json:"reaction"`
}

type CommentRequest struct {
	Principal auth.Principal
	PostID    uuid.UUID
	Body      string  `json:"body"`
	ParentID  *string `json:"parent_id"`
}

type CommentReactionRequest struct {
	Principal auth.Principal
	CommentID uuid.UUID
	Reaction  *string `json:"reaction"`
}

type CommentListRequest struct {
	Principal auth.Principal
	CommentID uuid.UUID
	Limit     int
	Cursor    string
}

type ListRequest struct {
	Principal auth.Principal
	PostID    uuid.UUID
	Limit     int
	Cursor    string
}

type SearchRequest struct {
	Principal auth.Principal
	Query     string
	Limit     int
	Cursor    string
}

type HashtagRequest struct {
	Principal auth.Principal
	Tag       string
	Limit     int
	Cursor    string
}

type FeedRequest struct {
	Principal auth.Principal
	Lat       *float64
	Lng       *float64
	Limit     int
	Cursor    string
}

type RelatedRequest struct {
	Principal auth.Principal
	PostID    uuid.UUID
	Query     string
	Limit     int
	Cursor    string
}

type TrendingRequest struct {
	Principal auth.Principal
	Limit     int
	Cursor    string
}

type RelatedPageJSON struct {
	Posts      []PostJSON `json:"posts"`
	NextCursor string     `json:"next_cursor,omitempty"`
	Source     string     `json:"source"`
}

type TrendingCreatorJSON struct {
	UserID string `json:"user_id"`
	Score  int64  `json:"score"`
}

type TrendingCreatorsPageJSON struct {
	Creators []TrendingCreatorJSON `json:"creators"`
}

type MediaJSON struct {
	ID                 string `json:"id"`
	Kind               string `json:"kind"`
	CloudinaryPublicID string `json:"cloudinary_public_id"`
	URL                string `json:"url"`
	SortOrder          int    `json:"sort_order"`
	DurationMs         int    `json:"duration_ms"`
	Width              int    `json:"width"`
	Height             int    `json:"height"`
	Mime               string `json:"mime"`
}

type PostJSON struct {
	ID            string      `json:"id"`
	AuthorID      string      `json:"author_id"`
	Title         string      `json:"title"`
	Caption       string      `json:"caption"`
	Lat           float64     `json:"lat"`
	Lng           float64     `json:"lng"`
	MediaMode     string      `json:"media_mode"`
	LikeCount     int64       `json:"like_count"`
	DislikeCount  int64       `json:"dislike_count"`
	CommentCount  int64       `json:"comment_count"`
	ShareCount    int64       `json:"share_count"`
	BookmarkCount int64       `json:"bookmark_count"`
	CreatedAt     string      `json:"created_at"`
	Media         []MediaJSON `json:"media"`
	Hashtags      []string    `json:"hashtags"`
}

type PostsPageJSON struct {
	Posts      []PostJSON `json:"posts"`
	NextCursor string     `json:"next_cursor,omitempty"`
}

type CommentJSON struct {
	ID           string  `json:"id"`
	PostID       string  `json:"post_id"`
	AuthorID     string  `json:"author_id"`
	ParentID     *string `json:"parent_id,omitempty"`
	Body         string  `json:"body"`
	LikeCount    int64   `json:"like_count"`
	DislikeCount int64   `json:"dislike_count"`
	ReplyCount   int64   `json:"reply_count"`
	CreatedAt    string  `json:"created_at"`
}

type CommentsPageJSON struct {
	Comments   []CommentJSON `json:"comments"`
	NextCursor string        `json:"next_cursor,omitempty"`
}

type ShareJSON struct {
	Post     PostJSON `json:"post"`
	ShareURL string   `json:"share_url"`
}

func MakeSignMediaEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(SignMediaRequest)
		return svc.SignMedia(ctx, postuc.SignMediaInput{UserID: req.Principal.UserID, Kind: req.Kind})
	}
}

func MakeCreatePostEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CreatePostRequest)
		authorID, err := ParseUserID(req.Principal.UserID)
		if err != nil {
			return nil, err
		}
		if req.Lat == nil {
			return nil, apperr.Validation(apperr.Required("lat"))
		}
		if req.Lng == nil {
			return nil, apperr.Validation(apperr.Required("lng"))
		}
		media := make([]domain.CreateMediaInput, 0, len(req.Media))
		for _, m := range req.Media {
			media = append(media, domain.CreateMediaInput{
				Kind: domain.MediaKind(m.Kind), CloudinaryPublicID: m.CloudinaryPublicID, URL: m.URL,
				SortOrder: m.SortOrder, DurationMs: m.DurationMs, Width: m.Width, Height: m.Height, Mime: m.Mime,
			})
		}
		p, err := svc.Create(ctx, postuc.CreatePostInput{
			AuthorID: authorID, Title: req.Title, Caption: req.Caption, Lat: *req.Lat, Lng: *req.Lng,
			MediaMode: domain.MediaMode(req.MediaMode), Media: media, Hashtags: req.Hashtags,
		})
		if err != nil {
			return nil, err
		}
		return toPostJSON(p), nil
	}
}

func MakeUpdatePostEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdatePostRequest)
		actorID, err := ParseUserID(req.Principal.UserID)
		if err != nil {
			return nil, err
		}
		in := postuc.UpdatePostInput{
			PostID: req.PostID, ActorID: actorID,
			Title: req.Title, Caption: req.Caption, Lat: req.Lat, Lng: req.Lng, Hashtags: req.Hashtags,
		}
		if req.MediaMode != nil {
			m := domain.MediaMode(*req.MediaMode)
			in.MediaMode = &m
		}
		if req.Media != nil {
			media := make([]domain.CreateMediaInput, 0, len(*req.Media))
			for _, m := range *req.Media {
				media = append(media, domain.CreateMediaInput{
					Kind: domain.MediaKind(m.Kind), CloudinaryPublicID: m.CloudinaryPublicID, URL: m.URL,
					SortOrder: m.SortOrder, DurationMs: m.DurationMs, Width: m.Width, Height: m.Height, Mime: m.Mime,
				})
			}
			in.Media = &media
		}
		p, err := svc.Update(ctx, in)
		if err != nil {
			return nil, err
		}
		return toPostJSON(p), nil
	}
}

func MakeGetPostEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(PostIDRequest)
		p, err := svc.Get(ctx, req.PostID)
		if err != nil {
			return nil, err
		}
		return toPostJSON(p), nil
	}
}

func MakeListUserPostsEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UserPostsRequest)
		page, err := svc.ListByAuthor(ctx, req.UserID, postuc.ListInput{Limit: req.Limit, Cursor: req.Cursor})
		if err != nil {
			return nil, err
		}
		return toPostsPageJSON(page), nil
	}
}

func MakeSetReactionEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ReactionRequest)
		userID, err := ParseUserID(req.Principal.UserID)
		if err != nil {
			return nil, err
		}
		var reaction *domain.Reaction
		if req.Reaction != nil {
			r := domain.Reaction(strings.ToLower(strings.TrimSpace(*req.Reaction)))
			if r != "" {
				reaction = &r
			}
		}
		p, err := svc.SetReaction(ctx, req.PostID, userID, reaction)
		if err != nil {
			return nil, err
		}
		return toPostJSON(p), nil
	}
}

func MakeAddCommentEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CommentRequest)
		authorID, err := ParseUserID(req.Principal.UserID)
		if err != nil {
			return nil, err
		}
		var parentID *uuid.UUID
		if req.ParentID != nil && strings.TrimSpace(*req.ParentID) != "" {
			id, err := ParsePostID(strings.TrimSpace(*req.ParentID))
			if err != nil {
				return nil, apperr.Validation(apperr.Field{
					Field: "parent_id", Code: "INVALID", Message: "parent_id must be a UUID",
				})
			}
			parentID = &id
		}
		c, err := svc.AddComment(ctx, req.PostID, authorID, req.Body, parentID)
		if err != nil {
			return nil, err
		}
		return toCommentJSON(c), nil
	}
}

func MakeListCommentsEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ListRequest)
		page, err := svc.ListComments(ctx, req.PostID, postuc.ListInput{Limit: req.Limit, Cursor: req.Cursor})
		if err != nil {
			return nil, err
		}
		return toCommentsPageJSON(page), nil
	}
}

func MakeListRepliesEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CommentListRequest)
		page, err := svc.ListReplies(ctx, req.CommentID, postuc.ListInput{Limit: req.Limit, Cursor: req.Cursor})
		if err != nil {
			return nil, err
		}
		return toCommentsPageJSON(page), nil
	}
}

func MakeSetCommentReactionEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CommentReactionRequest)
		userID, err := ParseUserID(req.Principal.UserID)
		if err != nil {
			return nil, err
		}
		var reaction *domain.Reaction
		if req.Reaction != nil {
			r := domain.Reaction(strings.ToLower(strings.TrimSpace(*req.Reaction)))
			if r != "" {
				reaction = &r
			}
		}
		c, err := svc.SetCommentReaction(ctx, req.CommentID, userID, reaction)
		if err != nil {
			return nil, err
		}
		return toCommentJSON(c), nil
	}
}

func MakeAddBookmarkEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(PostIDRequest)
		userID, err := ParseUserID(req.Principal.UserID)
		if err != nil {
			return nil, err
		}
		p, err := svc.AddBookmark(ctx, userID, req.PostID)
		if err != nil {
			return nil, err
		}
		return toPostJSON(p), nil
	}
}

func MakeRemoveBookmarkEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(PostIDRequest)
		userID, err := ParseUserID(req.Principal.UserID)
		if err != nil {
			return nil, err
		}
		p, err := svc.RemoveBookmark(ctx, userID, req.PostID)
		if err != nil {
			return nil, err
		}
		return toPostJSON(p), nil
	}
}

func MakeListBookmarksEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ListRequest)
		userID, err := ParseUserID(req.Principal.UserID)
		if err != nil {
			return nil, err
		}
		page, err := svc.ListBookmarks(ctx, userID, postuc.ListInput{Limit: req.Limit, Cursor: req.Cursor})
		if err != nil {
			return nil, err
		}
		return toPostsPageJSON(page), nil
	}
}

func MakeShareEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(PostIDRequest)
		actorID, err := ParseUserID(req.Principal.UserID)
		if err != nil {
			return nil, err
		}
		res, err := svc.Share(ctx, req.PostID, actorID)
		if err != nil {
			return nil, err
		}
		return ShareJSON{Post: toPostJSON(res.Post), ShareURL: res.ShareURL}, nil
	}
}

func MakeSearchEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(SearchRequest)
		page, err := svc.Search(ctx, req.Query, postuc.ListInput{Limit: req.Limit, Cursor: req.Cursor})
		if err != nil {
			return nil, err
		}
		return toPostsPageJSON(page), nil
	}
}

func MakeListByHashtagEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(HashtagRequest)
		page, err := svc.ListByHashtag(ctx, req.Tag, postuc.ListInput{Limit: req.Limit, Cursor: req.Cursor})
		if err != nil {
			return nil, err
		}
		return toPostsPageJSON(page), nil
	}
}

func MakeFeedEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(FeedRequest)
		viewerID, err := ParseUserID(req.Principal.UserID)
		if err != nil {
			return nil, err
		}
		page, err := svc.Feed(ctx, postuc.FeedInput{
			ViewerID: viewerID,
			Lat:      req.Lat,
			Lng:      req.Lng,
			Limit:    req.Limit,
			Cursor:   req.Cursor,
		})
		if err != nil {
			return nil, err
		}
		return toPostsPageJSON(page), nil
	}
}

func MakeRelatedEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(RelatedRequest)
		page, err := svc.Related(ctx, postuc.RelatedInput{
			SeedID: req.PostID,
			Query:  req.Query,
			Limit:  req.Limit,
			Cursor: req.Cursor,
		})
		if err != nil {
			return nil, err
		}
		posts := make([]PostJSON, 0, len(page.Posts))
		for _, p := range page.Posts {
			posts = append(posts, toPostJSON(p))
		}
		return RelatedPageJSON{Posts: posts, NextCursor: page.NextCursor, Source: page.Source}, nil
	}
}

func MakeTrendingPostsEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(TrendingRequest)
		page, err := svc.TrendingPosts(ctx, postuc.TrendingListInput{
			Limit: req.Limit, Cursor: req.Cursor,
		})
		if err != nil {
			return nil, err
		}
		return toPostsPageJSON(page), nil
	}
}

func MakeTrendingCreatorsEndpoint(svc PostService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(TrendingRequest)
		page, err := svc.TrendingCreators(ctx, postuc.TrendingListInput{Limit: req.Limit})
		if err != nil {
			return nil, err
		}
		creators := make([]TrendingCreatorJSON, 0, len(page.Creators))
		for _, c := range page.Creators {
			creators = append(creators, TrendingCreatorJSON{
				UserID: c.UserID.String(),
				Score:  c.Score,
			})
		}
		return TrendingCreatorsPageJSON{Creators: creators}, nil
	}
}

func ParseOptionalCoord(raw, field string) (*float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	n, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, apperr.Validation(apperr.Field{
			Field: field, Code: "INVALID", Message: field + " must be a number",
		})
	}
	return &n, nil
}

func ParseUserID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return uuid.Nil, apperr.Validation(apperr.Field{
			Field: "id", Code: "INVALID", Message: "id must be a UUID",
		})
	}
	return id, nil
}

func ParsePostID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return uuid.Nil, apperr.Validation(apperr.Field{
			Field: "id", Code: "INVALID", Message: "id must be a UUID",
		})
	}
	return id, nil
}

func ParseCommentID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return uuid.Nil, apperr.Validation(apperr.Field{
			Field: "id", Code: "INVALID", Message: "id must be a UUID",
		})
	}
	return id, nil
}

func ParseListLimit(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, apperr.Validation(apperr.Field{
			Field: "limit", Code: "INVALID", Message: "limit must be a non-negative integer",
		})
	}
	return n, nil
}

func toPostJSON(p domain.Post) PostJSON {
	media := make([]MediaJSON, 0, len(p.Media))
	for _, m := range p.Media {
		media = append(media, MediaJSON{
			ID: m.ID.String(), Kind: string(m.Kind), CloudinaryPublicID: m.CloudinaryPublicID,
			URL: m.URL, SortOrder: m.SortOrder, DurationMs: m.DurationMs,
			Width: m.Width, Height: m.Height, Mime: m.Mime,
		})
	}
	tags := p.Hashtags
	if tags == nil {
		tags = []string{}
	}
	return PostJSON{
		ID: p.ID.String(), AuthorID: p.AuthorID.String(), Title: p.Title, Caption: p.Caption,
		Lat: p.Lat, Lng: p.Lng, MediaMode: string(p.MediaMode),
		LikeCount: p.LikeCount, DislikeCount: p.DislikeCount, CommentCount: p.CommentCount,
		ShareCount: p.ShareCount, BookmarkCount: p.BookmarkCount,
		CreatedAt: p.CreatedAt.UTC().Format(time.RFC3339Nano), Media: media, Hashtags: tags,
	}
}

func toPostsPageJSON(page postuc.PageResult) PostsPageJSON {
	posts := make([]PostJSON, 0, len(page.Posts))
	for _, p := range page.Posts {
		posts = append(posts, toPostJSON(p))
	}
	return PostsPageJSON{Posts: posts, NextCursor: page.NextCursor}
}

func toCommentJSON(c domain.Comment) CommentJSON {
	out := CommentJSON{
		ID: c.ID.String(), PostID: c.PostID.String(), AuthorID: c.AuthorID.String(),
		Body: c.Body, LikeCount: c.LikeCount, DislikeCount: c.DislikeCount, ReplyCount: c.ReplyCount,
		CreatedAt: c.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
	if c.ParentID != nil {
		s := c.ParentID.String()
		out.ParentID = &s
	}
	return out
}

func toCommentsPageJSON(page postuc.CommentPageResult) CommentsPageJSON {
	comments := make([]CommentJSON, 0, len(page.Comments))
	for _, c := range page.Comments {
		comments = append(comments, toCommentJSON(c))
	}
	return CommentsPageJSON{Comments: comments, NextCursor: page.NextCursor}
}
