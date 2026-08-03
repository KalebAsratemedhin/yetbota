package follower

import "context"

type UserWithDepth struct {
	UserID string
	Depth  int
}

type Repository interface {
	FollowerTree(ctx context.Context, authorID string, maxDepth int) ([]UserWithDepth, error)
	FollowersOf(ctx context.Context, userIDs []string) (map[string][]string, error)
	CountFollowers(ctx context.Context, authorID string) (int64, error)
	ListCelebrityFollows(ctx context.Context, userID string, threshold int64) ([]string, error)
	// IsFollowing returns a map of userID → true if the user is following the user.
	IsFollowing(ctx context.Context, followerID string, userIDs []string) (map[string]bool, error)
}
