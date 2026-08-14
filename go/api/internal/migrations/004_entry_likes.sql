-- Add liked column to entries (per-user like, mirroring the starred pattern)
ALTER TABLE entries ADD COLUMN IF NOT EXISTS liked BOOLEAN NOT NULL DEFAULT false;
CREATE INDEX IF NOT EXISTS idx_entries_liked ON entries (user_id, liked);
