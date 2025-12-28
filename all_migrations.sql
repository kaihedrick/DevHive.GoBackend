-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";

-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL,
    email CITEXT NOT NULL UNIQUE,
    password_h TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create unique indexes for case-insensitive lookups
CREATE UNIQUE INDEX users_username_uidx ON users (lower(username));
CREATE UNIQUE INDEX users_email_uidx ON users (lower(email));

-- Create password_resets table
CREATE TABLE password_resets (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reset_token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_password_resets_reset_token ON password_resets (reset_token);
CREATE INDEX idx_password_resets_user_id ON password_resets (user_id);

-- Create projects table
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_projects_owner_id ON projects (owner_id);

-- Create project_members table (many-to-many relationship)
CREATE TABLE project_members (
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'member', 'viewer')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (project_id, user_id)
);

CREATE INDEX idx_project_members_project_id ON project_members (project_id);
CREATE INDEX idx_project_members_user_id ON project_members (user_id);

-- Create sprints table
CREATE TABLE sprints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    is_completed BOOLEAN NOT NULL DEFAULT false,
    is_started BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sprints_project_id ON sprints (project_id);
CREATE INDEX idx_sprints_dates ON sprints (start_date, end_date);

-- Create tasks table
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sprint_id UUID REFERENCES sprints(id) ON DELETE SET NULL,
    assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    description TEXT,
    status INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tasks_project_id ON tasks (project_id);
CREATE INDEX idx_tasks_sprint_id ON tasks (sprint_id);
CREATE INDEX idx_tasks_assignee_id ON tasks (assignee_id);
CREATE INDEX idx_tasks_status ON tasks (status);

-- Create messages table
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    message_type TEXT NOT NULL DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'file')),
    parent_message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_messages_project_id ON messages (project_id);
CREATE INDEX idx_messages_sender_id ON messages (sender_id);
CREATE INDEX idx_messages_parent_id ON messages (parent_message_id);
CREATE INDEX idx_messages_created_at ON messages (created_at);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sprints_updated_at BEFORE UPDATE ON sprints
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tasks_updated_at BEFORE UPDATE ON tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_messages_updated_at BEFORE UPDATE ON messages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
-- Remove title field from tasks table and make description the main field
-- This migration is idempotent - safe to run multiple times

-- First, check if title column exists and update existing records
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 
    FROM information_schema.columns 
    WHERE table_name = 'tasks' AND column_name = 'title'
  ) THEN
    -- Update existing records to move title to description if description is empty
    UPDATE tasks 
    SET description = title 
    WHERE (description IS NULL OR description = '') AND title IS NOT NULL;
    
    -- Now drop the title column
    ALTER TABLE tasks DROP COLUMN IF EXISTS title;
  END IF;
END $$;

-- Create refresh_tokens table for persistent authentication
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_token ON refresh_tokens (token);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);

-- Cleanup function for expired tokens (can be called periodically)
CREATE OR REPLACE FUNCTION cleanup_expired_refresh_tokens()
RETURNS void AS $$
BEGIN
    DELETE FROM refresh_tokens WHERE expires_at < now();
END;
$$ LANGUAGE plpgsql;






