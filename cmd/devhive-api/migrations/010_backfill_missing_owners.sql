-- Backfill owners into project_members for projects missing them
-- This ensures all project owners are in project_members table for consistency
-- Safe to run multiple times (idempotent)

INSERT INTO project_members (project_id, user_id, role, joined_at)
SELECT
    p.id as project_id,
    p.owner_id as user_id,
    'owner' as role,
    p.created_at as joined_at
FROM projects p
WHERE NOT EXISTS (
    SELECT 1
    FROM project_members pm
    WHERE pm.project_id = p.id
    AND pm.user_id = p.owner_id
)
ON CONFLICT (project_id, user_id) DO UPDATE
SET role = 'owner'
WHERE project_members.role != 'owner';



