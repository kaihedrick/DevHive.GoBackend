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
