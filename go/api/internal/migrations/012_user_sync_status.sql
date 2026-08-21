-- Track the post-login PDS backfill so the web UI can show a waiting state.
-- pds_sync_status is "syncing" while the background sync runs, "idle" once it
-- completes, or "failed" if it errored. pds_synced_at is stamped on completion.
ALTER TABLE users ADD COLUMN IF NOT EXISTS pds_sync_status TEXT NOT NULL DEFAULT 'idle';
ALTER TABLE users ADD COLUMN IF NOT EXISTS pds_synced_at TIMESTAMPTZ;
