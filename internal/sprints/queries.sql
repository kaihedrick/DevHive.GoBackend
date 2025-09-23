-- name: GetSprintByID :one
SELECT s.id, s.project_id, s.name, s.description, s.start_date, s.end_date, s.is_completed, s.is_started, s.created_at, s.updated_at,
       p.owner_id, u.username as owner_username, u.email as owner_email, u.first_name as owner_first_name, u.last_name as owner_last_name
FROM sprints s
JOIN projects p ON s.project_id = p.id
JOIN users u ON p.owner_id = u.id
WHERE s.id = $1;

-- name: ListSprintsByProject :many
SELECT s.id, s.project_id, s.name, s.description, s.start_date, s.end_date, s.is_completed, s.is_started, s.created_at, s.updated_at,
       p.owner_id, u.username as owner_username, u.email as owner_email, u.first_name as owner_first_name, u.last_name as owner_last_name
FROM sprints s
JOIN projects p ON s.project_id = p.id
JOIN users u ON p.owner_id = u.id
WHERE s.project_id = $1
ORDER BY s.start_date DESC
LIMIT $2 OFFSET $3;

-- name: CreateSprint :one
INSERT INTO sprints (project_id, name, description, start_date, end_date)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, project_id, name, description, start_date, end_date, is_completed, is_started, created_at, updated_at;

-- name: UpdateSprint :one
UPDATE sprints
SET name = $2, description = $3, start_date = $4, end_date = $5, updated_at = now()
WHERE id = $1
RETURNING id, project_id, name, description, start_date, end_date, is_completed, is_started, created_at, updated_at;

-- name: UpdateSprintStatus :one
UPDATE sprints
SET is_started = $2, is_completed = $3, updated_at = now()
WHERE id = $1
RETURNING id, project_id, name, description, start_date, end_date, is_completed, is_started, created_at, updated_at;

-- name: DeleteSprint :exec
DELETE FROM sprints WHERE id = $1;

