package post

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yetbota/content-service/internal/db"
	"github.com/yetbota/content-service/internal/domain/apperr"
	domain "github.com/yetbota/content-service/internal/domain/post"
)

var _ domain.Repository = (*Repo)(nil)

type Repo struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool, q: db.New(pool)}
}

func (r *Repo) Create(ctx context.Context, in domain.CreateInput) (domain.Post, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "begin tx")
	}
	defer tx.Rollback(ctx)
	q := r.q.WithTx(tx)

	postID := uuid.New()
	tagText := strings.Join(in.Hashtags, " ")
	row, err := q.CreatePost(ctx, db.CreatePostParams{
		ID:          pgUUID(postID),
		AuthorID:    pgUUID(in.AuthorID),
		Caption:     in.Caption,
		Lat:         in.Lat,
		Lng:         in.Lng,
		MediaMode:   string(in.MediaMode),
		HashtagText: tagText,
	})
	if err != nil {
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "create post")
	}

	mediaOut := make([]domain.Media, 0, len(in.Media))
	for _, m := range in.Media {
		mid := uuid.New()
		mr, err := q.InsertPostMedia(ctx, db.InsertPostMediaParams{
			ID:                 pgUUID(mid),
			PostID:             pgUUID(postID),
			Kind:               string(m.Kind),
			CloudinaryPublicID: m.CloudinaryPublicID,
			Url:                m.URL,
			SortOrder:          int32(m.SortOrder),
			DurationMs:         int32(m.DurationMs),
			Width:              int32(m.Width),
			Height:             int32(m.Height),
			Mime:               m.Mime,
		})
		if err != nil {
			return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "insert media")
		}
		mediaOut = append(mediaOut, mediaFromDB(mr))
	}

	tags := make([]string, 0, len(in.Hashtags))
	for _, tag := range in.Hashtags {
		hid := uuid.New()
		h, err := q.UpsertHashtag(ctx, db.UpsertHashtagParams{ID: pgUUID(hid), Tag: tag})
		if err != nil {
			return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "upsert hashtag")
		}
		hidUUID, err := uuidFromPg(h.ID)
		if err != nil {
			return domain.Post{}, err
		}
		if err := q.LinkPostHashtag(ctx, db.LinkPostHashtagParams{
			PostID:    pgUUID(postID),
			HashtagID: pgUUID(hidUUID),
		}); err != nil {
			return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "link hashtag")
		}
		tags = append(tags, h.Tag)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "commit post")
	}
	p, err := postFromDB(row)
	if err != nil {
		return domain.Post{}, err
	}
	p.Media = mediaOut
	p.Hashtags = tags
	return p, nil
}

func (r *Repo) ByID(ctx context.Context, id uuid.UUID) (domain.Post, error) {
	row, err := r.q.GetPostByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Post{}, apperr.ErrNotFound
		}
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "get post")
	}
	return r.hydrateOne(ctx, row)
}

func (r *Repo) ListByAuthor(ctx context.Context, authorID uuid.UUID, p domain.ListParams) (domain.Page, error) {
	rows, err := r.q.ListPostsByAuthor(ctx, db.ListPostsByAuthorParams{
		AuthorID:        pgUUID(authorID),
		CursorCreatedAt: cursorTime(p.Cursor),
		CursorID:        cursorID(p.Cursor),
		ResultLimit:     int32(p.Limit + 1),
	})
	if err != nil {
		return domain.Page{}, apperr.Wrap(err, apperr.CodeInternal, "list author posts")
	}
	return r.hydratePage(ctx, rows, p.Limit)
}

