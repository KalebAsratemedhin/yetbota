package post

import (
	"testing"
	"time"

	"github.com/google/uuid"

	domain "github.com/yetbota/content-service/internal/domain/post"
)

func TestAuthorReputationFromEvents_FreshLikeBeatsStale(t *testing.T) {
	cfg := ReputationConfig{}.withDefaults()
	now := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
	author := uuid.New()
	actor := uuid.New()
	postID := uuid.New()
	fresh := domain.InteractionEvent{
		PostID: postID, AuthorID: author, ActorID: actor,
		Kind: domain.InteractionLike, CreatedAt: now.Add(-1 * time.Hour),
	}
	stale := domain.InteractionEvent{
		PostID: postID, AuthorID: author, ActorID: actor,
		Kind: domain.InteractionLike, CreatedAt: now.Add(-72 * time.Hour),
	}
	sFresh, _, likes := authorReputationFromEvents([]domain.InteractionEvent{fresh}, now, cfg)
	sStale, _, _ := authorReputationFromEvents([]domain.InteractionEvent{stale}, now, cfg)
	if likes != 1 {
		t.Fatalf("likes=%d", likes)
	}
	if sFresh <= sStale {
		t.Fatalf("fresh %d should beat stale %d", sFresh, sStale)
	}
}

func TestAuthorReputationFromEvents_ShareWorthMoreThanLike(t *testing.T) {
	cfg := ReputationConfig{}.withDefaults()
	now := time.Now().UTC()
	author := uuid.New()
	evLike := domain.InteractionEvent{
		AuthorID: author, ActorID: uuid.New(), Kind: domain.InteractionLike, CreatedAt: now,
	}
	evShare := domain.InteractionEvent{
		AuthorID: author, ActorID: uuid.New(), Kind: domain.InteractionShare, CreatedAt: now,
	}
	sLike, _, _ := authorReputationFromEvents([]domain.InteractionEvent{evLike}, now, cfg)
	sShare, _, _ := authorReputationFromEvents([]domain.InteractionEvent{evShare}, now, cfg)
	if sShare <= sLike {
		t.Fatalf("share %d should beat like %d", sShare, sLike)
	}
}

func TestAuthorReputationFromEvents_UnlikeClawsBack(t *testing.T) {
	cfg := ReputationConfig{}.withDefaults()
	now := time.Now().UTC()
	author := uuid.New()
	events := []domain.InteractionEvent{
		{AuthorID: author, ActorID: uuid.New(), Kind: domain.InteractionLike, CreatedAt: now},
		{AuthorID: author, ActorID: uuid.New(), Kind: domain.InteractionUnlike, CreatedAt: now},
	}
	score, _, likes := authorReputationFromEvents(events, now, cfg)
	if likes != 0 {
		t.Fatalf("likes=%d", likes)
	}
	if score != 0 {
		t.Fatalf("expected 0 after clawback, got %d", score)
	}
}

func TestAuthorReputationFromEvents_CreateCounted(t *testing.T) {
	cfg := ReputationConfig{}.withDefaults()
	now := time.Now().UTC()
	author := uuid.New()
	events := []domain.InteractionEvent{
		{AuthorID: author, ActorID: author, Kind: domain.InteractionCreate, CreatedAt: now},
	}
	score, posts, _ := authorReputationFromEvents(events, now, cfg)
	if posts != 1 {
		t.Fatalf("posts=%d", posts)
	}
	if score <= 0 {
		t.Fatalf("create should yield tiny positive score, got %d", score)
	}
}
