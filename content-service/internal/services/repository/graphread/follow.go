package graphread

import (
	"context"
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	surrealdbDriver "github.com/beka-birhanu/yetbota/content-service/drivers/surrealdb"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainFollower "github.com/beka-birhanu/yetbota/content-service/internal/domain/follower"
)

type FollowRepository interface {
	FollowerTree(ctx context.Context, authorID string, maxDepth int) ([]domainFollower.UserWithDepth, error)
	ListFollowedIDs(ctx context.Context, userID string) ([]string, error)
	IsFollowing(ctx context.Context, followerID string, userIDs []string) (map[string]bool, error)
	FollowersOf(ctx context.Context, userIDs []string) (map[string][]string, error)
}

type followRepo struct {
	graph *surrealdbDriver.Client
}

type FollowConfig struct {
	Graph *surrealdbDriver.Client `validate:"required"`
}

func (c *FollowConfig) Validate() error {
	return validator.Validate.Struct(c)
}

func NewFollowRepo(c *FollowConfig) (FollowRepository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &followRepo{graph: c.Graph}, nil
}

func (r *followRepo) FollowerTree(ctx context.Context, authorID string, maxDepth int) ([]domainFollower.UserWithDepth, error) {
	if maxDepth < 1 {
		return nil, nil
	}
	seen := make(map[string]int)
	author := surrealdbDriver.UserID(authorID)
	for depth := 1; depth <= maxDepth; depth++ {
		query := fmt.Sprintf(`SELECT VALUE meta::id(id) FROM $author.{%d}(<-follows<-user);`, depth)
		ids, err := surrealdbDriver.Query[[]string](ctx, r.graph, query, map[string]any{"author": author})
		if err != nil {
			return nil, followGraphErr("follower tree query failed", err)
		}
		for _, id := range ids {
			if id == "" || id == authorID {
				continue
			}
			if _, ok := seen[id]; !ok {
				seen[id] = depth
			}
		}
	}
	users := make([]domainFollower.UserWithDepth, 0, len(seen))
	for id, depth := range seen {
		users = append(users, domainFollower.UserWithDepth{UserID: id, Depth: depth})
	}
	return users, nil
}

func (r *followRepo) ListFollowedIDs(ctx context.Context, userID string) ([]string, error) {
	ids, err := surrealdbDriver.Query[[]string](ctx, r.graph, `
		SELECT VALUE meta::id(out) FROM follows WHERE in = $user;
	`, map[string]any{"user": surrealdbDriver.UserID(userID)})
	if err != nil {
		return nil, followGraphErr("list followed ids query failed", err)
	}
	return ids, nil
}

func (r *followRepo) IsFollowing(ctx context.Context, followerID string, userIDs []string) (map[string]bool, error) {
	if len(userIDs) == 0 || followerID == "" {
		return nil, nil
	}

	targets := make([]any, 0, len(userIDs))
	for _, id := range userIDs {
		targets = append(targets, surrealdbDriver.UserID(id))
	}

	ids, err := surrealdbDriver.Query[[]string](ctx, r.graph, `
		SELECT VALUE meta::id(out) FROM follows
		WHERE in = $follower AND out IN $targets;
	`, map[string]any{
		"follower": surrealdbDriver.UserID(followerID),
		"targets":  targets,
	})
	if err != nil {
		return nil, followGraphErr("is following query failed", err)
	}

	following := make(map[string]bool, len(userIDs))
	for _, id := range ids {
		following[id] = true
	}
	return following, nil
}

func (r *followRepo) FollowersOf(ctx context.Context, userIDs []string) (map[string][]string, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	targets := make([]any, 0, len(userIDs))
	for _, id := range userIDs {
		targets = append(targets, surrealdbDriver.UserID(id))
	}

	rows, err := surrealdbDriver.Query[[]map[string]any](ctx, r.graph, `
		SELECT meta::id(out) AS userID, meta::id(in) AS followerID FROM follows
		WHERE out IN $targets;
	`, map[string]any{"targets": targets})
	if err != nil {
		return nil, followGraphErr("followers of users query failed", err)
	}

	res := make(map[string][]string)
	for _, row := range rows {
		userID, _ := row["userID"].(string)
		followerID, _ := row["followerID"].(string)
		if userID == "" || followerID == "" {
			continue
		}
		res[userID] = append(res[userID], followerID)
	}
	return res, nil
}

func followGraphErr(msg string, err error) error {
	serviceMsg := fmt.Sprintf("follow graph repo: %s", msg)
	if err != nil {
		serviceMsg = fmt.Sprintf("follow graph repo: %s: %v", msg, err)
	}
	return &toddlerr.Error{
		PublicStatusCode:  status.ServerError,
		ServiceStatusCode: status.ServerError,
		PublicMessage:     "something went wrong",
		ServiceMessage:    serviceMsg,
	}
}