func (r *Repo) SetReaction(ctx context.Context, postID, userID uuid.UUID, reaction *domain.Reaction) (domain.Post, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "begin reaction tx")
	}
	defer tx.Rollback(ctx)
	q := r.q.WithTx(tx)

	if _, err := q.GetPostByID(ctx, pgUUID(postID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Post{}, apperr.ErrNotFound
		}
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "get post")
	}

	prev, err := q.GetReaction(ctx, db.GetReactionParams{PostID: pgUUID(postID), UserID: pgUUID(userID)})
	hasPrev := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "get reaction")
	}

	likeDelta, dislikeDelta := int64(0), int64(0)
	if hasPrev {
		switch prev.Reaction {
		case "like":
			likeDelta--
		case "dislike":
			dislikeDelta--
		}
		if err := q.DeleteReaction(ctx, db.DeleteReactionParams{PostID: pgUUID(postID), UserID: pgUUID(userID)}); err != nil {
			return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "delete reaction")
		}
	}
	if reaction != nil {
		switch *reaction {
		case domain.ReactionLike:
			likeDelta++
		case domain.ReactionDislike:
			dislikeDelta++
		}
		if _, err := q.UpsertReaction(ctx, db.UpsertReactionParams{
			PostID: pgUUID(postID), UserID: pgUUID(userID), Reaction: string(*reaction),
		}); err != nil {
			return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "upsert reaction")
		}
	}
	if likeDelta != 0 {
		if _, err := q.AdjustPostLikeCount(ctx, db.AdjustPostLikeCountParams{ID: pgUUID(postID), LikeCount: likeDelta}); err != nil {
			return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "adjust likes")
		}
	}
	if dislikeDelta != 0 {
		if _, err := q.AdjustPostDislikeCount(ctx, db.AdjustPostDislikeCountParams{ID: pgUUID(postID), DislikeCount: dislikeDelta}); err != nil {
			return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "adjust dislikes")
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "commit reaction")
	}
	return r.ByID(ctx, postID)
}

func (r *Repo) AddComment(ctx context.Context, postID, authorID uuid.UUID, body string, parentID *uuid.UUID) (domain.Comment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Comment{}, apperr.Wrap(err, apperr.CodeInternal, "begin comment tx")
	}
	defer tx.Rollback(ctx)
	q := r.q.WithTx(tx)
	if _, err := q.GetPostByID(ctx, pgUUID(postID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Comment{}, apperr.ErrNotFound
		}
		return domain.Comment{}, apperr.Wrap(err, apperr.CodeInternal, "get post")
	}
	var parentPg pgtype.UUID
	if parentID != nil {
		parent, err := q.GetCommentByID(ctx, pgUUID(*parentID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.Comment{}, apperr.Validation(apperr.Field{
					Field: "parent_id", Code: "NOT_FOUND", Message: "parent comment not found",
				})
			}
			return domain.Comment{}, apperr.Wrap(err, apperr.CodeInternal, "get parent comment")
		}
		parentPostID, err := uuidFromPg(parent.PostID)
		if err != nil {
			return domain.Comment{}, err
		}
		if parentPostID != postID {
			return domain.Comment{}, apperr.Validation(apperr.Field{
				Field: "parent_id", Code: "INVALID", Message: "parent comment must belong to the same post",
			})
		}
		if parent.ParentID.Valid {
			return domain.Comment{}, apperr.Validation(apperr.Field{
				Field: "parent_id", Code: "INVALID", Message: "cannot reply to a reply",
			})
		}
		parentPg = pgUUID(*parentID)
	}
	cid := uuid.New()
	row, err := q.CreateComment(ctx, db.CreateCommentParams{
		ID: pgUUID(cid), PostID: pgUUID(postID), AuthorID: pgUUID(authorID),
		Body: body, ParentID: parentPg,
	})
	if err != nil {
		return domain.Comment{}, apperr.Wrap(err, apperr.CodeInternal, "create comment")
	}
	if _, err := q.IncrementCommentCount(ctx, pgUUID(postID)); err != nil {
		return domain.Comment{}, apperr.Wrap(err, apperr.CodeInternal, "inc comment count")
	}
	if parentID != nil {
		if _, err := q.IncrementReplyCount(ctx, pgUUID(*parentID)); err != nil {
			return domain.Comment{}, apperr.Wrap(err, apperr.CodeInternal, "inc reply count")
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Comment{}, apperr.Wrap(err, apperr.CodeInternal, "commit comment")
	}
	return commentFromDB(row)
}

