-- Global feeds + per-user subscriptions; global entries + per-user entry state.
--
-- BREAKING: feeds are no longer user-owned. A feed row is global (one per
-- feed_url) and carries shared fetch state. A user's subscription is a row in
-- `subscriptions` linking them to a global feed (with folder + title override).
-- Entries are global (one per feed+hash); read/starred/liked state is per-user
-- in `entry_state`. Shared articles now materialize as global entries so that
-- articles shared by followed users appear in the entry stream via a UNION at
-- query time, even when the viewer does not subscribe to the source feed.
--
-- Not in production; existing data is dropped rather than migrated.

-- Drop the per-user feeds/entries model. CASCADE removes enclosures and
-- feed_icons (FK'd to feeds/entries). shared_articles survives; we add an
-- entry link below.
DROP TABLE IF EXISTS entries CASCADE;
DROP TABLE IF EXISTS feeds CASCADE;

-- Global feeds: one row per feed_url, shared fetch state.
CREATE TABLE IF NOT EXISTS feeds (
  id SERIAL PRIMARY KEY,
  feed_url TEXT NOT NULL,
  site_url TEXT NOT NULL DEFAULT '',
  title VARCHAR(512) NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  etag_header TEXT NOT NULL DEFAULT '',
  last_modified_header TEXT NOT NULL DEFAULT '',
  parsing_error TEXT NOT NULL DEFAULT '',
  parsing_error_count INTEGER NOT NULL DEFAULT 0,
  disabled BOOLEAN NOT NULL DEFAULT false,
  scraper_rules TEXT NOT NULL DEFAULT '',
  rewrite_rules TEXT NOT NULL DEFAULT '',
  crawler BOOLEAN NOT NULL DEFAULT false,
  next_check_at TIMESTAMPTZ,
  last_fetch_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_feeds_feed_url ON feeds (feed_url);
CREATE INDEX IF NOT EXISTS idx_feeds_next_check_at ON feeds (next_check_at) WHERE disabled = false;

-- Per-user subscriptions: links a user to a global feed, with folder + title
-- override + the AT Proto record key for the subscription (was feeds.atproto_rkey).
CREATE TABLE IF NOT EXISTS subscriptions (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  feed_id INTEGER NOT NULL REFERENCES feeds (id) ON DELETE CASCADE,
  folder_id INTEGER REFERENCES folders (id) ON DELETE SET NULL,
  title_override VARCHAR(512),
  sort_order INTEGER NOT NULL DEFAULT 0,
  atproto_rkey TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, feed_id)
);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions (user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_feed_id ON subscriptions (feed_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_folder_id ON subscriptions (folder_id);

-- Global entries: one row per feed + hash. No user_id, no read/starred state.
CREATE TABLE IF NOT EXISTS entries (
  id BIGSERIAL PRIMARY KEY,
  feed_id INTEGER NOT NULL REFERENCES feeds (id) ON DELETE CASCADE,
  hash VARCHAR(255) NOT NULL,
  title TEXT NOT NULL DEFAULT '',
  url TEXT NOT NULL DEFAULT '',
  comments_url TEXT NOT NULL DEFAULT '',
  author VARCHAR(255) NOT NULL DEFAULT '',
  content TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  published_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  tags TEXT[] NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  document tsvector GENERATED ALWAYS AS (to_tsvector('english', coalesce(title, '') || ' ' || coalesce(content, ''))) STORED
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_entries_feed_hash ON entries (feed_id, hash);
CREATE INDEX IF NOT EXISTS idx_entries_feed_id ON entries (feed_id);
CREATE INDEX IF NOT EXISTS idx_entries_published_at ON entries (published_at DESC);
CREATE INDEX IF NOT EXISTS idx_entries_document ON entries USING gin (document);

-- Per-user entry state: read/starred/liked live here. Absence of a row means
-- the entry is unread, unstarred, unliked for that user.
CREATE TABLE IF NOT EXISTS entry_state (
  user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  entry_id BIGINT NOT NULL REFERENCES entries (id) ON DELETE CASCADE,
  status VARCHAR(50) NOT NULL DEFAULT 'unread',
  starred BOOLEAN NOT NULL DEFAULT false,
  liked BOOLEAN NOT NULL DEFAULT false,
  changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, entry_id)
);
CREATE INDEX IF NOT EXISTS idx_entry_state_user_status ON entry_state (user_id, status);
CREATE INDEX IF NOT EXISTS idx_entry_state_user_starred ON entry_state (user_id, starred);

-- Enclosures + feed icons, re-created against the new global tables.
CREATE TABLE IF NOT EXISTS enclosures (
  id BIGSERIAL PRIMARY KEY,
  entry_id BIGINT NOT NULL REFERENCES entries (id) ON DELETE CASCADE,
  url TEXT NOT NULL,
  mime_type VARCHAR(255) NOT NULL DEFAULT '',
  size BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_enclosures_entry_id ON enclosures (entry_id);

CREATE TABLE IF NOT EXISTS feed_icons (
  feed_id INTEGER PRIMARY KEY REFERENCES feeds (id) ON DELETE CASCADE,
  data BYTEA NOT NULL
);

-- Shared articles link to the global entry they describe, so the entry
-- stream can UNION in articles shared by followed users.
ALTER TABLE shared_articles ADD COLUMN IF NOT EXISTS entry_id BIGINT REFERENCES entries (id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_shared_articles_entry_id ON shared_articles (entry_id);
