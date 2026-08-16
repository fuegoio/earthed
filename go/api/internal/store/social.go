package store

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ErrHandleTaken is returned when a requested handle is already in use.
var ErrHandleTaken = fmt.Errorf("handle already taken")

// ErrHandleInvalid is returned when a handle does not match the allowed format.
var ErrHandleInvalid = fmt.Errorf("handle must be 3–64 characters: letters, digits, hyphens, or underscores")

// ErrProfileNotFound is returned when no profile row exists for the given handle.
var ErrProfileNotFound = fmt.Errorf("user not found")

// ErrAlreadyFollowing is returned when the follow relationship already exists.
var ErrAlreadyFollowing = fmt.Errorf("already following this user")

// ErrCannotFollowSelf is returned when a user tries to follow themselves.
var ErrCannotFollowSelf = fmt.Errorf("cannot follow yourself")

// ErrShareNotFound is returned when a shared article lookup returns no row.
var ErrShareNotFound = fmt.Errorf("shared article not found")

var handleRe = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,64}$`)

// --- Handles / Profiles ---

// UpsertHandle creates or updates the social handle for a user.
// Returns ErrHandleInvalid for bad format, ErrHandleTaken for conflicts.
func (s *Store) UpsertHandle(ctx context.Context, userID int, handle, bio string) (*UserProfile, error) {
	handle = strings.TrimSpace(handle)
	if !handleRe.MatchString(handle) {
		return nil, ErrHandleInvalid
	}

	var p UserProfile
	err := s.DB.QueryRowContext(ctx, `
		INSERT INTO user_profiles (user_id, handle, bio, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id) DO UPDATE
		  SET handle = EXCLUDED.handle,
		      bio    = EXCLUDED.bio,
		      updated_at = NOW()
		RETURNING user_id, handle, bio, created_at, updated_at`,
		userID, handle, bio,
	).Scan(&p.UserID, &p.Handle, &p.Bio, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "idx_user_profiles_handle") ||
			strings.Contains(err.Error(), "unique") {
			return nil, ErrHandleTaken
		}
		return nil, fmt.Errorf("upsert handle: %w", err)
	}
	return &p, nil
}

// GetProfileByHandle returns the public profile for the given handle.
// viewerID is used to populate IsFollowing (0 = no viewer context).
func (s *Store) GetProfileByHandle(ctx context.Context, handle string, viewerID int) (*UserProfile, error) {
	var p UserProfile
	err := s.DB.QueryRowContext(ctx, `
		SELECT
		  p.user_id,
		  p.handle,
		  p.bio,
		  p.created_at,
		  p.updated_at,
		  COALESCE(u.first_name, ''),
		  (SELECT COUNT(*) FROM user_follows WHERE followee_id = p.user_id),
		  (SELECT COUNT(*) FROM user_follows WHERE follower_id = p.user_id),
		  CASE WHEN $2 > 0 THEN
		    EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $2 AND followee_id = p.user_id)
		  ELSE false END
		FROM user_profiles p
		JOIN users u ON u.id = p.user_id
		WHERE p.handle = $1`,
		handle, viewerID,
	).Scan(
		&p.UserID, &p.Handle, &p.Bio,
		&p.CreatedAt, &p.UpdatedAt,
		&p.FirstName,
		&p.FollowerCount, &p.FollowingCount,
		&p.IsFollowing,
	)
	if err == sql.ErrNoRows {
		return nil, ErrProfileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return &p, nil
}

// GetProfileByUserID returns the social profile for the given user id, or nil.
func (s *Store) GetProfileByUserID(ctx context.Context, userID int) (*UserProfile, error) {
	var p UserProfile
	err := s.DB.QueryRowContext(ctx, `
		SELECT user_id, handle, bio, created_at, updated_at
		FROM user_profiles WHERE user_id = $1`, userID,
	).Scan(&p.UserID, &p.Handle, &p.Bio, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get profile by user id: %w", err)
	}
	return &p, nil
}

// --- Follows ---

// FollowUser creates a follow relationship. Returns ErrCannotFollowSelf or
// ErrAlreadyFollowing when applicable, ErrProfileNotFound when the followee
// has no profile.
func (s *Store) FollowUser(ctx context.Context, followerID int, followeeHandle string) error {
	// Resolve handle → user_id.
	var followeeID int
	err := s.DB.QueryRowContext(ctx,
		`SELECT user_id FROM user_profiles WHERE handle = $1`, followeeHandle,
	).Scan(&followeeID)
	if err == sql.ErrNoRows {
		return ErrProfileNotFound
	}
	if err != nil {
		return fmt.Errorf("resolve handle: %w", err)
	}
	if followerID == followeeID {
		return ErrCannotFollowSelf
	}

	_, err = s.DB.ExecContext(ctx,
		`INSERT INTO user_follows (follower_id, followee_id) VALUES ($1, $2)`,
		followerID, followeeID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return ErrAlreadyFollowing
		}
		return fmt.Errorf("follow user: %w", err)
	}
	return nil
}

// UnfollowUser removes the follow relationship.
func (s *Store) UnfollowUser(ctx context.Context, followerID int, followeeHandle string) error {
	var followeeID int
	err := s.DB.QueryRowContext(ctx,
		`SELECT user_id FROM user_profiles WHERE handle = $1`, followeeHandle,
	).Scan(&followeeID)
	if err == sql.ErrNoRows {
		return ErrProfileNotFound
	}
	if err != nil {
		return fmt.Errorf("resolve handle: %w", err)
	}

	_, err = s.DB.ExecContext(ctx,
		`DELETE FROM user_follows WHERE follower_id = $1 AND followee_id = $2`,
		followerID, followeeID,
	)
	return err
}

// ListFollowing returns the profiles of users that followerID is following.
func (s *Store) ListFollowing(ctx context.Context, followerID int) ([]UserProfile, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT
		  p.user_id, p.handle, p.bio, p.created_at, p.updated_at,
		  COALESCE(u.first_name, ''),
		  (SELECT COUNT(*) FROM user_follows WHERE followee_id = p.user_id),
		  (SELECT COUNT(*) FROM user_follows WHERE follower_id = p.user_id)
		FROM user_follows f
		JOIN user_profiles p ON p.user_id = f.followee_id
		JOIN users u ON u.id = p.user_id
		WHERE f.follower_id = $1
		ORDER BY f.created_at DESC`, followerID,
	)
	if err != nil {
		return nil, fmt.Errorf("list following: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []UserProfile
	for rows.Next() {
		var p UserProfile
		if err := rows.Scan(
			&p.UserID, &p.Handle, &p.Bio, &p.CreatedAt, &p.UpdatedAt,
			&p.FirstName, &p.FollowerCount, &p.FollowingCount,
		); err != nil {
			return nil, err
		}
		p.IsFollowing = true
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListFollowers returns the profiles of users following userID.
func (s *Store) ListFollowers(ctx context.Context, userID, viewerID int) ([]UserProfile, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT
		  p.user_id, p.handle, p.bio, p.created_at, p.updated_at,
		  COALESCE(u.first_name, ''),
		  (SELECT COUNT(*) FROM user_follows WHERE followee_id = p.user_id),
		  (SELECT COUNT(*) FROM user_follows WHERE follower_id = p.user_id),
		  CASE WHEN $2 > 0 THEN
		    EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $2 AND followee_id = p.user_id)
		  ELSE false END
		FROM user_follows f
		JOIN user_profiles p ON p.user_id = f.follower_id
		JOIN users u ON u.id = p.user_id
		WHERE f.followee_id = $1
		ORDER BY f.created_at DESC`, userID, viewerID,
	)
	if err != nil {
		return nil, fmt.Errorf("list followers: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []UserProfile
	for rows.Next() {
		var p UserProfile
		if err := rows.Scan(
			&p.UserID, &p.Handle, &p.Bio, &p.CreatedAt, &p.UpdatedAt,
			&p.FirstName, &p.FollowerCount, &p.FollowingCount, &p.IsFollowing,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// --- Shared articles ---

// ShareArticle creates or replaces a shared article for the user.
func (s *Store) ShareArticle(ctx context.Context, userID int,
	articleURL, title, description, feedURL, feedTitle, feedSiteURL, author string,
	publishedAt *time.Time,
) (*SharedArticle, error) {
	var sa SharedArticle
	err := s.DB.QueryRowContext(ctx, `
		INSERT INTO shared_articles
		  (user_id, article_url, title, description, feed_url, feed_title, feed_site_url, author, published_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (user_id, article_url) DO UPDATE
		  SET title        = EXCLUDED.title,
		      description  = EXCLUDED.description,
		      feed_url     = EXCLUDED.feed_url,
		      feed_title   = EXCLUDED.feed_title,
		      feed_site_url= EXCLUDED.feed_site_url,
		      author       = EXCLUDED.author,
		      published_at = EXCLUDED.published_at,
		      shared_at    = NOW()
		RETURNING id, user_id, article_url, title, description,
		          feed_url, feed_title, feed_site_url, author, published_at, shared_at`,
		userID, articleURL, title, description, feedURL, feedTitle, feedSiteURL, author, publishedAt,
	).Scan(
		&sa.ID, &sa.UserID, &sa.ArticleURL, &sa.Title, &sa.Description,
		&sa.FeedURL, &sa.FeedTitle, &sa.FeedSiteURL, &sa.Author, &sa.PublishedAt, &sa.SharedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("share article: %w", err)
	}
	return &sa, nil
}

// UnshareArticle removes a shared article owned by userID.
func (s *Store) UnshareArticle(ctx context.Context, id int64, userID int) error {
	res, err := s.DB.ExecContext(ctx,
		`DELETE FROM shared_articles WHERE id = $1 AND user_id = $2`, id, userID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrShareNotFound
	}
	return nil
}

// GetSharedArticleByURL returns the share for a given user+URL, or nil.
func (s *Store) GetSharedArticleByURL(ctx context.Context, userID int, articleURL string) (*SharedArticle, error) {
	var sa SharedArticle
	err := s.DB.QueryRowContext(ctx, `
		SELECT id, user_id, article_url, title, description,
		       feed_url, feed_title, feed_site_url, author, published_at, shared_at
		FROM shared_articles WHERE user_id = $1 AND article_url = $2`,
		userID, articleURL,
	).Scan(
		&sa.ID, &sa.UserID, &sa.ArticleURL, &sa.Title, &sa.Description,
		&sa.FeedURL, &sa.FeedTitle, &sa.FeedSiteURL, &sa.Author, &sa.PublishedAt, &sa.SharedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sa, nil
}

// ListSharedArticlesByUser returns shared articles for the given user_id,
// newest first. Used for profile view.
func (s *Store) ListSharedArticlesByUser(ctx context.Context, userID int) ([]SharedArticle, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, user_id, article_url, title, description,
		       feed_url, feed_title, feed_site_url, author, published_at, shared_at
		FROM shared_articles
		WHERE user_id = $1
		ORDER BY shared_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []SharedArticle
	for rows.Next() {
		var sa SharedArticle
		if err := rows.Scan(
			&sa.ID, &sa.UserID, &sa.ArticleURL, &sa.Title, &sa.Description,
			&sa.FeedURL, &sa.FeedTitle, &sa.FeedSiteURL, &sa.Author, &sa.PublishedAt, &sa.SharedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, sa)
	}
	return out, rows.Err()
}

// ListSocialTimeline returns shared articles from users that followerID follows,
// newest first. Includes sharer metadata for rendering.
func (s *Store) ListSocialTimeline(ctx context.Context, followerID, limit, offset int) ([]SharedArticle, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT
		  sa.id, sa.user_id, sa.article_url, sa.title, sa.description,
		  sa.feed_url, sa.feed_title, sa.feed_site_url, sa.author,
		  sa.published_at, sa.shared_at,
		  COALESCE(p.handle, ''),
		  COALESCE(u.first_name, '')
		FROM shared_articles sa
		JOIN user_follows f ON f.followee_id = sa.user_id AND f.follower_id = $1
		JOIN user_profiles p ON p.user_id = sa.user_id
		JOIN users u ON u.id = sa.user_id
		ORDER BY sa.shared_at DESC
		LIMIT $2 OFFSET $3`,
		followerID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []SharedArticle
	for rows.Next() {
		var sa SharedArticle
		if err := rows.Scan(
			&sa.ID, &sa.UserID, &sa.ArticleURL, &sa.Title, &sa.Description,
			&sa.FeedURL, &sa.FeedTitle, &sa.FeedSiteURL, &sa.Author,
			&sa.PublishedAt, &sa.SharedAt,
			&sa.SharerHandle, &sa.SharerFirstName,
		); err != nil {
			return nil, err
		}
		out = append(out, sa)
	}
	return out, rows.Err()
}

// --- Feed subscribers ---

// CountFeedSubscribers returns the number of users subscribed to a feed URL.
func (s *Store) CountFeedSubscribers(ctx context.Context, feedURL string) (int, error) {
	var n int
	err := s.DB.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT user_id) FROM feeds WHERE feed_url = $1`, feedURL,
	).Scan(&n)
	return n, err
}

// ListFeedSubscribers returns the public profiles of users subscribed to feedURL.
// Profiles without a handle are excluded (anonymous subscribers).
func (s *Store) ListFeedSubscribers(ctx context.Context, feedURL string) ([]UserProfile, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT
		  p.user_id, p.handle, p.bio, p.created_at, p.updated_at,
		  COALESCE(u.first_name, ''),
		  (SELECT COUNT(*) FROM user_follows WHERE followee_id = p.user_id),
		  (SELECT COUNT(*) FROM user_follows WHERE follower_id = p.user_id)
		FROM feeds f
		JOIN user_profiles p ON p.user_id = f.user_id
		JOIN users u ON u.id = f.user_id
		WHERE f.feed_url = $1
		ORDER BY p.handle ASC`, feedURL,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []UserProfile
	for rows.Next() {
		var p UserProfile
		if err := rows.Scan(
			&p.UserID, &p.Handle, &p.Bio, &p.CreatedAt, &p.UpdatedAt,
			&p.FirstName, &p.FollowerCount, &p.FollowingCount,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListPublicFeedsByUser returns the feed subscriptions of a user (for public profile view).
func (s *Store) ListPublicFeedsByUser(ctx context.Context, userID int) ([]Feed, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, user_id, folder_id, feed_url, site_url, title, description,
		       etag_header, last_modified_header,
		       parsing_error, parsing_error_count, disabled,
		       scraper_rules, rewrite_rules, crawler,
		       next_check_at, last_fetch_at, created_at, updated_at
		FROM feeds
		WHERE user_id = $1
		ORDER BY title ASC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []Feed
	for rows.Next() {
		var f Feed
		if err := rows.Scan(
			&f.ID, &f.UserID, &f.FolderID, &f.FeedURL, &f.SiteURL, &f.Title, &f.Description,
			&f.EtagHeader, &f.LastModified,
			&f.ParsingError, &f.ParsingErrorCount, &f.Disabled,
			&f.ScraperRules, &f.RewriteRules, &f.Crawler,
			&f.NextCheckAt, &f.LastFetchAt, &f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}
