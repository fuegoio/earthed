-- Social features: user handles, follows, and article sharing.

-- User-controlled handle (like @fuego). Unique, changeable.
-- Will correspond to AT Proto handle in the future.
CREATE TABLE IF NOT EXISTS user_profiles (
  user_id   INTEGER PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
  handle    VARCHAR(64) NOT NULL,
  bio       TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_profiles_handle ON user_profiles (handle);

-- Follow graph: follower follows followee.
CREATE TABLE IF NOT EXISTS user_follows (
  id          BIGSERIAL PRIMARY KEY,
  follower_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  followee_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (follower_id, followee_id)
);
CREATE INDEX IF NOT EXISTS idx_user_follows_follower ON user_follows (follower_id);
CREATE INDEX IF NOT EXISTS idx_user_follows_followee ON user_follows (followee_id);

-- Shared articles: a user shares an article (by URL, with cached metadata).
-- We store the original entry metadata so the card is renderable without
-- the sharer needing to subscribe to the same feed.
CREATE TABLE IF NOT EXISTS shared_articles (
  id               BIGSERIAL PRIMARY KEY,
  user_id          INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  article_url      TEXT NOT NULL,
  title            TEXT NOT NULL DEFAULT '',
  description      TEXT NOT NULL DEFAULT '',
  feed_url         TEXT NOT NULL DEFAULT '',
  feed_title       VARCHAR(512) NOT NULL DEFAULT '',
  feed_site_url    TEXT NOT NULL DEFAULT '',
  author           VARCHAR(255) NOT NULL DEFAULT '',
  published_at     TIMESTAMPTZ,
  shared_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, article_url)
);
CREATE INDEX IF NOT EXISTS idx_shared_articles_user_id    ON shared_articles (user_id);
CREATE INDEX IF NOT EXISTS idx_shared_articles_shared_at  ON shared_articles (shared_at DESC);
