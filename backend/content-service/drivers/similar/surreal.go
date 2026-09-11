package similar

import (
	"context"
	"sort"
	"strings"

	"github.com/google/uuid"
	surrealdb "github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"

	"github.com/yetbota/content-service/internal/domain/apperr"
)

// Neighbor is a cosine-similar post from the Surreal similar graph (internal parse helper).
type Neighbor struct {
	PostID uuid.UUID
	Score  float64
}

type Store struct {
	db *surrealdb.DB
}

func New(db *surrealdb.DB) *Store {
	return &Store{db: db}
}

// Neighbors returns unique similar post IDs for seed, highest score first.
// Merges outbound (seed->similar) and inbound (<-similar) edges from ai-service.
func (s *Store) Neighbors(ctx context.Context, seedID uuid.UUID) (ids []uuid.UUID, scores []float64, err error) {
	if s == nil || s.db == nil {
		return nil, nil, nil
	}
	from := models.NewRecordID("post_embedding", seedID.String())
	best := map[uuid.UUID]float64{}

	outRows, err := s.queryPeers(ctx, `
		SELECT meta::id(out) AS peer_id, score FROM $from->similar;
	`, from)
	if err != nil {
		return nil, nil, err
	}
	for _, n := range outRows {
		if n.Score > best[n.PostID] {
			best[n.PostID] = n.Score
		}
	}

	inRows, err := s.queryPeers(ctx, `
		SELECT meta::id(in) AS peer_id, score FROM $from<-similar;
	`, from)
	if err != nil {
		return nil, nil, err
	}
	for _, n := range inRows {
		if n.Score > best[n.PostID] {
			best[n.PostID] = n.Score
		}
	}

	type pair struct {
		id    uuid.UUID
		score float64
	}
	ordered := make([]pair, 0, len(best))
	for id, score := range best {
		if id == seedID {
			continue
		}
		ordered = append(ordered, pair{id: id, score: score})
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].score == ordered[j].score {
			return ordered[i].id.String() > ordered[j].id.String()
		}
		return ordered[i].score > ordered[j].score
	})
	ids = make([]uuid.UUID, len(ordered))
	scores = make([]float64, len(ordered))
	for i, p := range ordered {
		ids[i] = p.id
		scores[i] = p.score
	}
	return ids, scores, nil
}

func (s *Store) queryPeers(ctx context.Context, q string, from models.RecordID) ([]Neighbor, error) {
	res, err := surrealdb.Query[[]map[string]any](ctx, s.db, q, map[string]any{"from": from})
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "does not exist") {
			return nil, nil
		}
		return nil, apperr.Wrap(err, apperr.CodeInternal, "list similar edges")
	}
	rows := lastResult(res)
	out := make([]Neighbor, 0, len(rows))
	for _, row := range rows {
		raw, _ := row["peer_id"].(string)
		id, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil {
			continue
		}
		score := asFloat(row["score"])
		out = append(out, Neighbor{PostID: id, Score: score})
	}
	return out, nil
}

func asFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case uint64:
		return float64(t)
	default:
		return 0
	}
}

func lastResult[T any](res *[]surrealdb.QueryResult[T]) T {
	var zero T
	if res == nil || len(*res) == 0 {
		return zero
	}
	return (*res)[len(*res)-1].Result
}
