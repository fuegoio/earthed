-- Add a source column to feeds so the API can distinguish how a feed is
-- fetched: 'rss' (the default, fetched and parsed as RSS/Atom/JSON) or 'x'
-- (a person's X timeline, fetched from the official X API v2).
ALTER TABLE feeds ADD COLUMN IF NOT EXISTS source VARCHAR(20) NOT NULL DEFAULT 'rss';

-- Add an entry_type column to entries so consumers can render X posts
-- differently from RSS articles.
ALTER TABLE entries ADD COLUMN IF NOT EXISTS entry_type VARCHAR(20) NOT NULL DEFAULT 'article';
