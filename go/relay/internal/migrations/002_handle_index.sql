-- Add index on tracked_dids.handle for cross-instance user search.
CREATE INDEX IF NOT EXISTS idx_tracked_dids_handle ON tracked_dids (handle);
