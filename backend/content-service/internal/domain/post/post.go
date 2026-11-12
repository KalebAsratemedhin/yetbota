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

// Interaction event kinds logged for reputation (content-owned ledger).
const (
	InteractionCreate     = "create"
	InteractionLike       = "like"
	InteractionUnlike     = "unlike"
	InteractionDislike    = "dislike"
	InteractionUndislike  = "undislike"
	InteractionComment    = "comment"
	InteractionShare      = "share"
	InteractionBookmark   = "bookmark"
	InteractionUnbookmark = "unbookmark"
)

type InteractionEvent struct {
	ID        uuid.UUID
	PostID    uuid.UUID
	AuthorID  uuid.UUID
	ActorID   uuid.UUID
	Kind      string
	CreatedAt time.Time
}

type Post struct {
	ID            uuid.UUID
	AuthorID      uuid.UUID
	Title         string // place name
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
	// Bytes is optional client-reported size; when > 0, validated against MaxUploadBytes.
	Bytes int64
}

type CreateInput struct {
	AuthorID  uuid.UUID
	Title     string
	Caption   string
	Lat       float64
	Lng       float64
	MediaMode MediaMode
	Media     []CreateMediaInput
	Hashtags  []string
}

// UpdateInput is a full write of editable fields (caller merges with current post).
type UpdateInput struct {
	PostID    uuid.UUID
	Title     string
	Caption   string
	Lat       float64
	Lng       float64
	MediaMode MediaMode
	Media     []CreateMediaInput // if non-nil, replaces all media
	Hashtags  []string           // always replaces hashtag links
	ReplaceMedia bool
}

type Repository interface {
	Create(ctx context.Context, in CreateInput) (Post, error)
	Update(ctx context.Context, in UpdateInput) (Post, error)
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
	Share(ctx context.Context, postID, actorID uuid.UUID) (Post, error)
	Search(ctx context.Context, query string, p ListParams) (Page, error)
	ListByHashtag(ctx context.Context, tag string, p ListParams) (Page, error)
	ListByAuthorIDs(ctx context.Context, authorIDs []uuid.UUID, since time.Time, limit int) ([]Post, error)
	ListNearby(ctx context.Context, lat, lng, radiusKm float64, since time.Time, limit int) ([]Post, error)
	ListRecent(ctx context.Context, since time.Time, limit int) ([]Post, error)
	// ListTopByEngagement pages by absolute engagement (likes+comments+shares+bookmarks) DESC.
	ListTopByEngagement(ctx context.Context, cursorEngagement *int64, cursorID *uuid.UUID, limit int) (posts []Post, hasMore bool, err error)
	ListAuthorInteractions(ctx context.Context, authorID uuid.UUID, since time.Time, limit int) ([]InteractionEvent, error)
	ListInteractionsSince(ctx context.Context, since time.Time, limit int) ([]InteractionEvent, error)
	SearchExcluding(ctx context.Context, query string, excludeID uuid.UUID, p ListParams) (Page, error)
	ListSimilar(ctx context.Context, seedText string, excludeID uuid.UUID, excludeQuery string, cursor *SimilarCursor, limit int) (ScoredPage, error)
	SeedText(ctx context.Context, postID uuid.UUID) (string, error)
	ByIDs(ctx context.Context, ids []uuid.UUID) ([]Post, error)
	IDsMatchingQuery(ctx context.Context, ids []uuid.UUID, query string) ([]uuid.UUID, error)
}

type SimilarCursor struct {
	Rank      float64
	CreatedAt time.Time
	ID        uuid.UUID
}

type ScoredPost struct {
	Post Post
	Rank float64
}

type ScoredPage struct {
	Posts   []ScoredPost
	HasMore bool
}
