package processors

import (
	"context"
	"fmt"

	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainFeed "github.com/beka-birhanu/yetbota/content-service/internal/domain/feed"
	domainFollower "github.com/beka-birhanu/yetbota/content-service/internal/domain/follower"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	domainPostSim "github.com/beka-birhanu/yetbota/content-service/internal/domain/postsimilarity"
	domainPostvote "github.com/beka-birhanu/yetbota/content-service/internal/domain/postvote"
	domainStorage "github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
)

type feedUpdateActivity struct {
	followerRepo         domainFollower.Repository
	postSimRepo          domainPostSim.Repository
	feedRepo             domainFeed.Repository
	postRepo             domainPost.Repository
	postvoteRepo         domainPostvote.Repository
	batchStore           domainStorage.Set
	seenCache            domainStorage.Set
	seedBonus            float64
	qScale               float64
	epoch                int64
	halfLifeHours        float64
	minFeedScore         float64
	celebrityThreshold   int64
	maxCelebrityFeedSize int64
}

type feedUpdateActConfig struct {
	FollowerRepo         domainFollower.Repository `validate:"required"`
	PostSimRepo          domainPostSim.Repository  `validate:"required"`
	FeedRepo             domainFeed.Repository     `validate:"required"`
	PostRepo             domainPost.Repository     `validate:"required"`
	PostvoteRepo         domainPostvote.Repository `validate:"required"`
	BatchStore           domainStorage.Set         `validate:"required"`
	SeenCache            domainStorage.Set         `validate:"required"`
	SeedBonus            float64                   `validate:"required"`
	QScale               float64                   `validate:"required"`
	Epoch                int64                     `validate:"required"`
	HalfLifeHours        float64                   `validate:"required"`
	MinFeedScore         float64                   `validate:"required"`
	CelebrityThreshold   int64                     `validate:"required"`
	MaxCelebrityFeedSize int64                     `validate:"required"`
}

func (c *feedUpdateActConfig) validate() error {
	return validator.Validate.Struct(c)
}

func newFeedUpdateAct(cfg *feedUpdateActConfig) (*feedUpdateActivity, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &feedUpdateActivity{
		followerRepo:         cfg.FollowerRepo,
		postSimRepo:          cfg.PostSimRepo,
		feedRepo:             cfg.FeedRepo,
		postRepo:             cfg.PostRepo,
		postvoteRepo:         cfg.PostvoteRepo,
		batchStore:           cfg.BatchStore,
		seenCache:            cfg.SeenCache,
		seedBonus:            cfg.SeedBonus,
		qScale:               cfg.QScale,
		epoch:                cfg.Epoch,
		halfLifeHours:        cfg.HalfLifeHours,
		minFeedScore:         cfg.MinFeedScore,
		celebrityThreshold:   cfg.CelebrityThreshold,
		maxCelebrityFeedSize: cfg.MaxCelebrityFeedSize,
	}, nil
}

// FetchPostFanOutData computes postScore = seedBonus + Q(q) + F(t) and returns author ID.
// Q(q) = log₂(max(q·QScale, 1)), F(t) = (createdAt − epoch) / (halfLifeHours · 3600).
func (a *feedUpdateActivity) FetchPostFanOutData(ctx context.Context, postID string) (*PostFanOutData, error) {
	post, err := a.postRepo.Read(ctx, postID)
	if err != nil {
		return nil, err
	}
	return &PostFanOutData{
		Score:    a.computePostScore(post.Likes, post.Dislikes, post.CreatedAt.Unix()),
		AuthorID: post.UserID,
	}, nil
}

// GetFollowerTree checks the author's follower count. If it exceeds the celebrity threshold,
// returns IsCelebrity=true with no batch keys — the workflow publishes to the celebrity feed
// instead. Otherwise fetches the transitive follower tree and writes depth-batches to Redis.
func (a *feedUpdateActivity) GetFollowerTree(ctx context.Context, authorID string, postScore float64) (*FollowerFanOutResult, error) {
	count, err := a.followerRepo.CountFollowers(ctx, authorID)
	if err != nil {
		return nil, err
	}
	if count > a.celebrityThreshold {
		return &FollowerFanOutResult{IsCelebrity: true}, nil
	}

	maxDepth := a.computeMaxDepth(postScore)
	if maxDepth == 0 {
		return &FollowerFanOutResult{}, nil
	}
	followers, err := a.followerRepo.FollowerTree(ctx, authorID, maxDepth)
	if err != nil {
		return nil, err
	}
	keys, err := a.writeDepthBatches(ctx, followers, postScore)
	if err != nil {
		return nil, err
	}
	return &FollowerFanOutResult{BatchKeys: keys}, nil
}

