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
INSERT INTO refresh_tokens (user_id, token, expires_at, is_persistent)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, token, expires_at, is_persistent, created_at;

-- name: GetRefreshToken :one
SELECT id, user_id, token, expires_at, is_persistent, created_at
FROM refresh_tokens
WHERE token = $1;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens WHERE token = $1;

-- name: DeleteUserRefreshTokens :exec
DELETE FROM refresh_tokens WHERE user_id = $1;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM refresh_tokens WHERE expires_at < now();

-- OAuth State Queries
-- name: CreateOAuthState :one
INSERT INTO oauth_state (state_token, remember_me, redirect_url, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, state_token, remember_me, redirect_url, created_at, expires_at;

-- name: GetOAuthState :one
SELECT id, state_token, remember_me, redirect_url, created_at, expires_at
FROM oauth_state
WHERE state_token = $1 AND expires_at > now();

-- name: DeleteOAuthState :exec
DELETE FROM oauth_state WHERE state_token = $1;

-- name: DeleteExpiredOAuthStates :exec
DELETE FROM oauth_state WHERE expires_at < now();

-- OAuth Refresh Token Queries
-- name: CreateRefreshTokenWithGoogle :one
INSERT INTO refresh_tokens (user_id, token, expires_at, is_persistent, google_refresh_token, google_access_token, google_token_expiry)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, user_id, token, expires_at, is_persistent, google_refresh_token, google_access_token, google_token_expiry, created_at;

-- name: UpdateRefreshTokenGoogleTokens :exec
UPDATE refresh_tokens
SET google_access_token = $2, google_token_expiry = $3
WHERE token = $1;

