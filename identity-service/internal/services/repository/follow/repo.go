package follow

import (
	"context"
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	surrealdbDriver "github.com/beka-birhanu/yetbota/identity-service/drivers/surrealdb"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	domainFollow "github.com/beka-birhanu/yetbota/identity-service/internal/domain/follow"
)

type repo struct {
	graph *surrealdbDriver.Client
}

type Config struct {
	Graph *surrealdbDriver.Client `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewRepo(c *Config) (domainFollow.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repo{graph: c.Graph}, nil
}

func (r *repo) Follow(ctx context.Context, followerID, followeeID string) error {
	follower := surrealdbDriver.UserID(followerID)
	followee := surrealdbDriver.UserID(followeeID)
	edge := surrealdbDriver.FollowEdgeID(followerID, followeeID)
	_, err := surrealdbDriver.Query[any](ctx, r.graph, `
		UPSERT $follower;
		UPSERT $followee;
		RELATE OR UPDATE $follower->$edge->$followee;
	`, map[string]any{
		"follower": follower,
		"followee": followee,
		"edge":     edge,
	})
	if err != nil {
		return fromGraphError(err)
	}
	return nil
}

func (r *repo) Unfollow(ctx context.Context, followerID, followeeID string) error {
	edge := surrealdbDriver.FollowEdgeID(followerID, followeeID)
	deleted, err := surrealdbDriver.Query[[]map[string]any](ctx, r.graph, `
		DELETE $edge RETURN BEFORE;
	`, map[string]any{"edge": edge})
	if err != nil {
		return fromGraphError(err)
	}
	if len(deleted) == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "Follow relationship not found",
			ServiceMessage:    fmt.Sprintf("no FOLLOWS edge from %s to %s", followerID, followeeID),
		}
	}
	return nil
}

func (r *repo) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	edge := surrealdbDriver.FollowEdgeID(followerID, followeeID)
	rows, err := surrealdbDriver.Query[[]map[string]any](ctx, r.graph, `
		SELECT id FROM $edge;
	`, map[string]any{"edge": edge})
	if err != nil {
		return false, fromGraphError(err)
	}
	return len(rows) > 0, nil
}

func (r *repo) Followers(ctx context.Context, userID string, pagination *domainFollow.Pagination) ([]string, error) {
	skip, limit := paginationParams(pagination)
	ids, err := surrealdbDriver.Query[[]string](ctx, r.graph, `
		SELECT VALUE meta::id(in) FROM follows
		WHERE out = $user
		ORDER BY in
		START $skip
		LIMIT $limit;
	`, map[string]any{
		"user":  surrealdbDriver.UserID(userID),
		"skip":  skip,
		"limit": limit,
	})
	if err != nil {
		return nil, fromGraphError(err)
	}
	return ids, nil
}

func (r *repo) Following(ctx context.Context, userID string, pagination *domainFollow.Pagination) ([]string, error) {
	skip, limit := paginationParams(pagination)
	ids, err := surrealdbDriver.Query[[]string](ctx, r.graph, `
		SELECT VALUE meta::id(out) FROM follows
		WHERE in = $user
		ORDER BY out
		START $skip
		LIMIT $limit;
	`, map[string]any{
		"user":  surrealdbDriver.UserID(userID),
		"skip":  skip,
		"limit": limit,
	})
	if err != nil {
		return nil, fromGraphError(err)
	}
	return ids, nil
}

func (r *repo) CountFollowers(ctx context.Context, userID string) (int64, error) {
	rows, err := surrealdbDriver.Query[[]map[string]any](ctx, r.graph, `
		SELECT count() AS cnt FROM follows WHERE out = $user GROUP ALL;
	`, map[string]any{"user": surrealdbDriver.UserID(userID)})
	if err != nil {
		return 0, fromGraphError(err)
	}
	return countFromRows(rows), nil
}

func (r *repo) CountFollowing(ctx context.Context, userID string) (int64, error) {
	rows, err := surrealdbDriver.Query[[]map[string]any](ctx, r.graph, `
		SELECT count() AS cnt FROM follows WHERE in = $user GROUP ALL;
	`, map[string]any{"user": surrealdbDriver.UserID(userID)})
	if err != nil {
		return 0, fromGraphError(err)
	}
	return countFromRows(rows), nil
}
