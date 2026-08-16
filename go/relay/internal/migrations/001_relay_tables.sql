-- Planetary relay tables.
--
-- The relay tracks a set of DIDs across federated Planetary instances.
-- For each tracked DID it maintains a persistent WebSocket subscription
-- to the DID's PDS repo stream and aggregates io.planetary.* record events
-- into global counts.

-- Known Planetary instances that have announced users to this relay.
CREATE TABLE IF NOT EXISTS instances (
  id          SERIAL PRIMARY KEY,
  url         TEXT NOT NULL UNIQUE,        -- e.g. https://planetary.example.com
  first_seen  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_seen   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Every DID the relay is tracking, keyed by the DID string.
CREATE TABLE IF NOT EXISTS tracked_dids (
  id          BIGSERIAL PRIMARY KEY,
  did         TEXT NOT NULL UNIQUE,
  pds_url     TEXT NOT NULL,               -- PDS base URL for this DID's repo
  handle      TEXT NOT NULL DEFAULT '',    -- latest known Planetary handle
  instance_id INTEGER REFERENCES instances(id) ON DELETE SET NULL,
  -- Subscription state
  cursor_seq  BIGINT NOT NULL DEFAULT 0,   -- last seen repo event sequence
  status      VARCHAR(32) NOT NULL DEFAULT 'active', -- active | error | paused
  error_msg   TEXT NOT NULL DEFAULT '',
  -- Timestamps
  announced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_event_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_tracked_dids_pds_url ON tracked_dids (pds_url);
CREATE INDEX IF NOT EXISTS idx_tracked_dids_status  ON tracked_dids (status);

-- Global follower counts: for each (follower_did, followee_did) pair seen
-- across any tracked PDS repo, we store the follow record key so we can
-- decrement the count cleanly on delete.
CREATE TABLE IF NOT EXISTS observed_follows (
  id           BIGSERIAL PRIMARY KEY,
  follower_did TEXT NOT NULL,
  followee_did TEXT NOT NULL,
  rkey         TEXT NOT NULL,
  pds_url      TEXT NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (follower_did, followee_did, rkey)
);
CREATE INDEX IF NOT EXISTS idx_observed_follows_followee ON observed_follows (followee_did);
CREATE INDEX IF NOT EXISTS idx_observed_follows_follower ON observed_follows (follower_did);

-- Global share observations: one row per (did, rkey) share record seen.
CREATE TABLE IF NOT EXISTS observed_shares (
  id           BIGSERIAL PRIMARY KEY,
  did          TEXT NOT NULL,
  rkey         TEXT NOT NULL,
  article_url  TEXT NOT NULL DEFAULT '',
  feed_url     TEXT NOT NULL DEFAULT '',
  title        TEXT NOT NULL DEFAULT '',
  pds_url      TEXT NOT NULL,
  shared_at    TIMESTAMPTZ,
  observed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (did, rkey)
);
CREATE INDEX IF NOT EXISTS idx_observed_shares_did ON observed_shares (did);

-- Global feed subscription observations.
CREATE TABLE IF NOT EXISTS observed_subscriptions (
  id          BIGSERIAL PRIMARY KEY,
  did         TEXT NOT NULL,
  rkey        TEXT NOT NULL,
  feed_url    TEXT NOT NULL,
  pds_url     TEXT NOT NULL,
  created_at  TIMESTAMPTZ,
  observed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (did, rkey)
);
CREATE INDEX IF NOT EXISTS idx_observed_subs_feed_url ON observed_subscriptions (feed_url);
CREATE INDEX IF NOT EXISTS idx_observed_subs_did      ON observed_subscriptions (did);

-- Outbound event log for subscribeEvents WebSocket fanout to instances.
-- Instances read from this log when they reconnect with a cursor.
CREATE TABLE IF NOT EXISTS relay_events (
  seq         BIGSERIAL PRIMARY KEY,
  event_type  VARCHAR(32) NOT NULL, -- follow|unfollow|share|unshare|feedSubscription|feedUnsubscription
  did         TEXT NOT NULL,
  payload     JSONB NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_relay_events_created_at ON relay_events (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_relay_events_did        ON relay_events (did);