-- Create cache invalidation notification function
-- Uses single channel 'cache_invalidate' with JSON payload containing project_id
CREATE OR REPLACE FUNCTION notify_cache_invalidation()
RETURNS TRIGGER AS $$
DECLARE
  notification_payload JSONB;
  project_uuid UUID;
  record_id TEXT;
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

  -- Build record ID - for project_members, use composite key since there's no id column
  IF TG_TABLE_NAME = 'project_members' THEN
    record_id := COALESCE(NEW.project_id::text || ':' || NEW.user_id::text, OLD.project_id::text || ':' || OLD.user_id::text);
  ELSE
    record_id := COALESCE(NEW.id::text, OLD.id::text);
  END IF;
  
  -- Build minimal payload (< 1KB)
  notification_payload := json_build_object(
    'resource', TG_TABLE_NAME,
    'id', record_id,
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

-- Migration: Add project invites table for invite link system
-- This allows project owners/admins to create time-limited invite links
-- Invites expire after 30 minutes by default and can have optional max uses

-- Create project_invites table
CREATE TABLE project_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invite_token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    max_uses INTEGER DEFAULT NULL, -- NULL = unlimited uses
    used_count INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create indexes for performance
CREATE INDEX idx_project_invites_project_id ON project_invites (project_id);
CREATE INDEX idx_project_invites_invite_token ON project_invites (invite_token);
CREATE INDEX idx_project_invites_expires_at ON project_invites (expires_at);
CREATE INDEX idx_project_invites_active ON project_invites (is_active, expires_at);
CREATE INDEX idx_project_invites_created_by ON project_invites (created_by);

-- Add trigger for updated_at
CREATE TRIGGER update_project_invites_updated_at BEFORE UPDATE ON project_invites
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to automatically clean up expired invites (optional, can be called periodically)
CREATE OR REPLACE FUNCTION cleanup_expired_invites()
RETURNS void AS $$
BEGIN
    UPDATE project_invites 
    SET is_active = false 
    WHERE expires_at < NOW() AND is_active = true;
END;
$$ LANGUAGE plpgsql;






-- Fix notification trigger to handle project_members table correctly
-- project_members doesn't have an id column, so we use composite key (project_id:user_id)
CREATE OR REPLACE FUNCTION notify_cache_invalidation()
RETURNS TRIGGER AS $$
DECLARE
  notification_payload JSONB;
  project_uuid UUID;
  record_id TEXT;
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

  -- Build record ID - for project_members, use composite key since there's no id column
  IF TG_TABLE_NAME = 'project_members' THEN
    record_id := COALESCE(NEW.project_id::text || ':' || NEW.user_id::text, OLD.project_id::text || ':' || OLD.user_id::text);
  ELSE
    record_id := COALESCE(NEW.id::text, OLD.id::text);
  END IF;
  
  -- Build minimal payload (< 1KB)
  notification_payload := json_build_object(
    'resource', TG_TABLE_NAME,
    'id', record_id,
    'action', TG_OP,
    'project_id', project_uuid::text,
    'timestamp', NOW()
  );

  -- Use single channel with payload filtering
  PERFORM pg_notify('cache_invalidate', notification_payload::text);
  
  RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;





-- Ensure NOTIFY cache invalidation function and triggers exist
-- This migration is idempotent and will create/recreate the function and triggers
-- Safe to run multiple times

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

-- Create triggers for projects table (idempotent)
DROP TRIGGER IF EXISTS projects_cache_invalidate ON projects;
CREATE TRIGGER projects_cache_invalidate
  AFTER INSERT OR UPDATE OR DELETE ON projects
  FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

-- Create triggers for sprints table (idempotent)
DROP TRIGGER IF EXISTS sprints_cache_invalidate ON sprints;
CREATE TRIGGER sprints_cache_invalidate
  AFTER INSERT OR UPDATE OR DELETE ON sprints
  FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

-- Create triggers for tasks table (idempotent)
DROP TRIGGER IF EXISTS tasks_cache_invalidate ON tasks;
CREATE TRIGGER tasks_cache_invalidate
  AFTER INSERT OR UPDATE OR DELETE ON tasks
  FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

-- Create triggers for project_members table (idempotent)
DROP TRIGGER IF EXISTS project_members_cache_invalidate ON project_members;
CREATE TRIGGER project_members_cache_invalidate
  AFTER INSERT OR UPDATE OR DELETE ON project_members
  FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

-- Create triggers for messages table (idempotent)
DROP TRIGGER IF EXISTS messages_cache_invalidate ON messages;
CREATE TRIGGER messages_cache_invalidate
  AFTER INSERT OR UPDATE OR DELETE ON messages
  FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

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




-- Migration: Add Google OAuth 2.0 Support
-- This migration adds support for Google OAuth authentication and
-- persistent login (Remember Me) functionality

-- 1. Modify users table to support OAuth authentication
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS auth_provider TEXT DEFAULT 'local'
    CHECK (auth_provider IN ('local', 'google')),
  ADD COLUMN IF NOT EXISTS google_id TEXT UNIQUE,
  ADD COLUMN IF NOT EXISTS profile_picture_url TEXT;

-- Make password_h nullable for OAuth-only users
ALTER TABLE users ALTER COLUMN password_h DROP NOT NULL;

-- Indexes for efficient OAuth lookups
CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id)
  WHERE google_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_auth_provider ON users(auth_provider);

