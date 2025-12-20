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

