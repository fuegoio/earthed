-- Persist the OAuth session ID per user so the API can resume the indigo
-- OAuth session (DPoP-bound) to write io.sunred.* records to the user's PDS.
-- The session data itself lives in oauth_sessions (keyed by did + session_id);
-- this column records which session is the user's active one.
-- Also ensures pds_url is set at OAuth time (the resource/PDS host).
ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_session_id TEXT;
