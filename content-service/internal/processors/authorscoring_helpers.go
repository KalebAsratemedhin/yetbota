package processors

import (
	"math"

	"github.com/beka-birhanu/yetbota/content-service/drivers/config"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
)

func applyEarnedBadges(user *dbmodels.User, badges []config.Badge) {
	existing := make(map[string]bool, len(user.Badges))
	for _, b := range user.Badges {
		existing[b] = true
	}
	for _, badge := range badges {
		if !existing[badge.Name] && user.Score >= badge.MinScore {
			user.Badges = append(user.Badges, badge.Name)
		}
	}
}

func computeNewScore(current int64, m *dbmodels.Post, postCount int64) int64 {
	implied := impliedRating(engagementScore(m))
	K := kFactor(postCount)
	delta := K * math.Tanh((implied-float64(current))/400)
	return current + int64(math.Round(delta))
}

func engagementScore(m *dbmodels.Post) float64 {
	raw := float64(m.Likes)*2 + float64(m.CommentCount)*3 - float64(m.Dislikes)
	if raw <= 0 {
		return 0
	}
	return math.Log2(raw + 1)
}

func impliedRating(engagement float64) float64 {
	if engagement <= 0 {
		return 1100
	}
	return 1500 + 400*math.Log2(engagement)
}

func kFactor(postCount int64) float64 {
	switch {
	case postCount <= 30:
		return 32
	case postCount <= 100:
		return 16
	default:
		return 8
	}
}
