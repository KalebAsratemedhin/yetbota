package processors

import (
	"context"
	"fmt"
	"math"

	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	domainFollower "github.com/beka-birhanu/yetbota/content-service/internal/domain/follower"
	"github.com/google/uuid"
)

func (a *feedUpdateActivity) computePostScore(likes, dislikes int, createdAtUnix int64) float64 {
	q := wilsonLowerBound(likes, dislikes)
	qComp := math.Log2(math.Max(q*a.qScale, 1))
	tComp := float64(createdAtUnix-a.epoch) / (a.halfLifeHours * 3600)
	return a.seedBonus + qComp + tComp
}

// wilsonLowerBound returns the lower bound of a 95% confidence interval on the upvote rate.
// Returns 0 when n=0.
func wilsonLowerBound(likes, dislikes int) float64 {
	n := float64(likes + dislikes)
	if n == 0 {
		return 0
	}
	const z = 1.96
	pHat := float64(likes) / n
	z2 := z * z
	return (pHat + z2/(2*n) - z*math.Sqrt((pHat*(1-pHat)+z2/(4*n))/n)) / (1 + z2/n)
}

// computeMaxDepth returns the deepest hop where postScore+distanceAttenuation(d) >= minFeedScore.
// Derived analytically: postScore − log₂(depth) >= minFeedScore → depth <= 2^(postScore−minFeedScore).
// Capped at 1000 to bound graph traversal cost.
func (a *feedUpdateActivity) computeMaxDepth(postScore float64) int {
	if postScore < a.minFeedScore {
		return 0
	}
	// 2^10 = 1024 > 1000, so any larger exponent exceeds the cap; short-circuit
	// to avoid math.Pow overflowing float64→int (which yields MinInt64).
	if postScore-a.minFeedScore >= 10 {
		return 1000
	}
	d := int(math.Pow(2, postScore-a.minFeedScore))
	if d > 1000 {
		return 1000
	}
	return d
}

// distanceAttenuation returns −log₂(depth), composing additively with postScore.
// depth=1 → 0 (no attenuation), depth=2 → −1, depth=4 → −2.
func distanceAttenuation(depth int) float64 {
	return -math.Log2(float64(depth))
}

// writeDepthBatches stores user→depth for users whose fanout score meets minFeedScore.
func (a *feedUpdateActivity) writeDepthBatches(ctx context.Context, users []domainFollower.UserWithDepth, postScore float64) ([]string, error) {
	depths := make(map[string]float64, len(users))
	for _, u := range users {
		if postScore+distanceAttenuation(u.Depth) >= a.minFeedScore {
			depths[u.UserID] = float64(u.Depth)
		}
	}
	return a.writeDepthBatchMap(ctx, depths)
}

// writeDepthBatchMap writes a pre-built user→depth map to Redis in batches.
func (a *feedUpdateActivity) writeDepthBatchMap(ctx context.Context, depths map[string]float64) ([]string, error) {
	return a.writeBatches(ctx, depths, "fanout:depth")
}

func (a *feedUpdateActivity) writeBatches(ctx context.Context, data map[string]float64, prefix string) ([]string, error) {
	if len(data) == 0 {
		return nil, nil
	}
	allUsers := make([]string, 0, len(data))
	for uid := range data {
		allUsers = append(allUsers, uid)
	}
	var keys []string
	for i := 0; i < len(allUsers); i += constants.FeedFanOutBatchSize {
		end := i + constants.FeedFanOutBatchSize
		if end > len(allUsers) {
			end = len(allUsers)
		}
		batch := make(map[string]float64, end-i)
		for _, uid := range allUsers[i:end] {
			batch[uid] = data[uid]
		}
		key := fmt.Sprintf("%s:%s", prefix, uuid.New().String())
		if err := a.batchStore.StoreBatch(ctx, key, batch, constants.FanOutBatchTTLSeconds); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}
