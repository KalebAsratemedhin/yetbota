package follow

import "context"

type Pagination struct {
	Limit int
	Page  int
}

// Repository manages the user follow network in the graph database.
type Repository interface {
	Follow(ctx context.Context, followerID, followeeID string) error
	Unfollow(ctx context.Context, followerID, followeeID string) error
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)
	Followers(ctx context.Context, userID string, pagination *Pagination) ([]string, error)
	Following(ctx context.Context, userID string, pagination *Pagination) ([]string, error)
	CountFollowers(ctx context.Context, userID string) (int64, error)
	CountFollowing(ctx context.Context, userID string) (int64, error)
}
