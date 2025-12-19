-- Verify Migration: 004_add_cache_invalidation_triggers.sql
-- Run these queries to verify the migration was applied correctly

-- 1. Check if migration was recorded
SELECT version, applied_at 
FROM schema_migrations 
WHERE version = '004_add_cache_invalidation_triggers.sql';

-- 2. Check if the function exists
SELECT 
    p.proname AS function_name,
    pg_get_functiondef(p.oid) AS function_definition
FROM pg_proc p
JOIN pg_namespace n ON p.pronamespace = n.oid
WHERE n.nspname = 'public' 
AND p.proname = 'notify_cache_invalidation';

-- 3. Check all triggers created by the migration
SELECT 
    c.relname AS table_name,
    t.tgname AS trigger_name,
    CASE 
        WHEN t.tgenabled = 'O' THEN 'enabled'
        ELSE 'disabled'
    END AS status,
    pg_get_triggerdef(t.oid) AS trigger_definition
FROM pg_trigger t
JOIN pg_class c ON t.tgrelid = c.oid
JOIN pg_namespace n ON c.relnamespace = n.oid
WHERE n.nspname = 'public'
AND t.tgname LIKE '%cache_invalidate%'
AND NOT t.tgisinternal
ORDER BY c.relname, t.tgname;

-- 4. Count triggers per table (should be 4 triggers total)
SELECT 
    c.relname AS table_name,
    COUNT(*) AS trigger_count
FROM pg_trigger t
JOIN pg_class c ON t.tgrelid = c.oid
JOIN pg_namespace n ON c.relnamespace = n.oid
WHERE n.nspname = 'public'
AND t.tgname LIKE '%cache_invalidate%'
AND NOT t.tgisinternal
GROUP BY c.relname
ORDER BY c.relname;

-- Expected results:
-- - 1 row in schema_migrations with version '004_add_cache_invalidation_triggers.sql'
-- - 1 function: notify_cache_invalidation
-- - 4 triggers: 
--   * projects_cache_invalidate on projects
--   * sprints_cache_invalidate on sprints
--   * tasks_cache_invalidate on tasks
--   * project_members_cache_invalidate on project_members