// PublishToCelebrityFeed adds the post to the author's celebrity feed sorted set.
func (a *feedUpdateActivity) PublishToCelebrityFeed(ctx context.Context, authorID, postID string, score float64) error {
	return a.feedRepo.PublishCelebrityPost(ctx, authorID, postID, score, a.maxCelebrityFeedSize)
}

// seenKey builds the Redis key checked by seenCache for a (userID, postID) pair.
func seenKey(userID, postID string) string {
	return fmt.Sprintf("%s:%s", userID, postID)
}

// checkSeen returns a map[userID]bool indicating which users have seen postID.
func (a *feedUpdateActivity) checkSeen(ctx context.Context, postID string, userIDs []string) (map[string]bool, error) {
	keys := make([]string, len(userIDs))
	for i, uid := range userIDs {
		keys[i] = seenKey(uid, postID)
	}
	raw, err := a.seenCache.Exists(ctx, keys)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(userIDs))
	for _, uid := range userIDs {
		result[uid] = raw[seenKey(uid, postID)]
	}
	return result, nil
}

// CachePostScore stores the effective post score for future delta comparison.
func (a *feedUpdateActivity) CachePostScore(ctx context.Context, postID string, score float64) error {
	return a.feedRepo.CachePostScore(ctx, postID, score)
}

// GetSimilarPostsTree fetches the similarity graph for postID, collects direct interactors of
// similar posts, computes per-interactor fan-out scores, and returns Redis batch keys plus
// the similar post metadata needed for step 5 fan-out.
func (a *feedUpdateActivity) GetSimilarPostsTree(ctx context.Context, postID string, postScore float64) (*SimFanOutResult, error) {
	maxDepth := a.computeMaxDepth(postScore)
	if maxDepth == 0 {
		return &SimFanOutResult{}, nil
	}
	simNodes, err := a.postSimRepo.SimilarPostsTree(ctx, postID, maxDepth)
	if err != nil {
		return nil, err
	}
	if len(simNodes) == 0 {
		return &SimFanOutResult{}, nil
	}

	simPostIDs := make([]string, len(simNodes))
	simDepthByPost := make(map[string]int, len(simNodes))
	for i, sp := range simNodes {
		simPostIDs[i] = sp.PostID
		simDepthByPost[sp.PostID] = sp.Depth
	}

	// Fetch similar posts to compute their individual scores.
	posts, err := a.postRepo.List(ctx, &domainPost.ListOptions{
		IDs:      simPostIDs,
		Page:     1,
		PageSize: len(simPostIDs),
	})
	if err != nil {
		return nil, err
	}
	simPostInfos := make([]SimilarPostInfo, 0, len(posts))
	for _, p := range posts {
		simPostInfos = append(simPostInfos, SimilarPostInfo{
			PostID: p.ID,
			Depth:  simDepthByPost[p.ID],
			Score:  a.computePostScore(p.Likes, p.Dislikes, p.CreatedAt.Unix()),
		})
	}

	// Direct interactors of similar posts — keep best (shallowest) sim depth per user.
	interactorsMap, err := a.postvoteRepo.ListVotersByPostIDs(ctx, simPostIDs)
	if err != nil {
		return nil, err
	}
	interactorBestDepth := make(map[string]int)
	for simPostID, users := range interactorsMap {
		depth := simDepthByPost[simPostID]
		for _, userID := range users {
			if existing, ok := interactorBestDepth[userID]; !ok || depth < existing {
				interactorBestDepth[userID] = depth
			}
		}
	}

	interactorDepths := make(map[string]float64, len(interactorBestDepth))
	for userID, depth := range interactorBestDepth {
		if postScore+distanceAttenuation(depth) >= a.minFeedScore {
			interactorDepths[userID] = float64(depth)
		}
	}

	userKeys, err := a.writeDepthBatchMap(ctx, interactorDepths)
	if err != nil {
		return nil, err
	}
	return &SimFanOutResult{BatchUserKeys: userKeys, SimPosts: simPostInfos}, nil
}

// FanOutFeedBatch fans out postID to all users in batchKey.
// New recipients are added via ZADD GT and saved. Existing recipients are updated via ZADD GT
// only if the user has not yet seen the post.
func (a *feedUpdateActivity) FanOutFeedBatch(ctx context.Context, postID, batchKey string, postScore float64) error {
	userDepths, err := a.batchStore.ReadBatch(ctx, batchKey)
	if err != nil {
		return err
	}
	if len(userDepths) == 0 {
		return nil
	}

	userIDs := make([]string, 0, len(userDepths))
	for uid := range userDepths {
		userIDs = append(userIDs, uid)
	}

	isRecipient, err := a.feedRepo.CheckRecipients(ctx, postID, userIDs)
	if err != nil {
		return err
	}

	existingUIDs := make([]string, 0, len(userIDs))
	for _, uid := range userIDs {
		if isRecipient[uid] {
			existingUIDs = append(existingUIDs, uid)
		}
	}

	seenMap := map[string]bool{}
	if len(existingUIDs) > 0 {
		seenMap, err = a.checkSeen(ctx, postID, existingUIDs)
		if err != nil {
			return err
		}
	}

	userScores := make(map[string]float64, len(userDepths))
	newDepths := make(map[string]float64)
	for uid, depth := range userDepths {
		score := postScore + distanceAttenuation(int(depth))
		if score < a.minFeedScore {
			continue
		}
		if isRecipient[uid] {
			if !seenMap[uid] {
				userScores[uid] = score
			}
		} else {
			userScores[uid] = score
			newDepths[uid] = depth
		}
	}

	if len(userScores) > 0 {
		if err := a.feedRepo.FanOutGT(ctx, postID, userScores); err != nil {
			return err
		}
	}
	if len(newDepths) > 0 {
		return a.feedRepo.AddRecipients(ctx, postID, newDepths)
	}
	return nil
}

