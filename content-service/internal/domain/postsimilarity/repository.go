package postsimilarity

import "context"

type PostWithDepth struct {
	PostID string
	Depth  int
}

type Repository interface {
	SimilarPostsTree(ctx context.Context, postID string, maxDepth int) ([]PostWithDepth, error)
}