-- Add constraint: users must have at least one authentication method
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'users_auth_method_check'
  ) THEN
    ALTER TABLE users
      ADD CONSTRAINT users_auth_method_check
      CHECK (
        (auth_provider = 'local' AND password_h IS NOT NULL) OR
        (auth_provider = 'google' AND google_id IS NOT NULL)
      );
  END IF;
END $$;

-- Add comments
COMMENT ON COLUMN users.auth_provider IS 'Authentication method: local (username/password) or google (OAuth)';
COMMENT ON COLUMN users.google_id IS 'Google unique user identifier (sub claim from Google)';
COMMENT ON COLUMN users.profile_picture_url IS 'User profile picture URL from Google';

-- 2. Modify refresh_tokens table for Remember Me and Google tokens
ALTER TABLE refresh_tokens
  ADD COLUMN IF NOT EXISTS is_persistent BOOLEAN NOT NULL DEFAULT true,
  ADD COLUMN IF NOT EXISTS google_refresh_token TEXT,
  ADD COLUMN IF NOT EXISTS google_access_token TEXT,
  ADD COLUMN IF NOT EXISTS google_token_expiry TIMESTAMPTZ;

-- Index for cleaning up expired Google tokens
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_google_expiry
  ON refresh_tokens(google_token_expiry)
  WHERE google_token_expiry IS NOT NULL;

-- Add comments
COMMENT ON COLUMN refresh_tokens.is_persistent IS
  'True for persistent login (Remember Me), false for session-only';
COMMENT ON COLUMN refresh_tokens.google_refresh_token IS
  'Google OAuth refresh token for re-authentication';
COMMENT ON COLUMN refresh_tokens.google_access_token IS
  'Google OAuth access token (cached for API calls)';
COMMENT ON COLUMN refresh_tokens.google_token_expiry IS
  'Expiration time for Google access token';

-- 3. Create oauth_state table for CSRF protection
CREATE TABLE IF NOT EXISTS oauth_state (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  state_token TEXT NOT NULL UNIQUE,
  remember_me BOOLEAN NOT NULL DEFAULT false,
  redirect_url TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL DEFAULT (now() + INTERVAL '10 minutes')
);

CREATE INDEX IF NOT EXISTS idx_oauth_state_token ON oauth_state(state_token);
CREATE INDEX IF NOT EXISTS idx_oauth_state_expires ON oauth_state(expires_at);

COMMENT ON TABLE oauth_state IS 'Temporary storage for OAuth state tokens (CSRF protection)';
COMMENT ON COLUMN oauth_state.state_token IS 'Random CSRF token for OAuth flow';
COMMENT ON COLUMN oauth_state.remember_me IS 'User preference for persistent login';
COMMENT ON COLUMN oauth_state.redirect_url IS 'Frontend URL to redirect after successful auth';
COMMENT ON COLUMN oauth_state.expires_at IS 'State tokens expire after 10 minutes';

-- 4. Create cleanup function for expired OAuth state tokens
CREATE OR REPLACE FUNCTION cleanup_expired_oauth_state()
RETURNS void AS $$
BEGIN
  DELETE FROM oauth_state WHERE expires_at < now();
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_expired_oauth_state IS 'Removes expired OAuth state tokens (call periodically or on-demand)';
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
