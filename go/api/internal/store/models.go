package store

import (
	"database/sql"
	"time"
)

// Category groups a user's feeds for organisational purposes.
type Category struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

// Feed represents a single RSS/Atom/JSON Feed subscription owned by a user.
type Feed struct {
	ID               int        `json:"id"`
	UserID           int        `json:"user_id"`
	CategoryID       *int       `json:"category_id,omitempty"`
	FeedURL          string     `json:"feed_url"`
	SiteURL          string     `json:"site_url"`
	Title            string     `json:"title"`
	EtagHeader       string     `json:"-"`
	LastModified     string     `json:"-"`
	ParsingError     string     `json:"parsing_error,omitempty"`
	ParsingErrorCount int       `json:"parsing_error_count"`
	Disabled         bool       `json:"disabled"`
	ScraperRules     string     `json:"scraper_rules,omitempty"`
	RewriteRules     string     `json:"rewrite_rules,omitempty"`
	Crawler          bool       `json:"crawler"`
	NextCheckAt      *time.Time `json:"next_check_at,omitempty"`
	LastFetchAt      *time.Time `json:"last_fetch_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// Entry is a single article/item belonging to a feed.
type Entry struct {
	ID          int64     `json:"id"`
	UserID      int       `json:"user_id"`
	FeedID      int       `json:"feed_id"`
	Hash        string    `json:"hash"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	CommentsURL string   `json:"comments_url,omitempty"`
	Author      string    `json:"author,omitempty"`
	Content     string    `json:"content"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	Starred     bool      `json:"starred"`
	PublishedAt time.Time `json:"published_at"`
	ChangedAt   time.Time `json:"changed_at"`
	Tags        []string  `json:"tags,omitempty"`
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
type APIToken struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	Label     string     `json:"label"`
	TokenHash string     `json:"-"`
	CreatedAt time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}

// User represents an authenticated account. Limen owns the users table;
// this struct maps the columns the API needs to read.
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name,omitempty"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
}

// Store wraps a *sql.DB with query helpers for the Planetary schema.
type Store struct {
	DB *sql.DB
}

// New returns a Store backed by the given database.
func New(db *sql.DB) *Store {
	return &Store{DB: db}
}
