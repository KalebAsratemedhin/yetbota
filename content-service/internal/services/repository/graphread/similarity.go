package graphread

import (
	"context"
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	surrealdbDriver "github.com/beka-birhanu/yetbota/content-service/drivers/surrealdb"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainPostSim "github.com/beka-birhanu/yetbota/content-service/internal/domain/postsimilarity"
)

type SimilarityRepository interface {
	SimilarPostsTree(ctx context.Context, postID string, maxDepth int) ([]domainPostSim.PostWithDepth, error)
}

type similarityRepo struct {
	graph *surrealdbDriver.Client
}

type SimilarityConfig struct {
	Graph *surrealdbDriver.Client `validate:"required"`
}

func (c *SimilarityConfig) Validate() error {
	return validator.Validate.Struct(c)
}

func NewSimilarityRepo(c *SimilarityConfig) (SimilarityRepository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &similarityRepo{graph: c.Graph}, nil
}

func (r *similarityRepo) SimilarPostsTree(ctx context.Context, postID string, maxDepth int) ([]domainPostSim.PostWithDepth, error) {
	if maxDepth < 1 {
		return nil, nil
	}
	seen := make(map[string]int)
	post := surrealdbDriver.PostID(postID)
	for depth := 1; depth <= maxDepth; depth++ {
		query := fmt.Sprintf(`SELECT VALUE meta::id(id) FROM $post.{%d}(<->similar_to<->post);`, depth)
		ids, err := surrealdbDriver.Query[[]string](ctx, r.graph, query, map[string]any{"post": post})
		if err != nil {
			return nil, similarityGraphErr("similar posts tree query failed", err)
		}
		for _, id := range ids {
			if id == "" || id == postID {
				continue
			}
			if _, ok := seen[id]; !ok {
				seen[id] = depth
			}
		}
	}
	posts := make([]domainPostSim.PostWithDepth, 0, len(seen))
	for id, depth := range seen {
		posts = append(posts, domainPostSim.PostWithDepth{PostID: id, Depth: depth})
	}
	return posts, nil
}

func similarityGraphErr(msg string, err error) error {
	serviceMsg := fmt.Sprintf("similarity graph repo: %s", msg)
	if err != nil {
		serviceMsg = fmt.Sprintf("similarity graph repo: %s: %v", msg, err)
	}
	return &toddlerr.Error{
		PublicStatusCode:  status.ServerError,
		ServiceStatusCode: status.ServerError,
		PublicMessage:     "something went wrong",
		ServiceMessage:    serviceMsg,
	}
}
