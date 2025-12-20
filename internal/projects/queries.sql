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
-- Get project owner as a member (for consistency with ListProjectMembers)
-- Returns same structure as ListProjectMembers but includes avatar_url
SELECT $1::uuid as project_id, u.id as user_id, u.username, u.email, u.first_name, u.last_name, u.avatar_url, p.created_at as joined_at, 'owner' as role
FROM projects p
JOIN users u ON p.owner_id = u.id
WHERE p.id = $1

UNION ALL

-- Get additional project members
SELECT pm.project_id, u.id as user_id, u.username, u.email, u.first_name, u.last_name, u.avatar_url, pm.joined_at, pm.role
FROM project_members pm
JOIN users u ON pm.user_id = u.id
WHERE pm.project_id = $1

ORDER BY joined_at;

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
-- Check if user is project owner OR project member
SELECT (
    EXISTS(SELECT 1 FROM projects p WHERE p.id = $1 AND p.owner_id = $2) OR
    EXISTS(SELECT 1 FROM project_members pm WHERE pm.project_id = $1 AND pm.user_id = $2)
)::boolean as has_access;

-- name: CheckProjectOwner :one
SELECT EXISTS(
    SELECT 1 FROM projects p
    WHERE p.id = $1 AND p.owner_id = $2
) as is_owner;

-- name: ProjectExists :one
SELECT EXISTS(
    SELECT 1 FROM projects p
    WHERE p.id = $1
) as exists;

-- name: CheckProjectOwnerOrAdmin :one
-- Check if user is project owner OR has admin role in project_members
SELECT (
    EXISTS(SELECT 1 FROM projects p WHERE p.id = $1 AND p.owner_id = $2) OR
    EXISTS(SELECT 1 FROM project_members pm WHERE pm.project_id = $1 AND pm.user_id = $2 AND pm.role = 'admin')
)::boolean as is_owner_or_admin;

-- name: GetUserProjectRole :one
-- Get the user's role in a project (returns 'owner', 'admin', 'member', 'viewer', or NULL if not a member)
SELECT 
    CASE 
        WHEN p.owner_id = $2 THEN 'owner'
        ELSE COALESCE(pm.role, NULL)
    END as role
FROM projects p
LEFT JOIN project_members pm ON p.id = pm.project_id AND pm.user_id = $2
WHERE p.id = $1;

-- name: CreateProjectInvite :one
INSERT INTO project_invites (project_id, created_by, invite_token, expires_at, max_uses)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, project_id, created_by, invite_token, expires_at, max_uses, used_count, is_active, created_at, updated_at;

-- name: GetProjectInviteByToken :one
SELECT id, project_id, created_by, invite_token, expires_at, max_uses, used_count, is_active, created_at, updated_at
FROM project_invites
WHERE invite_token = $1 AND is_active = true;

-- name: GetProjectInviteByID :one
SELECT id, project_id, created_by, invite_token, expires_at, max_uses, used_count, is_active, created_at, updated_at
FROM project_invites
WHERE id = $1;

-- name: IncrementInviteUseCount :exec
UPDATE project_invites
SET used_count = used_count + 1, updated_at = now()
WHERE id = $1;

-- name: DeactivateInvite :exec
UPDATE project_invites
SET is_active = false, updated_at = now()
WHERE id = $1;

-- name: ListProjectInvites :many
SELECT id, project_id, created_by, invite_token, expires_at, max_uses, used_count, is_active, created_at, updated_at
FROM project_invites
WHERE project_id = $1 AND is_active = true
ORDER BY created_at DESC;

