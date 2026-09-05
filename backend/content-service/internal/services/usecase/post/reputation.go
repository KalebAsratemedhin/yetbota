package post

import (
	"math"
	"sort"
	"time"

	"github.com/google/uuid"

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
		c.ShareWeight = 3
	}
	if c.BookmarkWeight <= 0 {
		c.BookmarkWeight = 3
	}
	if c.DislikeWeight <= 0 {
		c.DislikeWeight = 1
	}
	if c.CreateWeight <= 0 {
		c.CreateWeight = 0.1
	}
	if c.Alpha <= 0 {
		c.Alpha = 0.8
	}
	if c.Gravity <= 0 {
		c.Gravity = 1.5
	}
	// Larger time base so brand-new events are not worth ~w/2.8 each (was inflating to ~120).
	if c.TimeBaseHours <= 0 {
		c.TimeBaseHours = 4
	}
	if c.Scale <= 0 {
		c.Scale = 40
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

// eventContribution = w(kind) / (ageHours + T0)^G
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

// authorReputationFromEvents = round(scale × log10(1 + max(0, Σ contributions)^α))
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
	score = int64(math.Round(cfg.Scale * math.Log10(1+compressed)))
	if score < 0 {
		score = 0
	}
	return score, postsCreated, likesReceived
}

// rankTrendingCreators groups events by author, scores with reputation formula, returns top N.
func rankTrendingCreators(events []domain.InteractionEvent, now time.Time, cfg ReputationConfig, limit int) []TrendingCreator {
	cfg = cfg.withDefaults()
	byAuthor := map[uuid.UUID][]domain.InteractionEvent{}
	for _, ev := range events {
		byAuthor[ev.AuthorID] = append(byAuthor[ev.AuthorID], ev)
	}
	type scored struct {
		id    uuid.UUID
		raw   float64
		score int64
	}
	items := make([]scored, 0, len(byAuthor))
	for authorID, evs := range byAuthor {
		score, _, _ := authorReputationFromEvents(evs, now, cfg)
		if score <= 0 {
			continue
		}
		var raw float64
		for _, ev := range evs {
			raw += eventContribution(ev, now, cfg)
		}
		items = append(items, scored{id: authorID, raw: raw, score: score})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].raw == items[j].raw {
			return items[i].id.String() > items[j].id.String()
		}
		return items[i].raw > items[j].raw
	})
	if limit > len(items) {
		limit = len(items)
	}
	out := make([]TrendingCreator, 0, limit)
	for _, it := range items[:limit] {
		out = append(out, TrendingCreator{UserID: it.id, Score: it.score})
	}
	return out
}
