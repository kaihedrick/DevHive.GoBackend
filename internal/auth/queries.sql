-- name: CreatePasswordReset :one
INSERT INTO password_resets (user_id, reset_token, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, reset_token, expires_at, created_at;

-- name: GetPasswordResetByToken :one
SELECT pr.id, pr.user_id, pr.reset_token, pr.expires_at, pr.created_at,
       u.username, u.email, u.first_name, u.last_name
FROM password_resets pr
JOIN users u ON pr.user_id = u.id
WHERE pr.reset_token = $1;

-- name: DeletePasswordReset :exec
DELETE FROM password_resets WHERE reset_token = $1;

-- name: DeleteExpiredPasswordResets :exec
DELETE FROM password_resets WHERE expires_at < now();

-- Refresh Token Queries
-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, token, expires_at, created_at;

-- name: GetRefreshToken :one
SELECT id, user_id, token, expires_at, created_at
FROM refresh_tokens
WHERE token = $1;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens WHERE token = $1;

-- name: DeleteUserRefreshTokens :exec
DELETE FROM refresh_tokens WHERE user_id = $1;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens WHERE expires_at < now();

