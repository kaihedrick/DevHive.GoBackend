-- name: GetUserByID :one
SELECT id, username, email, first_name, last_name, active, avatar_url, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, email, password_h, first_name, last_name, active, avatar_url, created_at, updated_at
FROM users
WHERE lower(username) = lower($1);

-- name: GetUserByEmail :one
SELECT id, username, email, password_h, first_name, last_name, active, avatar_url, created_at, updated_at
FROM users
WHERE lower(email) = lower($1);

-- name: CreateUser :one
INSERT INTO users (username, email, password_h, first_name, last_name)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, username, email, first_name, last_name, active, avatar_url, created_at, updated_at;

-- name: UpdateUser :one
UPDATE users
SET username = $2, email = $3, first_name = $4, last_name = $5, avatar_url = $6, updated_at = now()
WHERE id = $1
RETURNING id, username, email, first_name, last_name, active, avatar_url, created_at, updated_at;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_h = $2, updated_at = now()
WHERE id = $1;

-- name: DeactivateUser :exec
UPDATE users
SET active = false, updated_at = now()
WHERE id = $1;

-- name: ListUsers :many
SELECT id, username, email, first_name, last_name, active, avatar_url, created_at, updated_at
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