func (r *Repo) ListComments(ctx context.Context, postID uuid.UUID, p domain.ListParams) (domain.CommentPage, error) {
	if _, err := r.q.GetPostByID(ctx, pgUUID(postID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CommentPage{}, apperr.ErrNotFound
		}
		return domain.CommentPage{}, apperr.Wrap(err, apperr.CodeInternal, "get post")
	}
	rows, err := r.q.ListTopLevelCommentsByPost(ctx, db.ListTopLevelCommentsByPostParams{
		PostID:          pgUUID(postID),
		CursorCreatedAt: cursorTime(p.Cursor),
		CursorID:        cursorID(p.Cursor),
		ResultLimit:     int32(p.Limit + 1),
	})
	if err != nil {
		return domain.CommentPage{}, apperr.Wrap(err, apperr.CodeInternal, "list comments")
	}
	return commentPageFromRows(rows, p.Limit)
}

func (r *Repo) ListReplies(ctx context.Context, parentID uuid.UUID, p domain.ListParams) (domain.CommentPage, error) {
	if _, err := r.q.GetCommentByID(ctx, pgUUID(parentID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CommentPage{}, apperr.ErrNotFound
		}
		return domain.CommentPage{}, apperr.Wrap(err, apperr.CodeInternal, "get parent comment")
	}
	rows, err := r.q.ListRepliesByParent(ctx, db.ListRepliesByParentParams{
		ParentID:        pgUUID(parentID),
		CursorCreatedAt: cursorTime(p.Cursor),
		CursorID:        cursorID(p.Cursor),
		ResultLimit:     int32(p.Limit + 1),
	})
	if err != nil {
		return domain.CommentPage{}, apperr.Wrap(err, apperr.CodeInternal, "list replies")
	}
	return commentPageFromRows(rows, p.Limit)
}

func (r *Repo) CommentByID(ctx context.Context, id uuid.UUID) (domain.Comment, error) {
	row, err := r.q.GetCommentByID(ctx, pgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Comment{}, apperr.ErrNotFound
		}
		return domain.Comment{}, apperr.Wrap(err, apperr.CodeInternal, "get comment")
	}
	return commentFromDB(row)
}

func (r *Repo) SetCommentReaction(ctx context.Context, commentID, userID uuid.UUID, reaction *domain.Reaction) (domain.Comment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Comment{}, apperr.Wrap(err, apperr.CodeInternal, "begin comment reaction tx")
	}
	defer tx.Rollback(ctx)
	q := r.q.WithTx(tx)

	if _, err := q.GetCommentByID(ctx, pgUUID(commentID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Comment{}, apperr.ErrNotFound
		}
		return domain.Comment{}, apperr.Wrap(err, apperr.CodeInternal, "get comment")
	}

	prev, err := q.GetCommentReaction(ctx, db.GetCommentReactionParams{
		CommentID: pgUUID(commentID), UserID: pgUUID(userID),
	})
	hasPrev := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return domain.Comment{}, apperr.Wrap(err, apperr.CodeInternal, "get comment reaction")
	}

	likeDelta, dislikeDelta := int64(0), int64(0)
	if hasPrev {
		switch prev.Reaction {
		case "like":
			likeDelta--
		case "dislike":
			dislikeDelta--
		}
		if err := q.DeleteCommentReaction(ctx, db.DeleteCommentReactionParams{
			CommentID: pgUUID(commentID), UserID: pgUUID(userID),
		}); err != nil {
			return domain.Comment{}, apperr.Wrap(err, apperr.CodeInternal, "delete comment reaction")
		}
	}
	if reaction != nil {
		switch *reaction {
		case domain.ReactionLike:
			likeDelta++
		case domain.ReactionDislike:
			dislikeDelta++
		}
		if _, err := q.UpsertCommentReaction(ctx, db.UpsertCommentReactionParams{
			CommentID: pgUUID(commentID), UserID: pgUUID(userID), Reaction: string(*reaction),
		}); err != nil {
			return domain.Comment{}, apperr.Wrap(err, apperr.CodeInternal, "upsert comment reaction")
		}
	}
	if likeDelta != 0 {
		if _, err := q.AdjustCommentLikeCount(ctx, db.AdjustCommentLikeCountParams{
			ID: pgUUID(commentID), LikeCount: likeDelta,
		}); err != nil {
			return domain.Comment{}, apperr.Wrap(err, apperr.CodeInternal, "adjust comment likes")
		}
	}
	if dislikeDelta != 0 {
		if _, err := q.AdjustCommentDislikeCount(ctx, db.AdjustCommentDislikeCountParams{
			ID: pgUUID(commentID), DislikeCount: dislikeDelta,
		}); err != nil {
			return domain.Comment{}, apperr.Wrap(err, apperr.CodeInternal, "adjust comment dislikes")
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Comment{}, apperr.Wrap(err, apperr.CodeInternal, "commit comment reaction")
	}
	return r.CommentByID(ctx, commentID)
}

