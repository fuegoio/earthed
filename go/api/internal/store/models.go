package store

import (
	"database/sql"
	"time"
)

// Folder groups a user's feeds for organisational purposes. Folders can be
// nested via ParentID and ordered via SortOrder.
type Folder struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ParentID  *int      `json:"parent_id,omitempty"`
	Title     string    `json:"title"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Feed represents a single RSS/Atom/JSON Feed subscription owned by a user.
type Feed struct {
	ID                int        `json:"id"`
	UserID            int        `json:"user_id"`
	FolderID          *int       `json:"folder_id,omitempty"`
	FeedURL           string     `json:"feed_url"`
	SiteURL           string     `json:"site_url"`
	Title             string     `json:"title"`
	Description       string     `json:"description,omitempty"`
	EtagHeader        string     `json:"-"`
	LastModified      string     `json:"-"`
	ParsingError      string     `json:"parsing_error,omitempty"`
	ParsingErrorCount int        `json:"parsing_error_count"`
	Disabled          bool       `json:"disabled"`
	ScraperRules      string     `json:"scraper_rules,omitempty"`
	RewriteRules      string     `json:"rewrite_rules,omitempty"`
	Crawler           bool       `json:"crawler"`
	NextCheckAt       *time.Time `json:"next_check_at,omitempty"`
	LastFetchAt       *time.Time `json:"last_fetch_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// Entry is a single article/item belonging to a feed.
type Entry struct {
	ID          int64       `json:"id"`
	UserID      int         `json:"user_id"`
	FeedID      int         `json:"feed_id"`
	Hash        string      `json:"hash"`
	Title       string      `json:"title"`
	URL         string      `json:"url"`
	CommentsURL string      `json:"comments_url,omitempty"`
	Author      string      `json:"author,omitempty"`
	Content     string      `json:"-"`
	Description string      `json:"description,omitempty"`
	Status      string      `json:"status"`
	Starred     bool        `json:"starred"`
	PublishedAt time.Time   `json:"published_at"`
	ChangedAt   time.Time   `json:"changed_at"`
	Tags        []string    `json:"tags,omitempty"`
	Enclosures  []Enclosure `json:"enclosures,omitempty"`
}

// Enclosure is a media attachment (podcast, image, file) on an entry.
type Enclosure struct {
	ID       int64  `json:"id"`
	EntryID  int64  `json:"entry_id"`
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

// FeedIcon stores the favicon data for a feed.
type FeedIcon struct {
	FeedID int    `json:"feed_id"`
	Data   []byte `json:"-"`
}

// APIToken is a hashed API token issued to a user for bearer auth.
// ExpiresAt is nil for non-expiring tokens (e.g. created via the web UI);
// device-flow tokens carry a 14-day expiry. Origin records how the token
// was issued: "manual" or "device_flow".
type APIToken struct {
	ID         int        `json:"id"`
	UserID     int        `json:"user_id"`
	Label      string     `json:"label"`
	TokenHash  string     `json:"-"`
	Origin     string     `json:"origin"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// DeviceCode is an in-flight RFC 8628 device authorization grant.
// DeviceCode is stored hashed (SHA-256), like APIToken.TokenHash; the
// plaintext is only known to the CLI/TUI that initiated the flow.
// TokenPlaintext is set on confirm and returned once to the polling CLI,
// then the grant row is deleted (single-use).
type DeviceCode struct {
	ID             int64      `json:"-"`
	DeviceCode     string     `json:"-"`         // hash
	UserCode       string     `json:"user_code"` // "PLN-XXXX-XXXX"
	Status         string     `json:"status"`    // pending|authorized|denied|expired
	UserID         *int       `json:"-"`
	TokenID        *int       `json:"-"`
	TokenPlaintext *string    `json:"-"` // populated on authorize, consumed once
	IntervalSecs   int        `json:"interval"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      time.Time  `json:"expires_at"`
	LastPolledAt   *time.Time `json:"-"`
}

// FeedList is a curated, shareable collection of feeds owned by a user.
type FeedList struct {
	ID          int            `json:"id"`
	UserID      int            `json:"user_id"`
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	IsPublic    bool           `json:"is_public"`
	Feeds       []FeedListFeed `json:"feeds,omitempty"`
	FeedCount   int            `json:"feed_count"`
	IsFollowing bool           `json:"is_following,omitempty"`
	OwnerEmail  string         `json:"owner_email,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// FeedListFeed is a single feed reference within a feed list. It stores a
// snapshot of the feed metadata (url, title) so the list is portable and can
// be displayed/imported by other users without referencing the owner's feeds.
type FeedListFeed struct {
	ID         int    `json:"id"`
	FeedListID int    `json:"feed_list_id"`
	FeedURL    string `json:"feed_url"`
	SiteURL    string `json:"site_url"`
	Title      string `json:"title"`
	Position   int    `json:"position"`
}

// User represents an authenticated account. Limen owns the users table;
// this struct maps the columns the API needs to read.
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name,omitempty"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	// Handle is the user's social handle (e.g. "fuego"). Populated when a
	// profile row exists; empty string otherwise.
	Handle string `json:"handle,omitempty"`
}

// UserProfile holds the social profile data for a user.
type UserProfile struct {
	UserID    int       `json:"user_id"`
	Handle    string    `json:"handle"`
	Bio       string    `json:"bio,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// Denormalised fields set by query joins.
	FirstName      string `json:"first_name,omitempty"`
	FollowerCount  int    `json:"follower_count"`
	FollowingCount int    `json:"following_count"`
	IsFollowing    bool   `json:"is_following,omitempty"`
}

// SharedArticle is an article that a user shared on the social timeline.
type SharedArticle struct {
	ID           int64      `json:"id"`
	UserID       int        `json:"user_id"`
	ArticleURL   string     `json:"article_url"`
	Title        string     `json:"title"`
	Description  string     `json:"description,omitempty"`
	FeedURL      string     `json:"feed_url,omitempty"`
	FeedTitle    string     `json:"feed_title,omitempty"`
	FeedSiteURL  string     `json:"feed_site_url,omitempty"`
	Author       string     `json:"author,omitempty"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	SharedAt     time.Time  `json:"shared_at"`
	// Sharer info, populated on social timeline queries.
	SharerHandle    string `json:"sharer_handle,omitempty"`
	SharerFirstName string `json:"sharer_first_name,omitempty"`
}

// Store wraps a *sql.DB with query helpers for the Earthed schema.
type Store struct {
	DB *sql.DB
}

// New returns a Store backed by the given database.
func New(db *sql.DB) *Store {
	return &Store{DB: db}
}
