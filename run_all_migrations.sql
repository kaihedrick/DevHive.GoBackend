-- DevHive Database Migrations - Run All
-- Run this script in your Neon SQL Editor to apply all migrations
-- This script is idempotent and can be run multiple times safely

-- Create migrations tracking table
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================================
-- MIGRATION 001: Initial Schema
-- ============================================================

