-- Remove title field from tasks table and make description the main field
-- First, update existing records to move title to description if description is empty
UPDATE tasks 
SET description = title 
WHERE description IS NULL OR description = '';

-- Now drop the title column
ALTER TABLE tasks DROP COLUMN title;

