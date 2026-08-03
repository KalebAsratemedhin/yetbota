package processors

// PostFanOutData holds the data needed to start a fan-out for a post.
type PostFanOutData struct {
	Score    float64
	AuthorID string
}

// SimilarPostInfo carries the data needed to fan out a similar post.
type SimilarPostInfo struct {
	PostID string
	Depth  int
	Score  float64
}

// FollowerFanOutResult holds Redis depth-batch keys from GetFollowerTree.
// IsCelebrity is true when the author's follower count exceeds the celebrity threshold;
// in that case BatchKeys is empty and a celebrity feed entry is published instead.
type FollowerFanOutResult struct {
	BatchKeys   []string
	IsCelebrity bool
}

// SimFanOutResult holds Redis depth-batch keys and similar post metadata from GetSimilarPostsTree.
type SimFanOutResult struct {
	BatchUserKeys []string
	SimPosts      []SimilarPostInfo
}
