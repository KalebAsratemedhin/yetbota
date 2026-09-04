package post

import (
	"math"
	"time"

	domain "github.com/yetbota/content-service/internal/domain/post"
)

// ReputationConfig scores authors from post_interaction_events (weighted + gravity decay).
type ReputationConfig struct {
	LikeWeight     float64
	CommentWeight  float64
	ShareWeight    float64
	BookmarkWeight float64
	DislikeWeight  float64
	CreateWeight   float64
	Alpha          float64
	Gravity        float64
	TimeBaseHours  float64
	Scale          float64
	LookbackDays   int
	CandidateLimit int
}

func (c ReputationConfig) withDefaults() ReputationConfig {
	if c.LikeWeight <= 0 {
		c.LikeWeight = 1
	}
	if c.CommentWeight <= 0 {
		c.CommentWeight = 3
	}
	if c.ShareWeight <= 0 {
		c.ShareWeight = 4
	}
	if c.BookmarkWeight <= 0 {
		c.BookmarkWeight = 3
	}
	if c.DislikeWeight <= 0 {
		c.DislikeWeight = 1
	}
	if c.CreateWeight <= 0 {
		c.CreateWeight = 0.25
	}
	if c.Alpha <= 0 {
		c.Alpha = 0.8
	}
	if c.Gravity <= 0 {
		c.Gravity = 1.5
	}
	if c.TimeBaseHours <= 0 {
		c.TimeBaseHours = 2
	}
	if c.Scale <= 0 {
		c.Scale = 100
	}
	if c.LookbackDays <= 0 {
		c.LookbackDays = 90
	}
	if c.CandidateLimit <= 0 {
		c.CandidateLimit = 5000
	}
	return c
}

func interactionWeight(kind string, cfg ReputationConfig) float64 {
	switch kind {
	case domain.InteractionCreate:
		return cfg.CreateWeight
	case domain.InteractionLike:
		return cfg.LikeWeight
	case domain.InteractionUnlike:
		return -cfg.LikeWeight
	case domain.InteractionDislike:
		return -cfg.DislikeWeight
	case domain.InteractionUndislike:
		return cfg.DislikeWeight
	case domain.InteractionComment:
		return cfg.CommentWeight
	case domain.InteractionShare:
		return cfg.ShareWeight
	case domain.InteractionBookmark:
		return cfg.BookmarkWeight
	case domain.InteractionUnbookmark:
		return -cfg.BookmarkWeight
	default:
		return 0
	}
}

// eventContribution = w(kind) / (ageHours + T0)^G  (freshness of the interaction).
func eventContribution(ev domain.InteractionEvent, now time.Time, cfg ReputationConfig) float64 {
	w := interactionWeight(ev.Kind, cfg)
	if w == 0 {
		return 0
	}
	ageHours := now.Sub(ev.CreatedAt.UTC()).Hours()
	if ageHours < 0 {
		ageHours = 0
	}
	return w / math.Pow(ageHours+cfg.TimeBaseHours, cfg.Gravity)
}

// authorReputationFromEvents = round(scale × log(1 + max(0, Σ contributions)^α))
func authorReputationFromEvents(events []domain.InteractionEvent, now time.Time, cfg ReputationConfig) (score int64, postsCreated, likesReceived int64) {
	cfg = cfg.withDefaults()
	var raw float64
	var likes int64
	for _, ev := range events {
		switch ev.Kind {
		case domain.InteractionCreate:
			postsCreated++
		case domain.InteractionLike:
			likes++
		case domain.InteractionUnlike:
			likes--
		}
		raw += eventContribution(ev, now, cfg)
	}
	if likes < 0 {
		likes = 0
	}
	likesReceived = likes
	if raw < 0 {
		raw = 0
	}
	compressed := math.Pow(raw, cfg.Alpha)
	score = int64(math.Round(cfg.Scale * math.Log(1+compressed)))
	if score < 0 {
		score = 0
	}
	return score, postsCreated, likesReceived
}
