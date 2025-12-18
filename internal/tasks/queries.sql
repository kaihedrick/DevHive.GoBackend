-- name: GetTaskByID :one
SELECT t.id, t.project_id, t.sprint_id, t.assignee_id, t.description, t.status, t.created_at, t.updated_at,
       u.username as assignee_username, u.first_name as assignee_first_name, u.last_name as assignee_last_name,
       p.owner_id, owner.username as owner_username, owner.email as owner_email, owner.first_name as owner_first_name, owner.last_name as owner_last_name
FROM tasks t
LEFT JOIN users u ON t.assignee_id = u.id
JOIN projects p ON t.project_id = p.id
JOIN users owner ON p.owner_id = owner.id
WHERE t.id = $1;

-- name: ListTasksByProject :many
SELECT t.id, t.project_id, t.sprint_id, t.assignee_id, t.description, t.status, t.created_at, t.updated_at,
       u.username as assignee_username, u.first_name as assignee_first_name, u.last_name as assignee_last_name,
       p.owner_id, owner.username as owner_username, owner.email as owner_email, owner.first_name as owner_first_name, owner.last_name as owner_last_name
FROM tasks t
LEFT JOIN users u ON t.assignee_id = u.id
JOIN projects p ON t.project_id = p.id
JOIN users owner ON p.owner_id = owner.id
WHERE t.project_id = $1
ORDER BY t.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListTasksBySprint :many
SELECT t.id, t.project_id, t.sprint_id, t.assignee_id, t.description, t.status, t.created_at, t.updated_at,
       u.username as assignee_username, u.first_name as assignee_first_name, u.last_name as assignee_last_name,
       p.owner_id, owner.username as owner_username, owner.email as owner_email, owner.first_name as owner_first_name, owner.last_name as owner_last_name
FROM tasks t
LEFT JOIN users u ON t.assignee_id = u.id
JOIN projects p ON t.project_id = p.id
JOIN users owner ON p.owner_id = owner.id
WHERE t.sprint_id = $1
ORDER BY t.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateTask :one
INSERT INTO tasks (project_id, sprint_id, assignee_id, description, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, project_id, sprint_id, assignee_id, description, status, created_at, updated_at;

-- name: UpdateTask :one
UPDATE tasks
SET description = $2, assignee_id = $3, updated_at = now()
WHERE id = $1
RETURNING id, project_id, sprint_id, assignee_id, description, status, created_at, updated_at;

-- name: UpdateTaskStatus :one
UPDATE tasks
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING id, project_id, sprint_id, assignee_id, description, status, created_at, updated_at;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = $1;

