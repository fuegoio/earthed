// Package store wraps *sql.DB with domain-specific query helpers for the
// Planetary RSS reader schema.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// ErrFeedNotFound is returned when a feed lookup returns no row.
var ErrFeedNotFound = fmt.Errorf("feed not found")

// ErrEntryNotFound is returned when an entry lookup returns no row.
var ErrEntryNotFound = fmt.Errorf("entry not found")

// ErrFeedListNotFound is returned when a feed list lookup returns no row.
var ErrFeedListNotFound = fmt.Errorf("feed list not found")

// ErrFeedListOwnList is returned when a user tries to follow their own list.
var ErrFeedListOwnList = fmt.Errorf("cannot follow your own list")

// ErrFeedListNotPublic is returned when a user tries to follow a private list.
var ErrFeedListNotPublic = fmt.Errorf("feed list is not public")

// --- Categories ---

// CreateCategory inserts a new category for the given user.
func (s *Store) CreateCategory(ctx context.Context, userID int, title string) (*Category, error) {
	var c Category
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO categories (user_id, title) VALUES ($1, $2)
		 RETURNING id, user_id, title, created_at`,
		userID, title,
	).Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}
	return &c, nil
}

// ListCategories returns all categories for the given user.
func (s *Store) ListCategories(ctx context.Context, userID int) ([]Category, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, user_id, title, created_at FROM categories WHERE user_id = $1 ORDER BY title`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var cats []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

// GetCategoryByID returns the category with the given id scoped to userID, or nil.
func (s *Store) GetCategoryByID(ctx context.Context, id, userID int) (*Category, error) {
	var c Category
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, user_id, title, created_at FROM categories WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateCategory sets the title of the given category.
func (s *Store) UpdateCategory(ctx context.Context, id, userID int, title string) (*Category, error) {
	var c Category
	err := s.DB.QueryRowContext(ctx,
		`UPDATE categories SET title = $3 WHERE id = $1 AND user_id = $2
		 RETURNING id, user_id, title, created_at`,
		id, userID, title,
	).Scan(&c.ID, &c.UserID, &c.Title, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update category: %w", err)
	}
	return &c, nil
}

// DeleteCategory removes a category. Feeds in the category are re-assigned to
// NULL category_id via ON DELETE SET NULL.
func (s *Store) DeleteCategory(ctx context.Context, id, userID int) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM categories WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

// --- Feeds ---

// CreateFeed inserts a new feed subscription for the given user.
func (s *Store) CreateFeed(ctx context.Context, userID int, categoryID *int, feedURL, siteURL, title string) (*Feed, error) {
	var f Feed
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO feeds (user_id, category_id, feed_url, site_url, title)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, category_id, feed_url, site_url, title,
		           etag_header, last_modified_header, parsing_error, parsing_error_count,
		           disabled, scraper_rules, rewrite_rules, crawler,
		           next_check_at, last_fetch_at, created_at, updated_at`,
		userID, categoryID, feedURL, siteURL, title,
	).Scan(&f.ID, &f.UserID, &f.CategoryID, &f.FeedURL, &f.SiteURL, &f.Title,
		&f.EtagHeader, &f.LastModified, &f.ParsingError, &f.ParsingErrorCount,
		&f.Disabled, &f.ScraperRules, &f.RewriteRules, &f.Crawler,
		&f.NextCheckAt, &f.LastFetchAt, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create feed: %w", err)
	}
	return &f, nil
}

// ListFeeds returns all feeds for the given user.
func (s *Store) ListFeeds(ctx context.Context, userID int) ([]Feed, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, user_id, category_id, feed_url, site_url, title,
		        etag_header, last_modified_header, parsing_error, parsing_error_count,
		        disabled, scraper_rules, rewrite_rules, crawler,
		        next_check_at, last_fetch_at, created_at, updated_at
		 FROM feeds WHERE user_id = $1 ORDER BY title`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var feeds []Feed
	for rows.Next() {
		var f Feed
		if err := scanFeed(rows, &f); err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}
	return feeds, rows.Err()
}

// GetFeedByID returns the feed with the given id scoped to userID, or nil.
func (s *Store) GetFeedByID(ctx context.Context, id, userID int) (*Feed, error) {
	var f Feed
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, user_id, category_id, feed_url, site_url, title,
		        etag_header, last_modified_header, parsing_error, parsing_error_count,
		        disabled, scraper_rules, rewrite_rules, crawler,
		        next_check_at, last_fetch_at, created_at, updated_at
		 FROM feeds WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(&f.ID, &f.UserID, &f.CategoryID, &f.FeedURL, &f.SiteURL, &f.Title,
		&f.EtagHeader, &f.LastModified, &f.ParsingError, &f.ParsingErrorCount,
		&f.Disabled, &f.ScraperRules, &f.RewriteRules, &f.Crawler,
		&f.NextCheckAt, &f.LastFetchAt, &f.CreatedAt, &f.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// GetFeedByURL returns the user's feed matching feedURL, or nil.
func (s *Store) GetFeedByURL(ctx context.Context, feedURL string, userID int) (*Feed, error) {
	var f Feed
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, user_id, category_id, feed_url, site_url, title,
		        etag_header, last_modified_header, parsing_error, parsing_error_count,
		        disabled, scraper_rules, rewrite_rules, crawler,
		        next_check_at, last_fetch_at, created_at, updated_at
		 FROM feeds WHERE feed_url = $1 AND user_id = $2`, feedURL, userID,
	).Scan(&f.ID, &f.UserID, &f.CategoryID, &f.FeedURL, &f.SiteURL, &f.Title,
		&f.EtagHeader, &f.LastModified, &f.ParsingError, &f.ParsingErrorCount,
		&f.Disabled, &f.ScraperRules, &f.RewriteRules, &f.Crawler,
		&f.NextCheckAt, &f.LastFetchAt, &f.CreatedAt, &f.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// UpdateFeed updates mutable fields on the given feed.
func (s *Store) UpdateFeed(ctx context.Context, id, userID int, categoryID *int, title, scraperRules, rewriteRules string, disabled, crawler bool) (*Feed, error) {
	var f Feed
	err := s.DB.QueryRowContext(ctx,
		`UPDATE feeds SET category_id = $3, title = $4, scraper_rules = $5,
		         rewrite_rules = $6, disabled = $7, crawler = $8, updated_at = NOW()
		 WHERE id = $1 AND user_id = $2
		 RETURNING id, user_id, category_id, feed_url, site_url, title,
		           etag_header, last_modified_header, parsing_error, parsing_error_count,
		           disabled, scraper_rules, rewrite_rules, crawler,
		           next_check_at, last_fetch_at, created_at, updated_at`,
		id, userID, categoryID, title, scraperRules, rewriteRules, disabled, crawler,
	).Scan(&f.ID, &f.UserID, &f.CategoryID, &f.FeedURL, &f.SiteURL, &f.Title,
		&f.EtagHeader, &f.LastModified, &f.ParsingError, &f.ParsingErrorCount,
		&f.Disabled, &f.ScraperRules, &f.RewriteRules, &f.Crawler,
		&f.NextCheckAt, &f.LastFetchAt, &f.CreatedAt, &f.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update feed: %w", err)
	}
	return &f, nil
}

// DeleteFeed removes a feed and all its entries (via ON DELETE CASCADE).
func (s *Store) DeleteFeed(ctx context.Context, id, userID int) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM feeds WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

// UpdateFeedFetchState stores the ETag/Last-Modified headers and sets
// last_fetch_at + next_check_at after a refresh.
func (s *Store) UpdateFeedFetchState(ctx context.Context, feedID int, etag, lastModified string, parsingError string, errorCount int, nextCheckAt time.Time) error {
	_, err := s.DB.ExecContext(ctx,
		`UPDATE feeds SET etag_header = $2, last_modified_header = $3,
		         parsing_error = $4, parsing_error_count = $5,
		         last_fetch_at = NOW(), next_check_at = $6, updated_at = NOW()
		 WHERE id = $1`,
		feedID, etag, lastModified, parsingError, errorCount, nextCheckAt)
	return err
}

// ListFeedsDueForRefresh returns up to limit feeds whose next_check_at <= now.
func (s *Store) ListFeedsDueForRefresh(ctx context.Context, limit int) ([]Feed, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, user_id, category_id, feed_url, site_url, title,
		        etag_header, last_modified_header, parsing_error, parsing_error_count,
		        disabled, scraper_rules, rewrite_rules, crawler,
		        next_check_at, last_fetch_at, created_at, updated_at
		 FROM feeds WHERE next_check_at <= NOW() AND disabled = false
		 ORDER BY next_check_at ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var feeds []Feed
	for rows.Next() {
		var f Feed
		if err := scanFeed(rows, &f); err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}
	return feeds, rows.Err()
}

// --- Entries ---

// CreateEntry inserts a new entry, returning whether it was actually inserted
// (false when the hash already existed for this feed).
func (s *Store) CreateEntry(ctx context.Context, userID, feedID int, hash, title, url, commentsURL, author, content, description string, publishedAt time.Time, tags []string) (int64, error) {
	var id int64
	tagArr := pq.Array(tags)
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO entries (user_id, feed_id, hash, title, url, comments_url, author, content, description, published_at, tags)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (feed_id, hash) DO NOTHING
		 RETURNING id`,
		userID, feedID, hash, title, url, commentsURL, author, content, description, publishedAt, tagArr,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("create entry: %w", err)
	}
	return id, nil
}

// ListEntries returns entries for the given user, optionally filtered by feed,
// category, status, and starred. Results are paginated.
func (s *Store) ListEntries(ctx context.Context, userID int, feedID *int, categoryID *int, status string, starred *bool, search string, limit, offset int) ([]Entry, error) {
	q := `SELECT e.id, e.user_id, e.feed_id, e.hash, e.title, e.url, e.comments_url,
	             e.author, e.content, e.description, e.status, e.starred,
	             e.published_at, e.changed_at, e.tags
	      FROM entries e`
	args := []interface{}{userID}
	argIdx := 2

	q += " WHERE e.user_id = $1"

	if categoryID != nil {
		q += fmt.Sprintf(" AND e.feed_id IN (SELECT f.id FROM feeds f WHERE f.user_id = $1 AND f.category_id = $%d)", argIdx)
		args = append(args, *categoryID)
		argIdx++
	}
	if feedID != nil {
		q += fmt.Sprintf(" AND e.feed_id = $%d", argIdx)
		args = append(args, *feedID)
		argIdx++
	}

	if status != "" {
		q += fmt.Sprintf(" AND e.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if starred != nil {
		q += fmt.Sprintf(" AND e.starred = $%d", argIdx)
		args = append(args, *starred)
		argIdx++
	}
	if search != "" {
		q += fmt.Sprintf(" AND e.document @@ plainto_tsquery($%d)", argIdx)
		args = append(args, search)
		argIdx++
	}

	q += fmt.Sprintf(" ORDER BY e.published_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.UserID, &e.FeedID, &e.Hash, &e.Title, &e.URL, &e.CommentsURL,
			&e.Author, &e.Content, &e.Description, &e.Status, &e.Starred,
			&e.PublishedAt, &e.ChangedAt, pq.Array(&e.Tags)); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// GetEntryByID returns the entry with the given id scoped to userID, or nil.
func (s *Store) GetEntryByID(ctx context.Context, id int64, userID int) (*Entry, error) {
	var e Entry
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, user_id, feed_id, hash, title, url, comments_url,
		        author, content, description, status, starred,
		        published_at, changed_at, tags
		 FROM entries WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(&e.ID, &e.UserID, &e.FeedID, &e.Hash, &e.Title, &e.URL, &e.CommentsURL,
		&e.Author, &e.Content, &e.Description, &e.Status, &e.Starred,
		&e.PublishedAt, &e.ChangedAt, pq.Array(&e.Tags))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// UpdateEntryStatus sets the status of a set of entries for the given user.
func (s *Store) UpdateEntryStatus(ctx context.Context, entryIDs []int64, userID int, status string) error {
	if len(entryIDs) == 0 {
		return nil
	}
	_, err := s.DB.ExecContext(ctx,
		`UPDATE entries SET status = $3, changed_at = NOW()
		 WHERE id = ANY($1) AND user_id = $2`,
		pq.Array(entryIDs), userID, status)
	return err
}

// ToggleEntryStarred flips the starred flag on a single entry.
func (s *Store) ToggleEntryStarred(ctx context.Context, id int64, userID int, starred bool) error {
	_, err := s.DB.ExecContext(ctx,
		`UPDATE entries SET starred = $3, changed_at = NOW() WHERE id = $1 AND user_id = $2`,
		id, userID, starred)
	return err
}

// MarkFeedEntriesRead sets all unread entries in the given feed to 'read'.
func (s *Store) MarkFeedEntriesRead(ctx context.Context, feedID, userID int) error {
	_, err := s.DB.ExecContext(ctx,
		`UPDATE entries SET status = 'read', changed_at = NOW()
		 WHERE feed_id = $1 AND user_id = $2 AND status = 'unread'`,
		feedID, userID)
	return err
}

// --- Enclosures ---

// CreateEnclosure inserts a media attachment for an entry.
func (s *Store) CreateEnclosure(ctx context.Context, entryID int64, url, mimeType string, size int64) error {
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO enclosures (entry_id, url, mime_type, size) VALUES ($1, $2, $3, $4)
		 ON CONFLICT DO NOTHING`,
		entryID, url, mimeType, size)
	return err
}

// ListEnclosuresByEntry returns all enclosures for the given entries.
func (s *Store) ListEnclosuresByEntry(ctx context.Context, entryIDs []int64) (map[int64][]Enclosure, error) {
	if len(entryIDs) == 0 {
		return nil, nil
	}
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, entry_id, url, mime_type, size FROM enclosures WHERE entry_id = ANY($1)`,
		pq.Array(entryIDs))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	m := make(map[int64][]Enclosure)
	for rows.Next() {
		var enc Enclosure
		if err := rows.Scan(&enc.ID, &enc.EntryID, &enc.URL, &enc.MimeType, &enc.Size); err != nil {
			return nil, err
		}
		m[enc.EntryID] = append(m[enc.EntryID], enc)
	}
	return m, rows.Err()
}

// --- API Tokens ---

// CreateAPIToken inserts an API token for the given user.
func (s *Store) CreateAPIToken(ctx context.Context, userID int, label, tokenHash string) (*APIToken, error) {
	var t APIToken
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO api_tokens (user_id, label, token_hash) VALUES ($1, $2, $3)
		 RETURNING id, user_id, label, token_hash, created_at, last_used_at`,
		userID, label, tokenHash,
	).Scan(&t.ID, &t.UserID, &t.Label, &t.TokenHash, &t.CreatedAt, &t.LastUsedAt)
	if err != nil {
		return nil, fmt.Errorf("create api token: %w", err)
	}
	return &t, nil
}

// ListAPITokens returns API tokens for the given user.
func (s *Store) ListAPITokens(ctx context.Context, userID int) ([]APIToken, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, user_id, label, token_hash, created_at, last_used_at
		 FROM api_tokens WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var tokens []APIToken
	for rows.Next() {
		var t APIToken
		if err := rows.Scan(&t.ID, &t.UserID, &t.Label, &t.TokenHash, &t.CreatedAt, &t.LastUsedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

// GetAPITokenByHash returns the API token matching the hash, or nil. It also
// bumps last_used_at.
func (s *Store) GetAPITokenByHash(ctx context.Context, tokenHash string) (*APIToken, error) {
	var t APIToken
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, user_id, label, token_hash, created_at, last_used_at
		 FROM api_tokens WHERE token_hash = $1`, tokenHash,
	).Scan(&t.ID, &t.UserID, &t.Label, &t.TokenHash, &t.CreatedAt, &t.LastUsedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_, _ = s.DB.ExecContext(ctx, `UPDATE api_tokens SET last_used_at = NOW() WHERE id = $1`, t.ID)
	return &t, nil
}

// DeleteAPIToken removes an API token scoped to the owning user.
func (s *Store) DeleteAPIToken(ctx context.Context, id, userID int) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM api_tokens WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

// --- Users (Limen-managed table, read-only from here) ---

// GetUserByID returns the user with the given id, or nil.
func (s *Store) GetUserByID(ctx context.Context, id int) (*User, error) {
	var u User
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, email, COALESCE(first_name, ''), false, created_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.FirstName, &u.IsAdmin, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// --- Cleanup ---

// PurgeOldEntries removes entries with status 'removed' older than maxAgeDays.
func (s *Store) PurgeOldEntries(ctx context.Context, maxAgeDays int) (int64, error) {
	res, err := s.DB.ExecContext(ctx,
		`DELETE FROM entries WHERE status = 'removed' AND changed_at < NOW() - ($1 || ' days')::interval`,
		fmt.Sprintf("%d", maxAgeDays))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// --- Feed Lists ---

// CreateFeedList inserts a new feed list for the given user.
func (s *Store) CreateFeedList(ctx context.Context, userID int, title, description string, isPublic bool) (*FeedList, error) {
	var fl FeedList
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO feed_lists (user_id, title, description, is_public)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, title, description, is_public, created_at, updated_at`,
		userID, title, description, isPublic,
	).Scan(&fl.ID, &fl.UserID, &fl.Title, &fl.Description, &fl.IsPublic, &fl.CreatedAt, &fl.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create feed list: %w", err)
	}
	return &fl, nil
}

// UpdateFeedList updates mutable fields on a feed list owned by userID.
func (s *Store) UpdateFeedList(ctx context.Context, id, userID int, title, description string, isPublic bool) (*FeedList, error) {
	var fl FeedList
	err := s.DB.QueryRowContext(ctx,
		`UPDATE feed_lists SET title = $3, description = $4, is_public = $5, updated_at = NOW()
		 WHERE id = $1 AND user_id = $2
		 RETURNING id, user_id, title, description, is_public, created_at, updated_at`,
		id, userID, title, description, isPublic,
	).Scan(&fl.ID, &fl.UserID, &fl.Title, &fl.Description, &fl.IsPublic, &fl.CreatedAt, &fl.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update feed list: %w", err)
	}
	return &fl, nil
}

// DeleteFeedList removes a feed list owned by userID. Cascades to feeds/follows.
func (s *Store) DeleteFeedList(ctx context.Context, id, userID int) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM feed_lists WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrFeedListNotFound
	}
	return nil
}

// ListMyFeedLists returns feed lists owned by userID, each with its feed count.
func (s *Store) ListMyFeedLists(ctx context.Context, userID int) ([]FeedList, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT fl.id, fl.user_id, fl.title, fl.description, fl.is_public,
		        fl.created_at, fl.updated_at,
		        (SELECT COUNT(*) FROM feed_list_feeds flf WHERE flf.feed_list_id = fl.id)
		 FROM feed_lists fl
		 WHERE fl.user_id = $1
		 ORDER BY fl.updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var lists []FeedList
	for rows.Next() {
		var fl FeedList
		if err := rows.Scan(&fl.ID, &fl.UserID, &fl.Title, &fl.Description, &fl.IsPublic, &fl.CreatedAt, &fl.UpdatedAt, &fl.FeedCount); err != nil {
			return nil, err
		}
		lists = append(lists, fl)
	}
	return lists, rows.Err()
}

// ListFollowedFeedLists returns feed lists followed by (but not owned by) userID.
func (s *Store) ListFollowedFeedLists(ctx context.Context, userID int) ([]FeedList, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT fl.id, fl.user_id, fl.title, fl.description, fl.is_public,
		        fl.created_at, fl.updated_at,
		        (SELECT COUNT(*) FROM feed_list_feeds flf WHERE flf.feed_list_id = fl.id),
		        u.email
		 FROM feed_list_follows flf
		 JOIN feed_lists fl ON fl.id = flf.feed_list_id
		 JOIN users u ON u.id = fl.user_id
		 WHERE flf.user_id = $1
		 ORDER BY flf.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var lists []FeedList
	for rows.Next() {
		var fl FeedList
		if err := rows.Scan(&fl.ID, &fl.UserID, &fl.Title, &fl.Description, &fl.IsPublic, &fl.CreatedAt, &fl.UpdatedAt, &fl.FeedCount, &fl.OwnerEmail); err != nil {
			return nil, err
		}
		fl.IsFollowing = true
		lists = append(lists, fl)
	}
	return lists, rows.Err()
}

// ListPublicFeedLists returns public feed lists not owned by userID, with
// follow status and owner email. Used for discovery.
func (s *Store) ListPublicFeedLists(ctx context.Context, userID, limit, offset int) ([]FeedList, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT fl.id, fl.user_id, fl.title, fl.description, fl.is_public,
		        fl.created_at, fl.updated_at,
		        (SELECT COUNT(*) FROM feed_list_feeds flf WHERE flf.feed_list_id = fl.id),
		        u.email,
		        EXISTS(SELECT 1 FROM feed_list_follows flf WHERE flf.feed_list_id = fl.id AND flf.user_id = $1)
		 FROM feed_lists fl
		 JOIN users u ON u.id = fl.user_id
		 WHERE fl.is_public = true AND fl.user_id <> $1
		 ORDER BY fl.created_at DESC
		 LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var lists []FeedList
	for rows.Next() {
		var fl FeedList
		if err := rows.Scan(&fl.ID, &fl.UserID, &fl.Title, &fl.Description, &fl.IsPublic, &fl.CreatedAt, &fl.UpdatedAt, &fl.FeedCount, &fl.OwnerEmail, &fl.IsFollowing); err != nil {
			return nil, err
		}
		lists = append(lists, fl)
	}
	return lists, rows.Err()
}

// GetFeedList returns a feed list visible to userID: owned, followed, or public.
// Returns nil if not found or not visible.
func (s *Store) GetFeedList(ctx context.Context, id, userID int) (*FeedList, error) {
	var fl FeedList
	err := s.DB.QueryRowContext(ctx,
		`SELECT fl.id, fl.user_id, fl.title, fl.description, fl.is_public,
		        fl.created_at, fl.updated_at,
		        (SELECT COUNT(*) FROM feed_list_feeds flf WHERE flf.feed_list_id = fl.id),
		        u.email,
		        EXISTS(SELECT 1 FROM feed_list_follows flf WHERE flf.feed_list_id = fl.id AND flf.user_id = $2)
		 FROM feed_lists fl
		 JOIN users u ON u.id = fl.user_id
		 WHERE fl.id = $1 AND (fl.user_id = $2 OR fl.is_public = true OR
		       EXISTS(SELECT 1 FROM feed_list_follows flf WHERE flf.feed_list_id = fl.id AND flf.user_id = $2))`, id, userID,
	).Scan(&fl.ID, &fl.UserID, &fl.Title, &fl.Description, &fl.IsPublic, &fl.CreatedAt, &fl.UpdatedAt, &fl.FeedCount, &fl.OwnerEmail, &fl.IsFollowing)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &fl, nil
}

// AddFeedListFeed adds a feed reference to a feed list owned by userID.
// If the feed URL already exists in the list, it is a no-op (returns the existing row).
func (s *Store) AddFeedListFeed(ctx context.Context, listID, userID int, feedURL, siteURL, title string) (*FeedListFeed, error) {
	var flf FeedListFeed
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO feed_list_feeds (feed_list_id, feed_url, site_url, title, position)
		 SELECT $1, $2, $3, $4, COALESCE(MAX(position), 0) + 1
		 FROM feed_list_feeds WHERE feed_list_id = $1
		 ON CONFLICT (feed_list_id, feed_url) DO UPDATE SET feed_url = EXCLUDED.feed_url
		 RETURNING id, feed_list_id, feed_url, site_url, title, position`,
		listID, feedURL, siteURL, title,
	).Scan(&flf.ID, &flf.FeedListID, &flf.FeedURL, &flf.SiteURL, &flf.Title, &flf.Position)
	if err != nil {
		return nil, fmt.Errorf("add feed list feed: %w", err)
	}
	// Touch updated_at on the parent list.
	_, _ = s.DB.ExecContext(ctx, `UPDATE feed_lists SET updated_at = NOW() WHERE id = $1`, listID)
	_ = userID // ownership enforced by caller (router) via DeleteFeedList-style guard
	return &flf, nil
}

// RemoveFeedListFeed removes a feed from a feed list owned by userID.
func (s *Store) RemoveFeedListFeed(ctx context.Context, listID, itemID, userID int) error {
	res, err := s.DB.ExecContext(ctx,
		`DELETE FROM feed_list_feeds flf USING feed_lists fl
		 WHERE flf.id = $1 AND flf.feed_list_id = $2 AND fl.id = flf.feed_list_id AND fl.user_id = $3`,
		itemID, listID, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrFeedListNotFound
	}
	_, _ = s.DB.ExecContext(ctx, `UPDATE feed_lists SET updated_at = NOW() WHERE id = $1`, listID)
	return nil
}

// ListFeedListFeeds returns all feed references in a feed list visible to userID.
func (s *Store) ListFeedListFeeds(ctx context.Context, listID, userID int) ([]FeedListFeed, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT flf.id, flf.feed_list_id, flf.feed_url, flf.site_url, flf.title, flf.position
		 FROM feed_list_feeds flf
		 JOIN feed_lists fl ON fl.id = flf.feed_list_id
		 WHERE flf.feed_list_id = $1 AND (fl.user_id = $2 OR fl.is_public = true OR
		       EXISTS(SELECT 1 FROM feed_list_follows flfo WHERE flfo.feed_list_id = fl.id AND flfo.user_id = $2))
		 ORDER BY flf.position ASC`, listID, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var feeds []FeedListFeed
	for rows.Next() {
		var flf FeedListFeed
		if err := rows.Scan(&flf.ID, &flf.FeedListID, &flf.FeedURL, &flf.SiteURL, &flf.Title, &flf.Position); err != nil {
			return nil, err
		}
		feeds = append(feeds, flf)
	}
	return feeds, rows.Err()
}

// FollowFeedList creates a follow relationship. Only valid for public lists not
// owned by userID.
func (s *Store) FollowFeedList(ctx context.Context, listID, userID int) error {
	var isPublic, isOwner bool
	err := s.DB.QueryRowContext(ctx,
		`SELECT is_public, (user_id = $2) FROM feed_lists WHERE id = $1`, listID, userID,
	).Scan(&isPublic, &isOwner)
	if err == sql.ErrNoRows {
		return ErrFeedListNotFound
	}
	if err != nil {
		return err
	}
	if isOwner {
		return ErrFeedListOwnList
	}
	if !isPublic {
		return ErrFeedListNotPublic
	}
	_, err = s.DB.ExecContext(ctx,
		`INSERT INTO feed_list_follows (feed_list_id, user_id) VALUES ($1, $2)
		 ON CONFLICT (feed_list_id, user_id) DO NOTHING`, listID, userID)
	return err
}

// UnfollowFeedList removes a follow relationship.
func (s *Store) UnfollowFeedList(ctx context.Context, listID, userID int) error {
	_, err := s.DB.ExecContext(ctx,
		`DELETE FROM feed_list_follows WHERE feed_list_id = $1 AND user_id = $2`, listID, userID)
	return err
}

// IsFeedListOwner returns true if userID owns listID and the list exists.
func (s *Store) IsFeedListOwner(ctx context.Context, listID, userID int) (bool, error) {
	var exists bool
	err := s.DB.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM feed_lists WHERE id = $1 AND user_id = $2)`, listID, userID,
	).Scan(&exists)
	return exists, err
}

// --- scan helper ---

func scanFeed(rows *sql.Rows, f *Feed) error {
	return rows.Scan(&f.ID, &f.UserID, &f.CategoryID, &f.FeedURL, &f.SiteURL, &f.Title,
		&f.EtagHeader, &f.LastModified, &f.ParsingError, &f.ParsingErrorCount,
		&f.Disabled, &f.ScraperRules, &f.RewriteRules, &f.Crawler,
		&f.NextCheckAt, &f.LastFetchAt, &f.CreatedAt, &f.UpdatedAt)
}
