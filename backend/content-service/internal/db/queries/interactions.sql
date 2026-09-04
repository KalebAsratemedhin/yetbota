-- name: InsertPostInteractionEvent :one
INSERT INTO post_interaction_events (id, post_id, author_id, actor_id, kind)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListAuthorInteractionEventsSince :many
SELECT id, post_id, author_id, actor_id, kind, created_at
FROM post_interaction_events
WHERE author_id = sqlc.arg(author_id)
  AND created_at >= sqlc.arg(since)::timestamptz
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(result_limit);
