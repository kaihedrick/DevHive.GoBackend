-- Update NOTIFY cache invalidation function to use camelCase 'projectId' in JSON payload
-- This is a re-application of the logic in 007 to ensure the function is updated
-- with the correct JSON keys (projectId instead of project_id)

-- Create or replace the NOTIFY function
CREATE OR REPLACE FUNCTION notify_cache_invalidation()
RETURNS TRIGGER AS $$
DECLARE
  notification_payload JSONB;
  project_uuid UUID;
  record_id TEXT;
  resource_name TEXT;
BEGIN
  -- Extract project_id based on resource type
  IF TG_TABLE_NAME = 'projects' THEN
    project_uuid := COALESCE(NEW.id, OLD.id);
  ELSIF TG_TABLE_NAME = 'sprints' THEN
    project_uuid := COALESCE(NEW.project_id, OLD.project_id);
  ELSIF TG_TABLE_NAME = 'tasks' THEN
    project_uuid := COALESCE(NEW.project_id, OLD.project_id);
  ELSIF TG_TABLE_NAME = 'messages' THEN
    project_uuid := COALESCE(NEW.project_id, OLD.project_id);
  ELSIF TG_TABLE_NAME = 'project_members' THEN
    project_uuid := COALESCE(NEW.project_id, OLD.project_id);
  ELSE
    -- Unknown table, skip notification
    RETURN COALESCE(NEW, OLD);
  END IF;

  -- Build record ID - for project_members, use composite key since there's no id column
  IF TG_TABLE_NAME = 'project_members' THEN
    record_id := COALESCE(NEW.project_id::text || ':' || NEW.user_id::text, OLD.project_id::text || ':' || OLD.user_id::text);
  ELSE
    record_id := COALESCE(NEW.id::text, OLD.id::text);
  END IF;
  
  -- Normalize resource name to singular for frontend consistency
  -- Frontend expects: 'project', 'sprint', 'task', 'message', 'project_members'
  IF TG_TABLE_NAME = 'projects' THEN
    resource_name := 'project';
  ELSIF TG_TABLE_NAME = 'sprints' THEN
    resource_name := 'sprint';
  ELSIF TG_TABLE_NAME = 'tasks' THEN
    resource_name := 'task';
  ELSIF TG_TABLE_NAME = 'messages' THEN
    resource_name := 'message';
  ELSIF TG_TABLE_NAME = 'project_members' THEN
    resource_name := 'project_members'; -- Keep plural for consistency
  ELSE
    resource_name := TG_TABLE_NAME; -- Fallback to table name
  END IF;
  
  -- Build minimal payload (< 1KB)
  notification_payload := json_build_object(
    'resource', resource_name,
    'id', record_id,
    'action', TG_OP,
    'projectId', project_uuid::text,
    'timestamp', NOW()
  );

  -- Use single channel with payload filtering
  PERFORM pg_notify('cache_invalidate', notification_payload::text);
  
  RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Re-create triggers to ensure they are using the latest function version
-- (Though CREATE OR REPLACE FUNCTION updates the function in place, so triggers using it will automatically use the new version)
-- We'll just ensure they exist.

DROP TRIGGER IF EXISTS projects_cache_invalidate ON projects;
CREATE TRIGGER projects_cache_invalidate
  AFTER INSERT OR UPDATE OR DELETE ON projects
  FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

DROP TRIGGER IF EXISTS sprints_cache_invalidate ON sprints;
CREATE TRIGGER sprints_cache_invalidate
  AFTER INSERT OR UPDATE OR DELETE ON sprints
  FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

DROP TRIGGER IF EXISTS tasks_cache_invalidate ON tasks;
CREATE TRIGGER tasks_cache_invalidate
  AFTER INSERT OR UPDATE OR DELETE ON tasks
  FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

DROP TRIGGER IF EXISTS project_members_cache_invalidate ON project_members;
CREATE TRIGGER project_members_cache_invalidate
  AFTER INSERT OR UPDATE OR DELETE ON project_members
  FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

DROP TRIGGER IF EXISTS messages_cache_invalidate ON messages;
CREATE TRIGGER messages_cache_invalidate
  AFTER INSERT OR UPDATE OR DELETE ON messages
  FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();
