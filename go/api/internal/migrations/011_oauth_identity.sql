-- AT Proto OAuth identity: replace email/password with DID-based identity.
--
-- BREAKING: this migration drops the Limen password columns and makes the DID
-- the user identity. Email/password auth is removed; users log in via AT Proto
-- OAuth with their PDS. There is no production data to preserve.
--
-- Three new tables back the OAuth client:
--   oauth_auth_requests — in-flight auth flows (state → AuthRequestData JSON)
--   oauth_sessions      — persisted OAuth sessions (did, session_id → JSON)
--   web_sessions        — Sunred's own session cookie → user_id mapping

-- The DID is now the identity. Existing email/password users are not migrated;
-- the columns are dropped. New users are created with a DID at OAuth callback.
ALTER TABLE users ADD COLUMN IF NOT EXISTS did TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS handle TEXT;
ALTER TABLE users DROP COLUMN IF EXISTS password;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at;
-- email becomes optional (may be provided by the PDS scope later)
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;
DROP INDEX IF EXISTS idx_users_email;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_did ON users (did) WHERE did IS NOT NULL;

-- user_profiles.handle is now sourced from users.handle (set at OAuth login).
-- Keep the profile table for bio + denormalised counts; handle moves to users.
ALTER TABLE user_profiles ALTER COLUMN handle DROP NOT NULL;
DROP INDEX IF EXISTS idx_user_profiles_handle;

-- Persisted OAuth auth-request state (the indigo ClientAuthStore contract).
CREATE TABLE IF NOT EXISTS oauth_auth_requests (
  state      TEXT PRIMARY KEY,
  data       JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_oauth_auth_requests_created_at ON oauth_auth_requests (created_at);

-- Persisted OAuth sessions (indigo ClientSessionData), keyed by DID + sessionID.
CREATE TABLE IF NOT EXISTS oauth_sessions (
  account_did  TEXT NOT NULL,
  session_id   TEXT NOT NULL,
  data         JSONB NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (account_did, session_id)
);

-- Sunred web session cookie: opaque token → user_id. Issued after a successful
-- OAuth callback so the browser authenticates to the API without round-tripping
-- the OAuth session on every request.
CREATE TABLE IF NOT EXISTS web_sessions (
  token      TEXT PRIMARY KEY,
  user_id    INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_web_sessions_user_id ON web_sessions (user_id);

-- Clean up the old Limen tables we no longer use.
DROP TABLE IF EXISTS verifications;
DROP TABLE IF EXISTS rate_limits;
-- accounts table was unused (Limen OAuth never wired); repurpose is unnecessary.
DROP TABLE IF EXISTS accounts;
