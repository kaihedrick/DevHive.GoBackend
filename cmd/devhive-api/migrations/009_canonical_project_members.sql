-- Canonical project_members model: Make project_members the single source of truth
-- This migration ensures all project owners are in project_members and adds constraints

-- Step 1: Verify PRIMARY KEY constraint exists (already enforced by schema)
-- The PRIMARY KEY (project_id, user_id) in the initial schema already enforces uniqueness
-- This step is informational only
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint c
        JOIN pg_class t ON c.conrelid = t.oid
        WHERE t.relname = 'project_members'
        AND c.contype = 'p'  -- primary key
    ) THEN
        RAISE NOTICE 'PRIMARY KEY constraint exists on project_members (enforces uniqueness)';
    ELSE
        RAISE WARNING 'PRIMARY KEY constraint missing on project_members - this should not happen!';
    END IF;
END $$;

-- Step 2: Backfill all project owners into project_members (idempotent)
-- Only inserts owners that don't already exist in project_members
INSERT INTO project_members (project_id, user_id, role, joined_at)
SELECT 
    p.id AS project_id,
    p.owner_id AS user_id,
    'owner' AS role,
    p.created_at AS joined_at
FROM projects p
LEFT JOIN project_members pm
    ON pm.project_id = p.id AND pm.user_id = p.owner_id
WHERE pm.user_id IS NULL
ON CONFLICT (project_id, user_id) DO UPDATE 
SET role = 'owner'
WHERE project_members.role != 'owner';

-- Step 3: Log how many owners were backfilled
DO $$
DECLARE
    backfilled_count INTEGER;
    total_projects INTEGER;
    owners_in_members INTEGER;
BEGIN
    SELECT COUNT(*) INTO total_projects FROM projects;
    
    SELECT COUNT(*) INTO owners_in_members
    FROM projects p
    WHERE EXISTS (
        SELECT 1 
        FROM project_members pm 
        WHERE pm.project_id = p.id 
        AND pm.user_id = p.owner_id
        AND pm.role = 'owner'
    );
    
    backfilled_count := owners_in_members;
    
    RAISE NOTICE 'Canonical model migration complete:';
    RAISE NOTICE '  Total projects: %', total_projects;
    RAISE NOTICE '  Owners in project_members: %', owners_in_members;
    RAISE NOTICE '  Coverage: %%%', CASE WHEN total_projects > 0 THEN (owners_in_members * 100 / total_projects) ELSE 0 END;
END $$;