// FanOutSimilarPostsToFollowerBatch fans out all simPosts to follower-tree users in depthKey,
// scoring each as simPost.Score + attn(sim_depth) + attn(d_follow). Skips users who have seen
// the sim post. ZADD GT keeps higher score.
func (a *feedUpdateActivity) FanOutSimilarPostsToFollowerBatch(ctx context.Context, simPosts []SimilarPostInfo, depthKey string) error {
	userDepths, err := a.batchStore.ReadBatch(ctx, depthKey)
	if err != nil {
		return err
	}
	if len(userDepths) == 0 {
		return nil
	}

	userIDs := make([]string, 0, len(userDepths))
	for uid := range userDepths {
		userIDs = append(userIDs, uid)
	}

	for _, sp := range simPosts {
		seenMap, err := a.checkSeen(ctx, sp.PostID, userIDs)
		if err != nil {
			return err
		}
		baseScore := sp.Score + distanceAttenuation(sp.Depth)
		userScores := make(map[string]float64, len(userDepths))
		for uid, depth := range userDepths {
			if seenMap[uid] {
				continue
			}
			score := baseScore + distanceAttenuation(int(depth))
			if score >= a.minFeedScore {
				userScores[uid] = score
			}
		}
		if len(userScores) > 0 {
			if err := a.feedRepo.FanOutGT(ctx, sp.PostID, userScores); err != nil {
				return err
			}
		}
	}
	return nil
}

// FanOutSimilarPostsToUserBatch fans out all simPosts to sim-interactor users in depthKey,
// scoring each as simPost.Score + attn(sim_depth). Skips users who have seen the sim post.
// ZADD GT keeps higher score.
func (a *feedUpdateActivity) FanOutSimilarPostsToUserBatch(ctx context.Context, simPosts []SimilarPostInfo, depthKey string) error {
	batch, err := a.batchStore.ReadBatch(ctx, depthKey)
	if err != nil {
		return err
	}
	if len(batch) == 0 {
		return nil
	}

	userIDs := make([]string, 0, len(batch))
	for uid := range batch {
		userIDs = append(userIDs, uid)
	}

	for _, sp := range simPosts {
		score := sp.Score + distanceAttenuation(sp.Depth)
		if score < a.minFeedScore {
			continue
		}
		seenMap, err := a.checkSeen(ctx, sp.PostID, userIDs)
		if err != nil {
			return err
		}
		userScores := make(map[string]float64, len(batch))
		for _, uid := range userIDs {
			if !seenMap[uid] {
				userScores[uid] = score
			}
		}
		if len(userScores) > 0 {
			if err := a.feedRepo.FanOutGT(ctx, sp.PostID, userScores); err != nil {
				return err
			}
		}
	}
	return nil
}

// FanOutSimilarPostsToAuthor fans out all simPosts to the post author,
// scoring each as simPost.Score + attn(sim_depth). Skips posts the author has already seen.
// ZADD GT keeps higher score.
func (a *feedUpdateActivity) FanOutSimilarPostsToAuthor(ctx context.Context, authorID string, simPosts []SimilarPostInfo) error {
	postIDs := make([]string, len(simPosts))
	for i, sp := range simPosts {
		postIDs[i] = sp.PostID
	}

	// Bulk seen check: one key per sim post for the author.
	seenKeys := make([]string, len(simPosts))
	for i, sp := range simPosts {
		seenKeys[i] = seenKey(authorID, sp.PostID)
	}
	seenRaw, err := a.seenCache.Exists(ctx, seenKeys)
	if err != nil {
		return err
	}

	for _, sp := range simPosts {
		if seenRaw[seenKey(authorID, sp.PostID)] {
			continue
		}
		score := sp.Score + distanceAttenuation(sp.Depth)
		if score < a.minFeedScore {
			continue
		}
		if err := a.feedRepo.FanOutGT(ctx, sp.PostID, map[string]float64{authorID: score}); err != nil {
			return err
		}
	}
	return nil
}
