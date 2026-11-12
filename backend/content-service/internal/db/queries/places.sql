-- name: CreatePlace :one
INSERT INTO places (
    id, name, lat, lng, summary, summary_model, post_count, embedding
) VALUES (
    $1, $2, $3, $4, COALESCE(sqlc.narg(summary), ''), COALESCE(sqlc.narg(summary_model), ''), 0, sqlc.narg(embedding)
)
RETURNING *;

-- name: GetPlaceByID :one
SELECT * FROM places WHERE id = $1;

-- name: GetPlaceByPostID :one
SELECT p.* FROM places p
INNER JOIN place_posts pp ON pp.place_id = p.id
WHERE pp.post_id = $1;

-- name: UpdatePlaceSummary :one
UPDATE places
SET summary = sqlc.arg(summary),
    summary_model = sqlc.arg(summary_model),
    summary_updated_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdatePlaceEmbedding :one
UPDATE places
SET embedding = sqlc.arg(embedding),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: AdjustPlacePostCount :one
UPDATE places
SET post_count = post_count + sqlc.arg(delta),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ListPlaces :many
SELECT * FROM places
WHERE (
    sqlc.narg(query)::text IS NULL
    OR lower(name) LIKE '%' || lower(sqlc.narg(query)::text) || '%'
  )
  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR (created_at, id) < (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(result_limit);

-- name: ListPlacesNearby :many
-- Bounding-box prefilter (~degrees), then Haversine meters in app or SQL.
SELECT * FROM places
WHERE lat BETWEEN sqlc.arg(min_lat) AND sqlc.arg(max_lat)
  AND lng BETWEEN sqlc.arg(min_lng) AND sqlc.arg(max_lng)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(result_limit);

-- name: ListTopPlaces :many
-- Absolute engagement = sum of attached posts' (likes+comments+shares+bookmarks).
-- cover_url = first media of the highest-engagement attached post (optional).
WITH place_eng AS (
  SELECT
    p.id,
    p.name,
    p.lat,
    p.lng,
    p.summary,
    p.summary_model,
    p.summary_updated_at,
    p.post_count,
    p.embedding,
    p.created_at,
    p.updated_at,
    COALESCE(SUM(po.like_count + po.comment_count + po.share_count + po.bookmark_count), 0)::bigint AS engagement_score
  FROM places p
  LEFT JOIN place_posts pp ON pp.place_id = p.id
  LEFT JOIN posts po ON po.id = pp.post_id
  GROUP BY p.id
)
SELECT
  pe.id,
  pe.name,
  pe.lat,
  pe.lng,
  pe.summary,
  pe.summary_model,
  pe.summary_updated_at,
  pe.post_count,
  pe.embedding,
  pe.created_at,
  pe.updated_at,
  pe.engagement_score,
  COALESCE((
    SELECT pm.url
    FROM place_posts pp2
    INNER JOIN posts po2 ON po2.id = pp2.post_id
    LEFT JOIN LATERAL (
      SELECT m.url
      FROM post_media m
      WHERE m.post_id = po2.id
      ORDER BY m.sort_order ASC, m.id ASC
      LIMIT 1
    ) pm ON true
    WHERE pp2.place_id = pe.id
    ORDER BY (po2.like_count + po2.comment_count + po2.share_count + po2.bookmark_count) DESC, po2.id DESC
    LIMIT 1
  ), '')::text AS cover_url
FROM place_eng pe
WHERE (
    sqlc.narg(cursor_engagement)::bigint IS NULL
    OR (pe.engagement_score, pe.id) < (sqlc.narg(cursor_engagement)::bigint, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY pe.engagement_score DESC, pe.id DESC
LIMIT sqlc.arg(result_limit);

-- name: AttachPostToPlace :one
INSERT INTO place_posts (place_id, post_id, attach_score, attach_method)
VALUES ($1, $2, $3, $4)
ON CONFLICT (post_id) DO UPDATE
SET place_id = EXCLUDED.place_id,
    attach_score = EXCLUDED.attach_score,
    attach_method = EXCLUDED.attach_method,
    attached_at = NOW()
RETURNING *;

-- name: GetPlacePost :one
SELECT * FROM place_posts WHERE post_id = $1;

-- name: ListPlacePostIDs :many
SELECT post_id, attached_at FROM place_posts
WHERE place_id = sqlc.arg(place_id)
  AND (
    sqlc.narg(cursor_attached_at)::timestamptz IS NULL
    OR (attached_at, post_id) < (sqlc.narg(cursor_attached_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY attached_at DESC, post_id DESC
LIMIT sqlc.arg(result_limit);

-- name: CreatePlaceEntry :one
INSERT INTO place_entries (
    id, place_id, category, body, source, source_post_id, author_id
) VALUES (
    $1, $2, $3, $4, $5, sqlc.narg(source_post_id), sqlc.narg(author_id)
)
RETURNING *;

-- name: ListPlaceEntries :many
SELECT * FROM place_entries
WHERE place_id = sqlc.arg(place_id)
  AND (
    sqlc.narg(category)::text IS NULL
    OR category = sqlc.narg(category)
  )
  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR (created_at, id) < (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(result_limit);

-- name: FindSimilarPlaceEntry :one
SELECT * FROM place_entries
WHERE place_id = sqlc.arg(place_id)
  AND category = sqlc.arg(category)
  AND lower(body) = lower(sqlc.arg(body))
LIMIT 1;
