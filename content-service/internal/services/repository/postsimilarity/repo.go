package postsimilarity

import (
	"context"

	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainPostSim "github.com/beka-birhanu/yetbota/content-service/internal/domain/postsimilarity"
	repoGraphRead "github.com/beka-birhanu/yetbota/content-service/internal/services/repository/graphread"
)

type repository struct {
	graph repoGraphRead.SimilarityRepository
}

type Config struct {
	Graph repoGraphRead.SimilarityRepository `validate:"required"`
}

func (c *Config) Validate() error {
	return validator.Validate.Struct(c)
}

func NewRepo(c *Config) (domainPostSim.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{graph: c.Graph}, nil
}

func (r *repository) SimilarPostsTree(ctx context.Context, postID string, maxDepth int) ([]domainPostSim.PostWithDepth, error) {
	return r.graph.SimilarPostsTree(ctx, postID, maxDepth)
}
