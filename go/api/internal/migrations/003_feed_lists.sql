-- Feed lists: curated, shareable collections of feeds.

CREATE TABLE IF NOT EXISTS feed_lists (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  title VARCHAR(255) NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  is_public BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_feed_lists_user_id ON feed_lists (user_id);
CREATE INDEX IF NOT EXISTS idx_feed_lists_public ON feed_lists (is_public) WHERE is_public = true;

CREATE TABLE IF NOT EXISTS feed_list_feeds (
  id SERIAL PRIMARY KEY,
  feed_list_id INTEGER NOT NULL REFERENCES feed_lists (id) ON DELETE CASCADE,
  feed_url TEXT NOT NULL,
  site_url TEXT NOT NULL DEFAULT '',
  title VARCHAR(512) NOT NULL DEFAULT '',
  position INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_feed_list_feeds_list_url ON feed_list_feeds (feed_list_id, feed_url);
CREATE INDEX IF NOT EXISTS idx_feed_list_feeds_list_id ON feed_list_feeds (feed_list_id);

CREATE TABLE IF NOT EXISTS feed_list_follows (
  id SERIAL PRIMARY KEY,
  feed_list_id INTEGER NOT NULL REFERENCES feed_lists (id) ON DELETE CASCADE,
  user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_feed_list_follows ON feed_list_follows (feed_list_id, user_id);
CREATE INDEX IF NOT EXISTS idx_feed_list_follows_user ON feed_list_follows (user_id);