func commentPageFromRows(rows []db.PostComment, limit int) (domain.CommentPage, error) {
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	out := make([]domain.Comment, 0, len(rows))
	for _, row := range rows {
		c, err := commentFromDB(row)
		if err != nil {
			return domain.CommentPage{}, err
		}
		out = append(out, c)
	}
	page := domain.CommentPage{Comments: out}
	if hasMore && len(out) > 0 {
		last := out[len(out)-1]
		page.NextCursor = &domain.PageCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	return page, nil
}

func (r *Repo) AddBookmark(ctx context.Context, userID, postID uuid.UUID) (domain.Post, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "begin bookmark tx")
	}
	defer tx.Rollback(ctx)
	q := r.q.WithTx(tx)
	if _, err := q.GetPostByID(ctx, pgUUID(postID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Post{}, apperr.ErrNotFound
		}
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "get post")
	}
	n, err := q.AddBookmark(ctx, db.AddBookmarkParams{UserID: pgUUID(userID), PostID: pgUUID(postID)})
	if err != nil {
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "add bookmark")
	}
	if n > 0 {
		if _, err := q.IncrementBookmarkCount(ctx, db.IncrementBookmarkCountParams{ID: pgUUID(postID), BookmarkCount: 1}); err != nil {
			return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "inc bookmark count")
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "commit bookmark")
	}
	return r.ByID(ctx, postID)
}

func (r *Repo) RemoveBookmark(ctx context.Context, userID, postID uuid.UUID) (domain.Post, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "begin unbookmark tx")
	}
	defer tx.Rollback(ctx)
	q := r.q.WithTx(tx)
	n, err := q.RemoveBookmark(ctx, db.RemoveBookmarkParams{UserID: pgUUID(userID), PostID: pgUUID(postID)})
	if err != nil {
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "remove bookmark")
	}
	if n > 0 {
		if _, err := q.IncrementBookmarkCount(ctx, db.IncrementBookmarkCountParams{ID: pgUUID(postID), BookmarkCount: -1}); err != nil {
			return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "dec bookmark count")
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "commit unbookmark")
	}
	return r.ByID(ctx, postID)
}

