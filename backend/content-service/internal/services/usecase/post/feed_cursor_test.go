package post

import (
	"testing"

	"github.com/google/uuid"
)

func TestFeedCursorRoundTrip(t *testing.T) {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	cur := encodeFeedCursor(feedPhaseHome, 1.25, id)
	phase, score, gotID, hasCursor, phaseStart, err := decodeFeedCursor(cur)
	if err != nil {
		t.Fatal(err)
	}
	if phase != feedPhaseHome || score != 1.25 || gotID != id || !hasCursor || phaseStart {
		t.Fatalf("got phase=%s score=%v id=%s has=%v start=%v", phase, score, gotID, hasCursor, phaseStart)
	}

	start := encodeFeedPhaseStart(feedPhaseSuggested)
	phase, _, _, hasCursor, phaseStart, err = decodeFeedCursor(start)
	if err != nil {
		t.Fatal(err)
	}
	if phase != feedPhaseSuggested || !hasCursor || !phaseStart {
		t.Fatalf("phase start: phase=%s has=%v start=%v", phase, hasCursor, phaseStart)
	}

	// Legacy rank cursor → home
	legacy := encodeRankCursor(3.5, id)
	phase, score, gotID, hasCursor, phaseStart, err = decodeFeedCursor(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if phase != feedPhaseHome || score != 3.5 || gotID != id || !hasCursor || phaseStart {
		t.Fatalf("legacy: phase=%s score=%v id=%s has=%v start=%v", phase, score, gotID, hasCursor, phaseStart)
	}
}
