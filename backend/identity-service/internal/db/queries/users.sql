-- name: CreateUser :one
INSERT INTO users (id, username, password_hash, full_name)
VALUES ($1, $2, $3, $4)
RETURNING id, username, password_hash, created_at, full_name, phone, address, bio;

-- name: GetUserByUsername :one
SELECT id, username, password_hash, created_at, full_name, phone, address, bio
FROM users
WHERE username = $1;

-- name: GetUserByID :one
SELECT id, username, password_hash, created_at, full_name, phone, address, bio
FROM users
WHERE id = $1;

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
RETURNING id, username, password_hash, created_at, full_name, phone, address, bio;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
