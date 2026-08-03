package feed

import "context"

type FeedItem struct {
	PostID string  `json:"postID" validate:"required,uuid4"`
	Score  float64 `json:"score" validate:"required,min=0"`
}

type Repository interface {
	Count(ctx context.Context, userID string) (int64, error)
	List(ctx context.Context, userID string, opts *ListOptions) ([]*FeedItem, error)
	AddBulk(ctx context.Context, userID string, items []*FeedItem) error
	// AddBulkGT adds items to the feed only when the new score is greater than the existing score.
	AddBulkGT(ctx context.Context, userID string, items []*FeedItem) error
	// FanOutGT writes postID to each user's feed at the given score (ZADD GT).
	// A single pipelined round-trip covers all users.
	FanOutGT(ctx context.Context, postID string, userScores map[string]float64) error
	// AddRecipients records which users received a post and at what depth (for future score propagation).
	AddRecipients(ctx context.Context, postID string, userDepths map[string]float64) error
	// GetRecipients returns the depth at which each user received the post.
	GetRecipients(ctx context.Context, postID string) (map[string]float64, error)
	// CheckRecipients reports whether each userID is already in the recipient set.
	CheckRecipients(ctx context.Context, postID string, userIDs []string) (map[string]bool, error)
	// CachePostScore stores the effective post score for delta comparison on the next interaction.
	CachePostScore(ctx context.Context, postID string, score float64) error
	// GetCachedPostScore retrieves the stored effective post score. Returns (0, false, nil) if not cached.
	GetCachedPostScore(ctx context.Context, postID string) (float64, bool, error)
	// PublishCelebrityPost adds postID at score to the author's celebrity feed sorted set (ZADD GT).
	// Trims the set to maxSize after adding.
	PublishCelebrityPost(ctx context.Context, authorID, postID string, score float64, maxSize int64) error
	// GetCelebrityPosts fetches and merges posts from the celebrity feeds of authorIDs, filtered and
	// ordered by opts. Returns at most opts.Limit items sorted by score descending.
	GetCelebrityPosts(ctx context.Context, authorIDs []string, opts *ListOptions) ([]*FeedItem, error)
}

type SeenRepository interface {
	AddBulk(ctx context.Context, userID string, postIDs []string) error
}

type ListOptions struct {
	MinScore float64
	MaxScore float64
	Limit    int
}
