-- Device authorization grant (RFC 8628) support for CLI/TUI login.
-- api_tokens gains an optional expiry and an origin marker so device-flow
-- tokens can be time-boxed (14 days) and distinguished from manual tokens.
ALTER TABLE api_tokens ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;
ALTER TABLE api_tokens ADD COLUMN IF NOT EXISTS origin VARCHAR(32) NOT NULL DEFAULT 'manual';

CREATE TABLE IF NOT EXISTS device_codes (
  id               BIGSERIAL PRIMARY KEY,
  device_code      TEXT NOT NULL,          -- sha256 hash, like api_tokens.token_hash
  user_code        TEXT NOT NULL,          -- human-readable "PLN-XXXX-XXXX"
  status           TEXT NOT NULL DEFAULT 'pending', -- pending|authorized|denied|expired
  user_id          INTEGER REFERENCES users (id) ON DELETE CASCADE,
  token_id         INTEGER REFERENCES api_tokens (id) ON DELETE SET NULL,
  token_plaintext  TEXT,                    -- set on confirm, returned once on consume, then deleted
  interval_s       INTEGER NOT NULL DEFAULT 5,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at       TIMESTAMPTZ NOT NULL,   -- created_at + 5 min
  last_polled_at   TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_device_codes_device_code ON device_codes (device_code);
CREATE UNIQUE INDEX IF NOT EXISTS idx_device_codes_user_code ON device_codes (user_code);
CREATE INDEX IF NOT EXISTS idx_device_codes_status_expires ON device_codes (status, expires_at);
