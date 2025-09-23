-- name: GetMessageByID :one
SELECT m.id, m.project_id, m.sender_id, m.content, m.message_type, m.parent_message_id, m.created_at, m.updated_at,
       u.username as sender_username, u.first_name as sender_first_name, u.last_name as sender_last_name, u.avatar_url as sender_avatar_url
FROM messages m
JOIN users u ON m.sender_id = u.id
WHERE m.id = $1;

-- name: ListMessagesByProject :many
SELECT m.id, m.project_id, m.sender_id, m.content, m.message_type, m.parent_message_id, m.created_at, m.updated_at,
       u.username as sender_username, u.first_name as sender_first_name, u.last_name as sender_last_name, u.avatar_url as sender_avatar_url
FROM messages m
JOIN users u ON m.sender_id = u.id
WHERE m.project_id = $1
ORDER BY m.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListMessagesByProjectAfter :many
SELECT m.id, m.project_id, m.sender_id, m.content, m.message_type, m.parent_message_id, m.created_at, m.updated_at,
       u.username as sender_username, u.first_name as sender_first_name, u.last_name as sender_last_name, u.avatar_url as sender_avatar_url
FROM messages m
JOIN users u ON m.sender_id = u.id
WHERE m.project_id = $1 AND m.id > $2
ORDER BY m.created_at ASC
LIMIT $3;

-- name: CreateMessage :one
INSERT INTO messages (project_id, sender_id, content, message_type, parent_message_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, project_id, sender_id, content, message_type, parent_message_id, created_at, updated_at;

-- name: UpdateMessage :one
UPDATE messages
SET content = $2, updated_at = now()
WHERE id = $1
RETURNING id, project_id, sender_id, content, message_type, parent_message_id, created_at, updated_at;

-- name: DeleteMessage :exec
DELETE FROM messages WHERE id = $1;

