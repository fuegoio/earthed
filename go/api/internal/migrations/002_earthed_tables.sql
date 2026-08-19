-- Earthed RSS reader tables: categories, feeds, entries, enclosures, api_tokens, feed_icons

CREATE TABLE IF NOT EXISTS api_tokens (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  label VARCHAR(255) NOT NULL DEFAULT '',
  token_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_used_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_tokens_token_hash ON api_tokens (token_hash);
CREATE INDEX IF NOT EXISTS idx_api_tokens_user_id ON api_tokens (user_id);

CREATE TABLE IF NOT EXISTS categories (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  title VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories (user_id);

CREATE TABLE IF NOT EXISTS feeds (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  category_id INTEGER REFERENCES categories (id) ON DELETE SET NULL,
  feed_url TEXT NOT NULL,
  site_url TEXT NOT NULL DEFAULT '',
  title VARCHAR(512) NOT NULL DEFAULT '',
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
CREATE INDEX IF NOT EXISTS idx_feeds_user_id ON feeds (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_feeds_user_feed_url ON feeds (user_id, feed_url);
CREATE INDEX IF NOT EXISTS idx_feeds_next_check_at ON feeds (next_check_at) WHERE disabled = false;

CREATE TABLE IF NOT EXISTS entries (
  id BIGSERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  feed_id INTEGER NOT NULL REFERENCES feeds (id) ON DELETE CASCADE,
  hash VARCHAR(255) NOT NULL,
  title TEXT NOT NULL DEFAULT '',
  url TEXT NOT NULL DEFAULT '',
  comments_url TEXT NOT NULL DEFAULT '',
  author VARCHAR(255) NOT NULL DEFAULT '',
  content TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  status VARCHAR(50) NOT NULL DEFAULT 'unread',
  starred BOOLEAN NOT NULL DEFAULT false,
  published_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  tags TEXT[] NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  document tsvector GENERATED ALWAYS AS (to_tsvector('english', coalesce(title, '') || ' ' || coalesce(content, ''))) STORED
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_entries_feed_hash ON entries (feed_id, hash);
CREATE INDEX IF NOT EXISTS idx_entries_user_id ON entries (user_id);
CREATE INDEX IF NOT EXISTS idx_entries_feed_id ON entries (feed_id);
CREATE INDEX IF NOT EXISTS idx_entries_status ON entries (user_id, status);
CREATE INDEX IF NOT EXISTS idx_entries_starred ON entries (user_id, starred);
CREATE INDEX IF NOT EXISTS idx_entries_published_at ON entries (user_id, published_at DESC);
CREATE INDEX IF NOT EXISTS idx_entries_document ON entries USING gin (document);

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
