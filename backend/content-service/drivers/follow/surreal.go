package follow

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	surrealdb "github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"

	"github.com/yetbota/content-service/internal/domain/apperr"
)

type Store struct {
	db *surrealdb.DB
}

func New(db *surrealdb.DB) *Store {
	return &Store{db: db}
}

// FollowingIDs returns all user IDs that userID follows (paginated in batches of 200).
func (s *Store) FollowingIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	var out []uuid.UUID
	var cursorTime any
	var cursorPeer string
	for {
		vars := map[string]any{
			"user": models.NewRecordID("user", userID.String()),
			"lim":  200,
		}
		cursorClause := ""
		if cursorPeer != "" {
			vars["cursor_time"] = cursorTime
			vars["cursor_peer_id"] = cursorPeer
			cursorClause = `
				AND (
					created_at < $cursor_time
					OR (created_at = $cursor_time AND meta::id(out) < $cursor_peer_id)
				)`
		}
		q := fmt.Sprintf(`
			SELECT meta::id(out) AS peer_id, created_at FROM follows
			WHERE in = $user %s
			ORDER BY created_at DESC, peer_id DESC
			LIMIT $lim
		`, cursorClause)
		res, err := surrealdb.Query[[]map[string]any](ctx, s.db, q, vars)
		if err != nil {
			msg := err.Error()
			if strings.Contains(msg, "does not exist") {
				return out, nil
			}
			return nil, apperr.Wrap(err, apperr.CodeInternal, "list following")
		}
		rows := lastResult(res)
		if len(rows) == 0 {
			break
		}
		for _, row := range rows {
			raw, _ := row["peer_id"].(string)
			id, err := uuid.Parse(strings.TrimSpace(raw))
			if err != nil {
				continue
			}
			out = append(out, id)
		}
		if len(rows) < 200 {
			break
		}
		last := rows[len(rows)-1]
		cursorTime = last["created_at"]
		cursorPeer, _ = last["peer_id"].(string)
	}
	return out, nil
}

func lastResult[T any](res *[]surrealdb.QueryResult[T]) T {
	var zero T
	if res == nil || len(*res) == 0 {
		return zero
	}
	return (*res)[len(*res)-1].Result
}
