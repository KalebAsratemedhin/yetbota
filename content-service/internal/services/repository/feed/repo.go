package feed

import (
	"context"
	"fmt"
	"strconv"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/feed"
	"github.com/redis/go-redis/v9"
)

// RedisRepository stores feed items in a Redis sorted set.
//
// Key format:
//
//	feed:{userID}
//
// Member  -> postID
// Score   -> ranking score
//
// Items are ordered by score.
type RedisRepository struct {
	client *redis.Client
	prefix string
}

type Config struct {
	RDB    *redis.Client `validate:"required"`
	Prefix string
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewRedisRepository(c *Config) (feed.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &RedisRepository{
		client: c.RDB,
		prefix: c.Prefix,
	}, nil
}

func (r *RedisRepository) Count(ctx context.Context, userID string) (int64, error) {
	key := r.key(userID)

	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return 0, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("feed repo: count failed: %v", err),
		}
	}

	return n, nil
}

func (r *RedisRepository) AddBulk(ctx context.Context, userID string, items []*feed.FeedItem) error {
	if len(items) == 0 {
		return nil
	}

	members := make([]redis.Z, 0, len(items))

	for _, item := range items {
		members = append(members, redis.Z{
			Score:  item.Score,
			Member: item.PostID,
		})
	}

	err := r.client.ZAdd(ctx, r.key(userID), members...).Err()
	if err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("feed repo: add bulk failed: %v", err),
		}
	}
	return nil
}

// AddBulkGT adds items to the user's feed sorted set using ZADD GT — only updates an
// existing member if the new score is strictly greater than the current score.
func (r *RedisRepository) AddBulkGT(ctx context.Context, userID string, items []*feed.FeedItem) error {
	if len(items) == 0 {
		return nil
	}

	members := make([]redis.Z, 0, len(items))
	for _, item := range items {
		members = append(members, redis.Z{Score: item.Score, Member: item.PostID})
	}

	err := r.client.ZAddArgs(ctx, r.key(userID), redis.ZAddArgs{
		GT:      true,
		Members: members,
	}).Err()
	if err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("feed repo: add bulk GT failed: %v", err),
		}
	}
	return nil
}

// FanOutGT writes postID to each user's feed sorted set in a single pipeline using ZADD GT.
func (r *RedisRepository) FanOutGT(ctx context.Context, postID string, userScores map[string]float64) error {
	if len(userScores) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()
	for userID, score := range userScores {
		pipe.ZAddArgs(ctx, r.key(userID), redis.ZAddArgs{
			GT:      true,
			Members: []redis.Z{{Score: score, Member: postID}},
		})
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("feed repo: fan out GT failed: %v", err),
		}
	}
	return nil
}

// AddRecipients stores user→depth in a Redis hash for future score-update propagation.
func (r *RedisRepository) AddRecipients(ctx context.Context, postID string, userDepths map[string]float64) error {
	if len(userDepths) == 0 {
		return nil
	}
	args := make([]any, 0, len(userDepths)*2)
	for uid, depth := range userDepths {
		args = append(args, uid, depth)
	}
	if err := r.client.HSet(ctx, r.recipientsKey(postID), args...).Err(); err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("feed repo: add recipients failed: %v", err),
		}
	}
	return nil
}

// GetRecipients returns the depth at which each user received the post.
func (r *RedisRepository) GetRecipients(ctx context.Context, postID string) (map[string]float64, error) {
	raw, err := r.client.HGetAll(ctx, r.recipientsKey(postID)).Result()
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("feed repo: get recipients failed: %v", err),
		}
	}
	result := make(map[string]float64, len(raw))
	for uid, val := range raw {
		if depth, err := strconv.ParseFloat(val, 64); err == nil {
			result[uid] = depth
		}
	}
	return result, nil
}

