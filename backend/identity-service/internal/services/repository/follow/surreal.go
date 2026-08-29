package follow

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	surrealdb "github.com/surrealdb/surrealdb.go"
	"github.com/surrealdb/surrealdb.go/pkg/models"

	"github.com/yetbota/identity-service/internal/domain/apperr"
	domain "github.com/yetbota/identity-service/internal/domain/follow"
)

const (
	defaultPageSize = 20
	maxPageSize     = 50
)

var _ domain.Store = (*Store)(nil)

type Store struct {
	db *surrealdb.DB
}

func NewStore(db *surrealdb.DB) *Store {
	return &Store{db: db}
}

func userRID(id uuid.UUID) models.RecordID {
	return models.NewRecordID("user", id.String())
}

func (s *Store) Follow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	res, err := surrealdb.Query[any](ctx, s.db, `
		UPSERT $follower;
		UPSERT $followee;
		DELETE $follower->follows WHERE out = $followee;
		RELATE $follower->follows->$followee SET created_at = time::now();
	`, map[string]any{
		"follower": userRID(followerID),
		"followee": userRID(followeeID),
	})
	if err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "surreal follow")
	}
	return firstQueryErr(res)
}

func (s *Store) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	res, err := surrealdb.Query[any](ctx, s.db, `
		DELETE $follower->follows WHERE out = $followee;
	`, map[string]any{
		"follower": userRID(followerID),
		"followee": userRID(followeeID),
	})
	if err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "surreal unfollow")
	}
	return firstQueryErr(res)
}

type edgeRow struct {
	In        models.RecordID         `json:"in"`
	Out       models.RecordID         `json:"out"`
	CreatedAt models.CustomDateTime   `json:"created_at"`
}

func (s *Store) Followers(ctx context.Context, userID uuid.UUID, p domain.ListParams) (domain.EdgePage, error) {
	return s.listEdges(ctx, userID, p, true)
}

func (s *Store) Following(ctx context.Context, userID uuid.UUID, p domain.ListParams) (domain.EdgePage, error) {
	return s.listEdges(ctx, userID, p, false)
}

func (s *Store) listEdges(ctx context.Context, userID uuid.UUID, p domain.ListParams, followers bool) (domain.EdgePage, error) {
	limit := normalizeLimit(p.Limit)
	vars := map[string]any{
		"user":  userRID(userID),
		"limit": limit + 1, // read one extra to detect next page
	}

	filter := "out = $user"
	peerField := "in"
	if !followers {
		filter = "in = $user"
		peerField = "out"
	}

	cursorClause := ""
	if cur, ok, err := decodeCursor(p.Cursor); err != nil {
		return domain.EdgePage{}, err
	} else if ok {
		vars["cursor_time"] = models.CustomDateTime{Time: cur.CreatedAt}
		vars["cursor_peer_id"] = cur.UserID.String()
		// Keyset: (created_at, peer_id) strictly older than cursor in DESC order.
		cursorClause = fmt.Sprintf(`
			AND (
				created_at < $cursor_time
				OR (created_at = $cursor_time AND meta::id(%s) < $cursor_peer_id)
			)`, peerField)
	}

	// Surreal requires ORDER BY expressions to appear in the SELECT list.
	sql := fmt.Sprintf(`
		SELECT in, out, created_at, meta::id(%s) AS peer_id FROM follows
		WHERE %s %s
		ORDER BY created_at DESC, peer_id DESC
		LIMIT $limit;
	`, peerField, filter, cursorClause)

	res, err := surrealdb.Query[[]edgeRow](ctx, s.db, sql, vars)
	if err != nil {
		if isMissingTable(err) {
			return domain.EdgePage{UserIDs: []uuid.UUID{}}, nil
		}
		return domain.EdgePage{}, apperr.Wrap(err, apperr.CodeInternal, "surreal list follows")
	}
	if err := firstQueryErr(res); err != nil {
		if isMissingTable(err) {
			return domain.EdgePage{UserIDs: []uuid.UUID{}}, nil
		}
		return domain.EdgePage{}, err
	}

	rows := lastResult(res)
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		rid := row.Out
		if followers {
			rid = row.In
		}
		id, err := parseRecordUUID(rid)
		if err != nil {
			return domain.EdgePage{}, apperr.Wrap(err, apperr.CodeInternal, "parse surreal user id")
		}
		ids = append(ids, id)
	}

	page := domain.EdgePage{UserIDs: ids}
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		peer := last.Out
		if followers {
			peer = last.In
		}
		peerID, err := parseRecordUUID(peer)
		if err != nil {
			return domain.EdgePage{}, apperr.Wrap(err, apperr.CodeInternal, "parse cursor peer")
		}
		page.NextCursor = encodeCursor(domain.Cursor{
			CreatedAt: last.CreatedAt.Time,
			UserID:    peerID,
		})
	}
	return page, nil
}

