-- Simplify user model: consolidate user_profiles into users, add is_remote.
--
-- The local database is a cache layer for ATProto data. Profile fields
-- (display_name, bio) are cached from the PDS's app.bsky.actor.profile record.
-- PDS credentials (pds_url, tokens) are set at OAuth time.
--
-- BREAKING: drops user_profiles table and vestigial users columns
-- (email, first_name, last_name, updated_at).

-- Add profile + PDS credential columns to users.
ALTER TABLE users ADD COLUMN IF NOT EXISTS display_name TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS bio TEXT NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_remote BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS pds_url TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS atproto_access_token TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS atproto_refresh_token TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS atproto_token_expires_at TIMESTAMPTZ;

-- Migrate data from user_profiles to users.
UPDATE users u SET
    bio = COALESCE(up.bio, ''),
    pds_url = up.pds_url,
    atproto_access_token = up.atproto_access_token,
    atproto_refresh_token = up.atproto_refresh_token,
    atproto_token_expires_at = up.atproto_token_expires_at
FROM user_profiles up
WHERE up.user_id = u.id;

-- Copy DID from user_profiles if missing on users.
UPDATE users u SET did = up.did
FROM user_profiles up
WHERE up.user_id = u.id AND u.did IS NULL AND up.did IS NOT NULL;

-- first_name becomes display_name (the de facto display name until PDS profile is fetched).
UPDATE users SET display_name = COALESCE(first_name, '') WHERE first_name IS NOT NULL AND first_name <> '';

-- Drop the user_profiles table.
DROP TABLE IF EXISTS user_profiles;

-- Drop vestigial columns.
ALTER TABLE users DROP COLUMN IF EXISTS email;
ALTER TABLE users DROP COLUMN IF EXISTS first_name;
ALTER TABLE users DROP COLUMN IF EXISTS last_name;
ALTER TABLE users DROP COLUMN IF EXISTS updated_at;

-- Add unique index on handle.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_handle ON users (handle) WHERE handle IS NOT NULL;