func (r *Repo) ListBookmarks(ctx context.Context, userID uuid.UUID, p domain.ListParams) (domain.Page, error) {
	rows, err := r.q.ListBookmarkedPosts(ctx, db.ListBookmarkedPostsParams{
		UserID:          pgUUID(userID),
		CursorCreatedAt: cursorTime(p.Cursor),
		CursorID:        cursorID(p.Cursor),
		ResultLimit:     int32(p.Limit + 1),
	})
	if err != nil {
		return domain.Page{}, apperr.Wrap(err, apperr.CodeInternal, "list bookmarks")
	}
	hasMore := len(rows) > p.Limit
	if hasMore {
		rows = rows[:p.Limit]
	}
	if len(rows) == 0 {
		return domain.Page{Posts: []domain.Post{}}, nil
	}
	posts := make([]db.Post, 0, len(rows))
	bookmarkTimes := make([]pgtype.Timestamptz, 0, len(rows))
	ids := make([]pgtype.UUID, len(rows))
	for i, row := range rows {
		posts = append(posts, db.Post{
			ID: row.ID, AuthorID: row.AuthorID, Caption: row.Caption,
			Lat: row.Lat, Lng: row.Lng, MediaMode: row.MediaMode,
			LikeCount: row.LikeCount, DislikeCount: row.DislikeCount,
			CommentCount: row.CommentCount, ShareCount: row.ShareCount,
			BookmarkCount: row.BookmarkCount,
			CreatedAt:     row.CreatedAt,
		})
		bookmarkTimes = append(bookmarkTimes, row.BookmarkedAt)
		ids[i] = row.ID
	}
	page, err := r.hydratePosts(ctx, posts, ids, p.Limit, false)
	if err != nil {
		return domain.Page{}, err
	}
	if hasMore && len(page.Posts) > 0 {
		lastIdx := len(page.Posts) - 1
		page.NextCursor = &domain.PageCursor{
			CreatedAt: bookmarkTimes[lastIdx].Time,
			ID:        page.Posts[lastIdx].ID,
		}
	}
	return page, nil
}

func (r *Repo) Share(ctx context.Context, postID uuid.UUID) (domain.Post, error) {
	if _, err := r.q.IncrementShareCount(ctx, pgUUID(postID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Post{}, apperr.ErrNotFound
		}
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "share post")
	}
	return r.ByID(ctx, postID)
}

func (r *Repo) Search(ctx context.Context, query string, p domain.ListParams) (domain.Page, error) {
	rows, err := r.q.SearchPosts(ctx, db.SearchPostsParams{
		Query:           query,
		CursorCreatedAt: cursorTime(p.Cursor),
		CursorID:        cursorID(p.Cursor),
		ResultLimit:     int32(p.Limit + 1),
	})
	if err != nil {
		return domain.Page{}, apperr.Wrap(err, apperr.CodeInternal, "search posts")
	}
	return r.hydratePage(ctx, rows, p.Limit)
}

func (r *Repo) ListByHashtag(ctx context.Context, tag string, p domain.ListParams) (domain.Page, error) {
	rows, err := r.q.ListPostsByHashtag(ctx, db.ListPostsByHashtagParams{
		Tag:             tag,
		CursorCreatedAt: cursorTime(p.Cursor),
		CursorID:        cursorID(p.Cursor),
		ResultLimit:     int32(p.Limit + 1),
	})
	if err != nil {
		return domain.Page{}, apperr.Wrap(err, apperr.CodeInternal, "list by hashtag")
	}
	return r.hydratePage(ctx, rows, p.Limit)
}

func (r *Repo) ListByAuthorIDs(ctx context.Context, authorIDs []uuid.UUID, since time.Time, limit int) ([]domain.Post, error) {
	if len(authorIDs) == 0 || limit <= 0 {
		return []domain.Post{}, nil
	}
	ids := make([]pgtype.UUID, len(authorIDs))
	for i, id := range authorIDs {
		ids[i] = pgUUID(id)
	}
	rows, err := r.q.ListPostsByAuthorIDs(ctx, db.ListPostsByAuthorIDsParams{
		AuthorIds: ids, Since: pgtype.Timestamptz{Time: since, Valid: true}, ResultLimit: int32(limit),
	})
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternal, "list posts by authors")
	}
	page, err := r.hydratePosts(ctx, rows, postIDs(rows), limit, false)
	if err != nil {
		return nil, err
	}
	return page.Posts, nil
}

