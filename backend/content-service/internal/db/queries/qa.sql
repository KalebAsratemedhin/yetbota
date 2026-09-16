-- name: CreateQuestion :one
INSERT INTO questions (
    id, author_id, body, post_id, lat, lng
) VALUES (
    $1, $2, $3, sqlc.narg(post_id), sqlc.narg(lat), sqlc.narg(lng)
)
RETURNING *;

-- name: GetQuestionByID :one
SELECT * FROM questions WHERE id = $1;

-- name: ListQuestions :many
SELECT * FROM questions
WHERE (
    sqlc.narg(post_id)::uuid IS NULL
    OR post_id = sqlc.narg(post_id)
  )
  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR (created_at, id) < (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(result_limit);

-- name: CloseQuestion :one
UPDATE questions
SET closed_at = NOW(), closed_by = sqlc.arg(closed_by)
WHERE id = sqlc.arg(id)
  AND closed_at IS NULL
RETURNING *;

-- name: UpdateQuestion :one
UPDATE questions
SET body = sqlc.arg(body),
    post_id = sqlc.narg(post_id),
    lat = sqlc.narg(lat),
    lng = sqlc.narg(lng)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateAnswer :one
UPDATE answers
SET body = sqlc.arg(body)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: GetQuestionVote :one
SELECT * FROM question_votes
WHERE question_id = $1 AND user_id = $2;

-- name: UpsertQuestionVote :one
INSERT INTO question_votes (question_id, user_id, vote)
VALUES ($1, $2, $3)
ON CONFLICT (question_id, user_id) DO UPDATE SET vote = EXCLUDED.vote
RETURNING *;

-- name: DeleteQuestionVote :exec
DELETE FROM question_votes
WHERE question_id = $1 AND user_id = $2;

-- name: AdjustQuestionUpvoteCount :one
UPDATE questions
SET upvote_count = upvote_count + sqlc.arg(delta)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: AdjustQuestionDownvoteCount :one
UPDATE questions
SET downvote_count = downvote_count + sqlc.arg(delta)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: AdjustQuestionAnswerCount :one
UPDATE questions
SET answer_count = answer_count + sqlc.arg(delta)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: CreateAnswer :one
INSERT INTO answers (id, question_id, author_id, body)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAnswerByID :one
SELECT * FROM answers WHERE id = $1;

-- name: ListAnswersByQuestion :many
SELECT * FROM answers
WHERE question_id = sqlc.arg(question_id)
  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR (created_at, id) < (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(result_limit);

-- name: GetAnswerVote :one
SELECT * FROM answer_votes
WHERE answer_id = $1 AND user_id = $2;

-- name: UpsertAnswerVote :one
INSERT INTO answer_votes (answer_id, user_id, vote)
VALUES ($1, $2, $3)
ON CONFLICT (answer_id, user_id) DO UPDATE SET vote = EXCLUDED.vote
RETURNING *;

-- name: DeleteAnswerVote :exec
DELETE FROM answer_votes
WHERE answer_id = $1 AND user_id = $2;

-- name: AdjustAnswerUpvoteCount :one
UPDATE answers
SET upvote_count = upvote_count + sqlc.arg(delta)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: AdjustAnswerDownvoteCount :one
UPDATE answers
SET downvote_count = downvote_count + sqlc.arg(delta)
WHERE id = sqlc.arg(id)
RETURNING *;
