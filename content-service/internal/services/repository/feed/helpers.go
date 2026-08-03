package feed

import (
	"fmt"
	"math"
	"strconv"

	feedDomain "github.com/beka-birhanu/yetbota/content-service/internal/domain/feed"
	"github.com/redis/go-redis/v9"
)

func (r *RedisRepository) key(userID string) string {
	return fmt.Sprintf("%s:%s", r.prefix, userID)
}

func (r *RedisRepository) celebrityKey(authorID string) string {
	return fmt.Sprintf("%s:CELEBRITY:%s", r.prefix, authorID)
}

func (r *RedisRepository) recipientsKey(postID string) string {
	return fmt.Sprintf("%s:RECIPIENTS:%s", r.prefix, postID)
}

func (r *RedisRepository) postScoreKey(postID string) string {
	return fmt.Sprintf("%s:SCORE:%s", r.prefix, postID)
}

func (r *RedisRepository) mapListOptions(opts *feedDomain.ListOptions) *redis.ZRangeBy {
	minScore := "-inf"
	maxScore := "+inf"
	limit := int64(50)

	if opts != nil {
		if opts.MinScore > 0 {
			minScore = formatFloat(opts.MinScore)
		}

		if opts.MaxScore > 0 {
			maxScore = formatFloat(opts.MaxScore)
		}

		if opts.Limit > 0 {
			limit = int64(opts.Limit)
		}
	}

	return &redis.ZRangeBy{
		Max:    maxScore,
		Min:    minScore,
		Offset: 0,
		Count:  limit,
	}
}

func formatFloat(v float64) string {
	if math.IsInf(v, 1) {
		return "+inf"
	}

	if math.IsInf(v, -1) {
		return "-inf"
	}

	return strconv.FormatFloat(v, 'f', -1, 64)
}
