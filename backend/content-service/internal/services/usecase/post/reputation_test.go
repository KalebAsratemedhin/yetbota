package post

import (
	"testing"
	"time"

	"github.com/google/uuid"

	domain "github.com/yetbota/content-service/internal/domain/post"
)

func TestAuthorReputationFromEvents_OneBurstStaysModest(t *testing.T) {
	cfg := ReputationConfig{}.withDefaults()
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	author := uuid.New()
	actor := uuid.New()
	events := []domain.InteractionEvent{
		{AuthorID: author, ActorID: author, Kind: domain.InteractionCreate, CreatedAt: now},
		{AuthorID: author, ActorID: actor, Kind: domain.InteractionLike, CreatedAt: now},
		{AuthorID: author, ActorID: actor, Kind: domain.InteractionComment, CreatedAt: now},
		{AuthorID: author, ActorID: actor, Kind: domain.InteractionShare, CreatedAt: now},
	}
	score, posts, likes := authorReputationFromEvents(events, now, cfg)
	if posts != 1 || likes != 1 {
		t.Fatalf("posts=%d likes=%d", posts, likes)
	}
	// Must stay well below Contributor (100); previously ~120.
	if score < 1 || score > 25 {
		t.Fatalf("expected modest score ~5–20, got %d", score)
	}
}

func TestAuthorReputationFromEvents_FreshLikeBeatsStale(t *testing.T) {
	cfg := ReputationConfig{}.withDefaults()
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	author := uuid.New()
	actor := uuid.New()
	fresh := domain.InteractionEvent{
		AuthorID: author, ActorID: actor, Kind: domain.InteractionLike, CreatedAt: now.Add(-1 * time.Hour),
	}
	stale := domain.InteractionEvent{
		AuthorID: author, ActorID: actor, Kind: domain.InteractionLike, CreatedAt: now.Add(-72 * time.Hour),
	}
	sFresh, _, _ := authorReputationFromEvents([]domain.InteractionEvent{fresh}, now, cfg)
	sStale, _, _ := authorReputationFromEvents([]domain.InteractionEvent{stale}, now, cfg)
	if sFresh <= sStale {
		t.Fatalf("fresh %d should beat stale %d", sFresh, sStale)
	}
}

func TestAuthorReputationFromEvents_ShareWorthMoreThanLike(t *testing.T) {
	cfg := ReputationConfig{}.withDefaults()
	now := time.Now().UTC()
	author := uuid.New()
	sLike, _, _ := authorReputationFromEvents([]domain.InteractionEvent{
		{AuthorID: author, ActorID: uuid.New(), Kind: domain.InteractionLike, CreatedAt: now},
	}, now, cfg)
	sShare, _, _ := authorReputationFromEvents([]domain.InteractionEvent{
		{AuthorID: author, ActorID: uuid.New(), Kind: domain.InteractionShare, CreatedAt: now},
	}, now, cfg)
	if sShare <= sLike {
		t.Fatalf("share %d should beat like %d", sShare, sLike)
	}
}

func TestRankTrendingCreators_OrdersByEngagement(t *testing.T) {
	cfg := ReputationConfig{}.withDefaults()
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	hot := uuid.New()
	cold := uuid.New()
	actor := uuid.New()
	events := []domain.InteractionEvent{
		{AuthorID: cold, ActorID: actor, Kind: domain.InteractionLike, CreatedAt: now},
		{AuthorID: hot, ActorID: actor, Kind: domain.InteractionLike, CreatedAt: now},
		{AuthorID: hot, ActorID: actor, Kind: domain.InteractionComment, CreatedAt: now},
		{AuthorID: hot, ActorID: actor, Kind: domain.InteractionShare, CreatedAt: now},
	}
	out := rankTrendingCreators(events, now, cfg, 10)
	if len(out) < 2 {
		t.Fatalf("expected >=2 creators, got %d", len(out))
	}
	if out[0].UserID != hot {
		t.Fatalf("expected hot author first, got %s", out[0].UserID)
	}
	if out[0].Score <= out[1].Score {
		t.Fatalf("hot score %d should beat cold %d", out[0].Score, out[1].Score)
	}
}

func TestRankTrendingCreators_RespectsLimit(t *testing.T) {
	cfg := ReputationConfig{}.withDefaults()
	now := time.Now().UTC()
	events := []domain.InteractionEvent{}
	for i := 0; i < 5; i++ {
		author := uuid.New()
		events = append(events, domain.InteractionEvent{
			AuthorID: author, ActorID: uuid.New(), Kind: domain.InteractionShare, CreatedAt: now,
		})
	}
	out := rankTrendingCreators(events, now, cfg, 2)
	if len(out) != 2 {
		t.Fatalf("limit 2 → got %d", len(out))
	}
}

