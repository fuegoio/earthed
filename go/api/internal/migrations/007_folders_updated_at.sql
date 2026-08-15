-- Add updated_at column to folders (categories table never had it).
ALTER TABLE folders ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