// CheckRecipients reports whether each userID is already in the recipient hash.
func (r *RedisRepository) CheckRecipients(ctx context.Context, postID string, userIDs []string) (map[string]bool, error) {
	if len(userIDs) == 0 {
		return map[string]bool{}, nil
	}
	fields := make([]string, len(userIDs))
	for i, uid := range userIDs {
		fields[i] = uid
	}
	vals, err := r.client.HMGet(ctx, r.recipientsKey(postID), fields...).Result()
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("feed repo: check recipients failed: %v", err),
		}
	}
	result := make(map[string]bool, len(userIDs))
	for i, val := range vals {
		result[userIDs[i]] = val != nil
	}
	return result, nil
}

// CachePostScore stores the effective post score for delta comparison on the next interaction.
func (r *RedisRepository) CachePostScore(ctx context.Context, postID string, score float64) error {
	if err := r.client.Set(ctx, r.postScoreKey(postID), score, 0).Err(); err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("feed repo: cache post score failed: %v", err),
		}
	}
	return nil
}

// GetCachedPostScore retrieves the stored effective post score. Returns (0, false, nil) if not cached.
func (r *RedisRepository) GetCachedPostScore(ctx context.Context, postID string) (float64, bool, error) {
	val, err := r.client.Get(ctx, r.postScoreKey(postID)).Float64()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("feed repo: get cached post score failed: %v", err),
		}
	}
	return val, true, nil
}

func (r *RedisRepository) List(ctx context.Context, userID string, opts *feed.ListOptions) ([]*feed.FeedItem, error) {
	key := r.key(userID)

	// Descending order (highest score first)
	values, err := r.client.ZRevRangeByScoreWithScores(ctx, key, r.mapListOptions(opts)).Result()
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("feed repo: list failed: %v", err),
		}
	}

	items := make([]*feed.FeedItem, 0, len(values))

	for _, value := range values {
		postID, ok := value.Member.(string)
		if !ok {
			continue
		}

		items = append(items, &feed.FeedItem{
			PostID: postID,
			Score:  value.Score,
		})
	}

	return items, nil
}

// PublishCelebrityPost adds postID at score to the author's celebrity feed (ZADD GT) and trims
// the sorted set to at most maxSize entries by removing the lowest-scored items.
func (r *RedisRepository) PublishCelebrityPost(ctx context.Context, authorID, postID string, score float64, maxSize int64) error {
	key := r.celebrityKey(authorID)
	pipe := r.client.Pipeline()
	pipe.ZAddArgs(ctx, key, redis.ZAddArgs{
		GT:      true,
		Members: []redis.Z{{Score: score, Member: postID}},
	})
	pipe.ZRemRangeByRank(ctx, key, 0, -(maxSize + 2))
	if _, err := pipe.Exec(ctx); err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("feed repo: publish celebrity post failed: %v", err),
		}
	}
	return nil
}

// GetCelebrityPosts fetches posts from each author's celebrity feed sorted set, merges them
// by score descending, and returns at most opts.Limit items.
func (r *RedisRepository) GetCelebrityPosts(ctx context.Context, authorIDs []string, opts *feed.ListOptions) ([]*feed.FeedItem, error) {
	if len(authorIDs) == 0 {
		return nil, nil
	}
	rangeOpts := r.mapListOptions(opts)
	pipe := r.client.Pipeline()
	cmds := make([]*redis.ZSliceCmd, len(authorIDs))
	for i, id := range authorIDs {
		cmds[i] = pipe.ZRevRangeByScoreWithScores(ctx, r.celebrityKey(id), rangeOpts)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("feed repo: get celebrity posts failed: %v", err),
		}
	}

	limit := int(rangeOpts.Count)
	if limit <= 0 {
		limit = 50
	}
	all := make([]*feed.FeedItem, 0, limit*len(authorIDs))
	for _, cmd := range cmds {
		for _, z := range cmd.Val() {
			postID, ok := z.Member.(string)
			if !ok {
				continue
			}
			all = append(all, &feed.FeedItem{PostID: postID, Score: z.Score})
		}
	}

	// Sort merged results by score descending.
	for i := 1; i < len(all); i++ {
		for j := i; j > 0 && all[j].Score > all[j-1].Score; j-- {
			all[j], all[j-1] = all[j-1], all[j]
		}
	}
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}
