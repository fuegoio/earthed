-- Rename "categories" to "folders" and add nesting + sort order support.

-- Rename the categories table to folders.
ALTER TABLE categories RENAME TO folders;

-- Add parent_id for nestable folders (self-reference, ON DELETE SET NULL).
ALTER TABLE folders ADD COLUMN parent_id INTEGER REFERENCES folders (id) ON DELETE SET NULL;
ALTER TABLE folders ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;

-- Rename feeds.category_id to feeds.folder_id and update the FK to reference folders.
ALTER TABLE feeds RENAME COLUMN category_id TO folder_id;
ALTER TABLE feeds DROP CONSTRAINT IF EXISTS feeds_category_id_fkey;
ALTER TABLE feeds ADD CONSTRAINT feeds_folder_id_fkey
  FOREIGN KEY (folder_id) REFERENCES folders (id) ON DELETE SET NULL;

-- Add sort_order to feeds for drag-and-drop ordering within a folder.
ALTER TABLE feeds ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;

-- Update indexes to match new names.
DROP INDEX IF EXISTS idx_categories_user_id;
CREATE INDEX IF NOT EXISTS idx_folders_user_id ON folders (user_id);
CREATE INDEX IF NOT EXISTS idx_folders_parent_id ON folders (parent_id);
CREATE INDEX IF NOT EXISTS idx_feeds_folder_id ON feeds (folder_id);
