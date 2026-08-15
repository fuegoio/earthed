-- Add description column to feeds table.
ALTER TABLE feeds ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
