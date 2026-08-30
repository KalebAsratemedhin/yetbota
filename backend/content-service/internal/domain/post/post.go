package post

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type MediaMode string

const (
	ModeImages      MediaMode = "images"
	ModeImagesAudio MediaMode = "images_audio"
	ModeVideo       MediaMode = "video"
)

type MediaKind string

const (
	KindImage MediaKind = "image"
	KindVideo MediaKind = "video"
	KindAudio MediaKind = "audio"
)

type Reaction string

const (
	ReactionLike    Reaction = "like"
	ReactionDislike Reaction = "dislike"
)

type Post struct {
	ID            uuid.UUID
	AuthorID      uuid.UUID
	Caption       string
	Lat           float64
	Lng           float64
	MediaMode     MediaMode
	LikeCount     int64
	DislikeCount  int64
	CommentCount  int64
	ShareCount    int64
	BookmarkCount int64
	CreatedAt     time.Time
	Media         []Media
	Hashtags      []string
}

type Media struct {
	ID                 uuid.UUID
	PostID             uuid.UUID
	Kind               MediaKind
	CloudinaryPublicID string
	URL                string
	SortOrder          int
	DurationMs         int
	Width              int
	Height             int
	Mime               string
}

type Comment struct {
	ID           uuid.UUID
	PostID       uuid.UUID
	AuthorID     uuid.UUID
	ParentID     *uuid.UUID
	Body         string
	LikeCount    int64
	DislikeCount int64
	ReplyCount   int64
	CreatedAt    time.Time
}

type PageCursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

type ListParams struct {
	Limit  int
	Cursor *PageCursor
}

type Page struct {
	Posts      []Post
	NextCursor *PageCursor
}

type CommentPage struct {
	Comments   []Comment
	NextCursor *PageCursor
}

type CreateMediaInput struct {
	Kind               MediaKind
	CloudinaryPublicID string
	URL                string
	SortOrder          int
	DurationMs         int
	Width              int
	Height             int
	Mime               string
}

type CreateInput struct {
	AuthorID  uuid.UUID
	Caption   string
	Lat       float64
	Lng       float64
	MediaMode MediaMode
	Media     []CreateMediaInput
	Hashtags  []string
}

type Repository interface {
	Create(ctx context.Context, in CreateInput) (Post, error)
	ByID(ctx context.Context, id uuid.UUID) (Post, error)
	ListByAuthor(ctx context.Context, authorID uuid.UUID, p ListParams) (Page, error)
	SetReaction(ctx context.Context, postID, userID uuid.UUID, reaction *Reaction) (Post, error)
	AddComment(ctx context.Context, postID, authorID uuid.UUID, body string, parentID *uuid.UUID) (Comment, error)
	ListComments(ctx context.Context, postID uuid.UUID, p ListParams) (CommentPage, error)
	ListReplies(ctx context.Context, parentID uuid.UUID, p ListParams) (CommentPage, error)
	SetCommentReaction(ctx context.Context, commentID, userID uuid.UUID, reaction *Reaction) (Comment, error)
	CommentByID(ctx context.Context, id uuid.UUID) (Comment, error)
	AddBookmark(ctx context.Context, userID, postID uuid.UUID) (Post, error)
	RemoveBookmark(ctx context.Context, userID, postID uuid.UUID) (Post, error)
	ListBookmarks(ctx context.Context, userID uuid.UUID, p ListParams) (Page, error)
	Share(ctx context.Context, postID uuid.UUID) (Post, error)
	Search(ctx context.Context, query string, p ListParams) (Page, error)
	ListByHashtag(ctx context.Context, tag string, p ListParams) (Page, error)
}
