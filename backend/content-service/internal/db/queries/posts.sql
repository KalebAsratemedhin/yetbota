-- name: CreatePost :one
INSERT INTO posts (
    id, author_id, caption, lat, lng, media_mode, search_vector
) VALUES (
    $1, $2, $3, $4, $5, $6,
    to_tsvector('english', coalesce($3, '') || ' ' || coalesce(sqlc.arg(hashtag_text)::text, ''))
)
RETURNING *;

-- name: GetPostByID :one
SELECT * FROM posts WHERE id = $1;

-- name: ListPostsByAuthor :many
SELECT * FROM posts
WHERE author_id = sqlc.arg(author_id)
  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR (created_at, id) < (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(result_limit);

-- name: InsertPostMedia :one
INSERT INTO post_media (
    id, post_id, kind, cloudinary_public_id, url, sort_order, duration_ms, width, height, mime
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: ListMediaByPostID :many
SELECT * FROM post_media
WHERE post_id = $1
ORDER BY sort_order ASC, id ASC;

-- name: ListMediaByPostIDs :many
SELECT * FROM post_media
WHERE post_id = ANY(sqlc.arg(post_ids)::uuid[])
ORDER BY post_id, sort_order ASC, id ASC;

-- name: UpsertHashtag :one
INSERT INTO hashtags (id, tag)
VALUES ($1, $2)
ON CONFLICT (tag) DO UPDATE SET tag = EXCLUDED.tag
RETURNING *;

-- name: LinkPostHashtag :exec
INSERT INTO post_hashtags (post_id, hashtag_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListHashtagsByPostID :many
SELECT h.* FROM hashtags h
INNER JOIN post_hashtags ph ON ph.hashtag_id = h.id
WHERE ph.post_id = $1
ORDER BY h.tag ASC;

-- name: ListHashtagsByPostIDs :many
SELECT ph.post_id, h.id, h.tag, h.created_at
FROM hashtags h
INNER JOIN post_hashtags ph ON ph.hashtag_id = h.id
WHERE ph.post_id = ANY(sqlc.arg(post_ids)::uuid[])
ORDER BY ph.post_id, h.tag ASC;

-- name: UpsertReaction :one
INSERT INTO post_reactions (post_id, user_id, reaction)
VALUES ($1, $2, $3)
ON CONFLICT (post_id, user_id) DO UPDATE SET reaction = EXCLUDED.reaction, created_at = NOW()
RETURNING *;

-- name: DeleteReaction :exec
DELETE FROM post_reactions WHERE post_id = $1 AND user_id = $2;

-- name: GetReaction :one
SELECT * FROM post_reactions WHERE post_id = $1 AND user_id = $2;

-- name: AdjustPostLikeCount :one
UPDATE posts SET like_count = like_count + $2 WHERE id = $1 RETURNING *;

-- name: AdjustPostDislikeCount :one
UPDATE posts SET dislike_count = dislike_count + $2 WHERE id = $1 RETURNING *;

-- name: CreateComment :one
INSERT INTO post_comments (id, post_id, author_id, parent_id, body)
VALUES ($1, $2, $3, sqlc.narg(parent_id), $4)
RETURNING *;

-- name: GetCommentByID :one
SELECT * FROM post_comments WHERE id = $1;

-- name: ListTopLevelCommentsByPost :many
SELECT * FROM post_comments
WHERE post_id = sqlc.arg(post_id)
  AND parent_id IS NULL
  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR (created_at, id) < (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(result_limit);

-- name: ListRepliesByParent :many
SELECT * FROM post_comments
WHERE parent_id = sqlc.arg(parent_id)
  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR (created_at, id) < (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(result_limit);

-- name: IncrementCommentCount :one
UPDATE posts SET comment_count = comment_count + 1 WHERE id = $1 RETURNING *;

-- name: IncrementReplyCount :one
UPDATE post_comments SET reply_count = reply_count + 1 WHERE id = $1 RETURNING *;

-- name: GetCommentReaction :one
SELECT * FROM comment_reactions WHERE comment_id = $1 AND user_id = $2;

-- name: UpsertCommentReaction :one
INSERT INTO comment_reactions (comment_id, user_id, reaction)
VALUES ($1, $2, $3)
ON CONFLICT (comment_id, user_id) DO UPDATE SET reaction = EXCLUDED.reaction, created_at = NOW()
RETURNING *;

-- name: DeleteCommentReaction :exec
DELETE FROM comment_reactions WHERE comment_id = $1 AND user_id = $2;

-- name: AdjustCommentLikeCount :one
UPDATE post_comments SET like_count = like_count + $2 WHERE id = $1 RETURNING *;

-- name: AdjustCommentDislikeCount :one
UPDATE post_comments SET dislike_count = dislike_count + $2 WHERE id = $1 RETURNING *;

-- name: AddBookmark :execrows
INSERT INTO post_bookmarks (user_id, post_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveBookmark :execrows
DELETE FROM post_bookmarks WHERE user_id = $1 AND post_id = $2;

-- name: IncrementBookmarkCount :one
UPDATE posts SET bookmark_count = bookmark_count + $2 WHERE id = $1 RETURNING *;

-- name: ListBookmarkedPosts :many
SELECT
    p.id, p.author_id, p.caption, p.lat, p.lng, p.media_mode,
    p.like_count, p.dislike_count, p.comment_count, p.share_count, p.bookmark_count,
    p.created_at,
    b.created_at AS bookmarked_at
FROM posts p
INNER JOIN post_bookmarks b ON b.post_id = p.id
WHERE b.user_id = sqlc.arg(user_id)
  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR (b.created_at, b.post_id) < (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY b.created_at DESC, b.post_id DESC
LIMIT sqlc.arg(result_limit);

-- name: IncrementShareCount :one
UPDATE posts SET share_count = share_count + 1 WHERE id = $1 RETURNING *;

-- name: SearchPosts :many
SELECT * FROM posts
WHERE search_vector @@ plainto_tsquery('english', sqlc.arg(query)::text)
  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR (created_at, id) < (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(result_limit);

-- name: ListPostsByHashtag :many
SELECT p.* FROM posts p
INNER JOIN post_hashtags ph ON ph.post_id = p.id
INNER JOIN hashtags h ON h.id = ph.hashtag_id
WHERE h.tag = sqlc.arg(tag)
  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR (p.created_at, p.id) < (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY p.created_at DESC, p.id DESC
LIMIT sqlc.arg(result_limit);

-- name: UpdatePostSearchVector :exec
UPDATE posts
SET search_vector = to_tsvector('english', coalesce(caption, '') || ' ' || coalesce(sqlc.arg(hashtag_text)::text, ''))
WHERE id = sqlc.arg(id);
