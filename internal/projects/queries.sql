-- name: GetProjectByID :one
SELECT p.id, p.owner_id, p.name, p.description, p.created_at, p.updated_at,
       u.id as owner_id, u.username as owner_username, u.email as owner_email,
       u.first_name as owner_first_name, u.last_name as owner_last_name
FROM projects p
JOIN users u ON p.owner_id = u.id
WHERE p.id = $1;

-- name: ListProjectsByUser :many
SELECT p.id, p.owner_id, p.name, p.description, p.created_at, p.updated_at,
       u.id as owner_id, u.username as owner_username, u.email as owner_email,
       u.first_name as owner_first_name, u.last_name as owner_last_name
FROM projects p
JOIN users u ON p.owner_id = u.id
LEFT JOIN project_members pm ON p.id = pm.project_id
WHERE p.owner_id = $1 OR pm.user_id = $1
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateProject :one
INSERT INTO projects (owner_id, name, description)
VALUES ($1, $2, $3)
RETURNING id, owner_id, name, description, created_at, updated_at;

-- name: UpdateProject :one
UPDATE projects
SET name = $2, description = $3, updated_at = now()
WHERE id = $1
RETURNING id, owner_id, name, description, created_at, updated_at;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = $1;

-- name: AddProjectMember :exec
INSERT INTO project_members (project_id, user_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (project_id, user_id) DO UPDATE SET role = $3;

-- name: RemoveProjectMember :exec
DELETE FROM project_members WHERE project_id = $1 AND user_id = $2;

-- name: GetProjectMembers :many
SELECT pm.project_id, pm.user_id, pm.role, pm.joined_at,
       u.username, u.email, u.first_name, u.last_name, u.avatar_url
FROM project_members pm
JOIN users u ON pm.user_id = u.id
WHERE pm.project_id = $1
ORDER BY pm.joined_at;

-- name: ListProjectMembers :many
-- Get project owner as a member
SELECT u.id, u.username, u.email, u.first_name, u.last_name, p.created_at as joined_at, 'owner' as role
FROM projects p
JOIN users u ON p.owner_id = u.id
WHERE p.id = $1

UNION ALL

-- Get additional project members
SELECT u.id, u.username, u.email, u.first_name, u.last_name, pm.joined_at, pm.role
FROM project_members pm
JOIN users u ON pm.user_id = u.id
WHERE pm.project_id = $1

ORDER BY joined_at;

-- name: CheckProjectAccess :one
SELECT EXISTS(
    SELECT 1 FROM projects p
    LEFT JOIN project_members pm ON p.id = pm.project_id
    WHERE p.id = $1 AND (p.owner_id = $2 OR pm.user_id = $2)
) as has_access;

