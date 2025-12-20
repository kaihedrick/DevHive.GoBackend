-- Backfill project owners into project_members table
-- This ensures consistency: owners should always exist in project_members
-- This migration is idempotent - safe to run multiple times

-- Insert owners into project_members if they don't already exist
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

-- Log how many owners were backfilled
DO $$
DECLARE
    backfilled_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO backfilled_count
    FROM projects p
    WHERE EXISTS (
        SELECT 1 
        FROM project_members pm 
        WHERE pm.project_id = p.id 
        AND pm.user_id = p.owner_id
        AND pm.role = 'owner'
    );
    
    RAISE NOTICE 'Backfilled % project owners into project_members', backfilled_count;
END $$;

