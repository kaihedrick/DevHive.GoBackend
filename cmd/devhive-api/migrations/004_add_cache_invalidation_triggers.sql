-- Create cache invalidation notification function
-- Uses single channel 'cache_invalidate' with JSON payload containing project_id
CREATE OR REPLACE FUNCTION notify_cache_invalidation()
RETURNS TRIGGER AS $$
DECLARE
  notification_payload JSONB;
  project_uuid UUID;
BEGIN
  -- Extract project_id based on resource type
  IF TG_TABLE_NAME = 'projects' THEN
    project_uuid := COALESCE(NEW.id, OLD.id);
  ELSIF TG_TABLE_NAME = 'sprints' THEN
    project_uuid := COALESCE(NEW.project_id, OLD.project_id);
  ELSIF TG_TABLE_NAME = 'tasks' THEN
    project_uuid := COALESCE(NEW.project_id, OLD.project_id);
  ELSIF TG_TABLE_NAME = 'project_members' THEN
    project_uuid := COALESCE(NEW.project_id, OLD.project_id);
  ELSE
    -- Unknown table, skip notification
    RETURN COALESCE(NEW, OLD);
  END IF;

  -- Build minimal payload (< 1KB)
  notification_payload := json_build_object(
    'resource', TG_TABLE_NAME,
    'id', COALESCE(NEW.id::text, OLD.id::text),
    'action', TG_OP,
    'project_id', project_uuid::text,
    'timestamp', NOW()
  );

  -- Use single channel with payload filtering
  PERFORM pg_notify('cache_invalidate', notification_payload::text);
  
  RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Create triggers for projects table
DROP TRIGGER IF EXISTS projects_cache_invalidate ON projects;
CREATE TRIGGER projects_cache_invalidate
  AFTER INSERT OR UPDATE OR DELETE ON projects
  FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

-- Create triggers for sprints table
DROP TRIGGER IF EXISTS sprints_cache_invalidate ON sprints;
CREATE TRIGGER sprints_cache_invalidate
  AFTER INSERT OR UPDATE OR DELETE ON sprints
  FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

-- Create triggers for tasks table
DROP TRIGGER IF EXISTS tasks_cache_invalidate ON tasks;
CREATE TRIGGER tasks_cache_invalidate
  AFTER INSERT OR UPDATE OR DELETE ON tasks
  FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

-- Create triggers for project_members table
DROP TRIGGER IF EXISTS project_members_cache_invalidate ON project_members;
CREATE TRIGGER project_members_cache_invalidate
  AFTER INSERT OR UPDATE OR DELETE ON project_members
  FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