func (r *Repo) ListRecent(ctx context.Context, since time.Time, limit int) ([]domain.Post, error) {
	if limit <= 0 {
		return []domain.Post{}, nil
	}
	rows, err := r.q.ListRecentPosts(ctx, db.ListRecentPostsParams{
		Since: pgtype.Timestamptz{Time: since, Valid: true}, ResultLimit: int32(limit),
	})
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternal, "list recent posts")
	}
	page, err := r.hydratePosts(ctx, rows, postIDs(rows), limit, false)
	if err != nil {
		return nil, err
	}
	return page.Posts, nil
}

func (r *Repo) ListNearby(ctx context.Context, lat, lng, radiusKm float64, since time.Time, limit int) ([]domain.Post, error) {
	if limit <= 0 || radiusKm <= 0 {
		return []domain.Post{}, nil
	}
	dLat := radiusKm / 111.0
	cosLat := math.Cos(lat * math.Pi / 180)
	if cosLat < 0.01 {
		cosLat = 0.01
	}
	dLng := radiusKm / (111.0 * cosLat)
	rows, err := r.q.ListNearbyPosts(ctx, db.ListNearbyPostsParams{
		Since: pgtype.Timestamptz{Time: since, Valid: true},
		MinLat: lat - dLat, MaxLat: lat + dLat,
		MinLng: lng - dLng, MaxLng: lng + dLng,
		ResultLimit: int32(limit * 3), // over-fetch then filter by haversine
	})
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternal, "list nearby posts")
	}
	filtered := make([]db.Post, 0, len(rows))
	for _, row := range rows {
		if haversineKm(lat, lng, row.Lat, row.Lng) <= radiusKm {
			filtered = append(filtered, row)
			if len(filtered) >= limit {
				break
			}
		}
	}
	page, err := r.hydratePosts(ctx, filtered, postIDs(filtered), limit, false)
	if err != nil {
		return nil, err
	}
	return page.Posts, nil
}

func postIDs(rows []db.Post) []pgtype.UUID {
	ids := make([]pgtype.UUID, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	return ids
}

func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	toRad := math.Pi / 180
	dLat := (lat2 - lat1) * toRad
	dLng := (lng2 - lng1) * toRad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*toRad)*math.Cos(lat2*toRad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * R * math.Asin(math.Sqrt(a))
}

func (r *Repo) hydrateOne(ctx context.Context, row db.Post) (domain.Post, error) {
	p, err := postFromDB(row)
	if err != nil {
		return domain.Post{}, err
	}
	mediaRows, err := r.q.ListMediaByPostID(ctx, pgUUID(p.ID))
	if err != nil {
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "list media")
	}
	p.Media = make([]domain.Media, 0, len(mediaRows))
	for _, m := range mediaRows {
		p.Media = append(p.Media, mediaFromDB(m))
	}
	tags, err := r.q.ListHashtagsByPostID(ctx, pgUUID(p.ID))
	if err != nil {
		return domain.Post{}, apperr.Wrap(err, apperr.CodeInternal, "list hashtags")
	}
	p.Hashtags = make([]string, 0, len(tags))
	for _, t := range tags {
		p.Hashtags = append(p.Hashtags, t.Tag)
	}
	return p, nil
}

