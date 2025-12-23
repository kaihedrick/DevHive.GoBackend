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





