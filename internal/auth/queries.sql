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