func (s *Store) CountFollowers(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.count(ctx, `SELECT count() AS count FROM follows WHERE out = $user GROUP ALL;`, userID)
}

func (s *Store) CountFollowing(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.count(ctx, `SELECT count() AS count FROM follows WHERE in = $user GROUP ALL;`, userID)
}

type countRow struct {
	Count int64 `json:"count"`
}

func (s *Store) count(ctx context.Context, sql string, userID uuid.UUID) (int64, error) {
	res, err := surrealdb.Query[[]countRow](ctx, s.db, sql, map[string]any{
		"user": userRID(userID),
	})
	if err != nil {
		if isMissingTable(err) {
			return 0, nil
		}
		return 0, apperr.Wrap(err, apperr.CodeInternal, "surreal count follows")
	}
	if err := firstQueryErr(res); err != nil {
		if isMissingTable(err) {
			return 0, nil
		}
		return 0, err
	}
	rows := lastResult(res)
	if len(rows) == 0 {
		return 0, nil
	}
	return rows[0].Count, nil
}

func isMissingTable(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "does not exist") && strings.Contains(msg, "follows")
}

func normalizeLimit(n int) int {
	if n <= 0 {
		return defaultPageSize
	}
	if n > maxPageSize {
		return maxPageSize
	}
	return n
}

func encodeCursor(c domain.Cursor) string {
	raw := fmt.Sprintf("%d|%s", c.CreatedAt.UTC().UnixNano(), c.UserID.String())
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(raw string) (domain.Cursor, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return domain.Cursor{}, false, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return domain.Cursor{}, false, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return domain.Cursor{}, false, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	ns, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return domain.Cursor{}, false, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return domain.Cursor{}, false, apperr.Validation(apperr.Field{
			Field: "cursor", Code: "INVALID", Message: "cursor is invalid",
		})
	}
	return domain.Cursor{CreatedAt: time.Unix(0, ns).UTC(), UserID: id}, true, nil
}

func firstQueryErr[T any](res *[]surrealdb.QueryResult[T]) error {
	if res == nil {
		return nil
	}
	for _, qr := range *res {
		if qr.Error != nil {
			return apperr.Wrap(qr.Error, apperr.CodeInternal, "surreal query")
		}
		if qr.Status != "" && qr.Status != "OK" {
			return apperr.New(apperr.CodeInternal, "surreal query status "+qr.Status)
		}
	}
	return nil
}

func lastResult[T any](res *[]surrealdb.QueryResult[T]) T {
	var zero T
	if res == nil || len(*res) == 0 {
		return zero
	}
	return (*res)[len(*res)-1].Result
}

func parseRecordUUID(rid models.RecordID) (uuid.UUID, error) {
	switch v := rid.ID.(type) {
	case string:
		return uuid.Parse(v)
	default:
		s := strings.TrimSpace(fmt.Sprint(v))
		return uuid.Parse(s)
	}
}