func (r *Repo) hydratePage(ctx context.Context, rows []db.Post, limit int) (domain.Page, error) {
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	ids := make([]pgtype.UUID, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	return r.hydratePosts(ctx, rows, ids, limit, hasMore)
}

func (r *Repo) hydratePosts(ctx context.Context, rows []db.Post, ids []pgtype.UUID, limit int, hasMore bool) (domain.Page, error) {
	_ = limit
	if len(rows) == 0 {
		return domain.Page{Posts: []domain.Post{}}, nil
	}
	posts := make([]domain.Post, 0, len(rows))
	for _, row := range rows {
		p, err := postFromDB(row)
		if err != nil {
			return domain.Page{}, err
		}
		posts = append(posts, p)
	}
	mediaRows, err := r.q.ListMediaByPostIDs(ctx, ids)
	if err != nil {
		return domain.Page{}, apperr.Wrap(err, apperr.CodeInternal, "batch media")
	}
	tagRows, err := r.q.ListHashtagsByPostIDs(ctx, ids)
	if err != nil {
		return domain.Page{}, apperr.Wrap(err, apperr.CodeInternal, "batch hashtags")
	}
	mediaBy := map[uuid.UUID][]domain.Media{}
	for _, m := range mediaRows {
		mm := mediaFromDB(m)
		mediaBy[mm.PostID] = append(mediaBy[mm.PostID], mm)
	}
	tagsBy := map[uuid.UUID][]string{}
	for _, t := range tagRows {
		pid, err := uuidFromPg(t.PostID)
		if err != nil {
			return domain.Page{}, err
		}
		tagsBy[pid] = append(tagsBy[pid], t.Tag)
	}
	for i := range posts {
		posts[i].Media = mediaBy[posts[i].ID]
		if posts[i].Media == nil {
			posts[i].Media = []domain.Media{}
		}
		posts[i].Hashtags = tagsBy[posts[i].ID]
		if posts[i].Hashtags == nil {
			posts[i].Hashtags = []string{}
		}
	}
	page := domain.Page{Posts: posts}
	if hasMore && len(posts) > 0 {
		last := posts[len(posts)-1]
		page.NextCursor = &domain.PageCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	return page, nil
}

func postFromDB(row db.Post) (domain.Post, error) {
	id, err := uuidFromPg(row.ID)
	if err != nil {
		return domain.Post{}, err
	}
	aid, err := uuidFromPg(row.AuthorID)
	if err != nil {
		return domain.Post{}, err
	}
	return domain.Post{
		ID: id, AuthorID: aid, Caption: row.Caption, Lat: row.Lat, Lng: row.Lng,
		MediaMode: domain.MediaMode(row.MediaMode),
		LikeCount: row.LikeCount, DislikeCount: row.DislikeCount,
		CommentCount: row.CommentCount, ShareCount: row.ShareCount, BookmarkCount: row.BookmarkCount,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func mediaFromDB(row db.PostMedium) domain.Media {
	id, _ := uuidFromPg(row.ID)
	pid, _ := uuidFromPg(row.PostID)
	return domain.Media{
		ID: id, PostID: pid, Kind: domain.MediaKind(row.Kind),
		CloudinaryPublicID: row.CloudinaryPublicID, URL: row.Url,
		SortOrder: int(row.SortOrder), DurationMs: int(row.DurationMs),
		Width: int(row.Width), Height: int(row.Height), Mime: row.Mime,
	}
}

func commentFromDB(row db.PostComment) (domain.Comment, error) {
	id, err := uuidFromPg(row.ID)
	if err != nil {
		return domain.Comment{}, err
	}
	pid, err := uuidFromPg(row.PostID)
	if err != nil {
		return domain.Comment{}, err
	}
	aid, err := uuidFromPg(row.AuthorID)
	if err != nil {
		return domain.Comment{}, err
	}
	c := domain.Comment{
		ID: id, PostID: pid, AuthorID: aid, Body: row.Body,
		LikeCount: row.LikeCount, DislikeCount: row.DislikeCount, ReplyCount: row.ReplyCount,
		CreatedAt: row.CreatedAt.Time,
	}
	if row.ParentID.Valid {
		parent, err := uuidFromPg(row.ParentID)
		if err != nil {
			return domain.Comment{}, err
		}
		c.ParentID = &parent
	}
	return c, nil
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func uuidFromPg(v pgtype.UUID) (uuid.UUID, error) {
	if !v.Valid {
		return uuid.Nil, apperr.New(apperr.CodeInternal, "invalid uuid")
	}
	return uuid.UUID(v.Bytes), nil
}

func cursorTime(c *domain.PageCursor) pgtype.Timestamptz {
	if c == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: c.CreatedAt, Valid: true}
}

func cursorID(c *domain.PageCursor) pgtype.UUID {
	if c == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgUUID(c.ID)
}
