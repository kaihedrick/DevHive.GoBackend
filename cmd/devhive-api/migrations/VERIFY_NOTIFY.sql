-- Manual verification script for NOTIFY triggers
-- Run this directly in your database to check if NOTIFY system is set up correctly

-- 1. Check if function exists
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM pg_proc p
            JOIN pg_namespace n ON p.pronamespace = n.oid
            WHERE n.nspname = 'public' AND p.proname = 'notify_cache_invalidation'
        ) THEN '✅ Function EXISTS'
        ELSE '❌ Function MISSING'
    END as function_status;

-- 2. List all triggers
SELECT 
    c.relname AS table_name,
    t.tgname AS trigger_name,
    CASE 
        WHEN t.tgenabled = 'O' THEN 'enabled'
        ELSE 'disabled'
    END AS status
FROM pg_trigger t
JOIN pg_class c ON t.tgrelid = c.oid
JOIN pg_namespace n ON c.relnamespace = n.oid
WHERE n.nspname = 'public'
AND t.tgname LIKE '%cache_invalidate%'
AND NOT t.tgisinternal
ORDER BY c.relname, t.tgname;

-- 3. Count triggers per table
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

-- Expected: 4 triggers total (projects, sprints, tasks, project_members)



