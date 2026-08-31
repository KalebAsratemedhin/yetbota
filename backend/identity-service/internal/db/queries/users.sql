-- name: CreateUser :one
INSERT INTO users (id, username, password_hash, full_name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2
WHERE id = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET
    full_name = COALESCE(sqlc.narg('full_name'), full_name),
    phone = COALESCE(sqlc.narg('phone'), phone),
    address = COALESCE(sqlc.narg('address'), address),
    bio = COALESCE(sqlc.narg('bio'), bio)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: UpdateUserAvatar :one
UPDATE users
SET
    avatar_url = $2,
    avatar_public_id = $3
WHERE id = $1
RETURNING *;

-- name: SearchUsers :many
SELECT * FROM users
WHERE (
    username ILIKE '%' || sqlc.arg(query)::text || '%'
    OR full_name ILIKE '%' || sqlc.arg(query)::text || '%'
)
AND id <> sqlc.arg(exclude_id)
ORDER BY
    CASE WHEN username ILIKE sqlc.arg(query)::text || '%' THEN 0 ELSE 1 END,
    username ASC
LIMIT sqlc.arg(result_limit);

-- name: GetUsersByIDs :many
SELECT * FROM users
WHERE id = ANY(sqlc.arg(ids)::uuid[]);

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
