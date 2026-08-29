package follow

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ListParams is keyset pagination for follow edge lists.
type ListParams struct {
	Limit  int
	Cursor string // opaque; empty = first page
}

// EdgePage is one page of related user ids in list order (newest first).
type EdgePage struct {
	UserIDs    []uuid.UUID
	NextCursor string // empty when no further page
}

// Store persists follow edges in the shared graph DB (SurrealDB).
// Other services may read the same store with equivalent SurrealQL.
type Store interface {
	Follow(ctx context.Context, followerID, followeeID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error
	Followers(ctx context.Context, userID uuid.UUID, p ListParams) (EdgePage, error)
	Following(ctx context.Context, userID uuid.UUID, p ListParams) (EdgePage, error)
	CountFollowers(ctx context.Context, userID uuid.UUID) (int64, error)
	CountFollowing(ctx context.Context, userID uuid.UUID) (int64, error)
}

// Cursor encodes (created_at, peer user id) for stable keyset pagination.
type Cursor struct {
	CreatedAt time.Time
	UserID    uuid.UUID
}
