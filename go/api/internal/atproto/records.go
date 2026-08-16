package atproto

import "time"

// Lexicon collection IDs for io.planetary.* record types.
const (
	CollectionProfile      = "io.planetary.actor.profile"
	CollectionFollow       = "io.planetary.graph.follow"
	CollectionShare        = "io.planetary.share.article"
	CollectionSubscription = "io.planetary.feed.subscription"
	CollectionFeedList     = "io.planetary.feed.list"
)

// ProfileRecord is the io.planetary.actor.profile record.
// Stored at rkey "self" — one per user repo.
type ProfileRecord struct {
	Type        string `json:"$type"`
	Handle      string `json:"handle"`
	Bio         string `json:"bio,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	InstanceURL string `json:"instanceUrl,omitempty"`
	CreatedAt   string `json:"createdAt"`
}

// FollowRecord is the io.planetary.graph.follow record.
type FollowRecord struct {
	Type      string `json:"$type"`
	Subject   string `json:"subject"` // followee DID
	CreatedAt string `json:"createdAt"`
}

// ShareRecord is the io.planetary.share.article record.
type ShareRecord struct {
	Type        string `json:"$type"`
	ArticleURL  string `json:"articleUrl"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	FeedURL     string `json:"feedUrl,omitempty"`
	FeedTitle   string `json:"feedTitle,omitempty"`
	FeedSiteURL string `json:"feedSiteUrl,omitempty"`
	Author      string `json:"author,omitempty"`
	PublishedAt string `json:"publishedAt,omitempty"`
	SharedAt    string `json:"sharedAt"`
}

// SubscriptionRecord is the io.planetary.feed.subscription record.
type SubscriptionRecord struct {
	Type      string `json:"$type"`
	FeedURL   string `json:"feedUrl"`
	SiteURL   string `json:"siteUrl,omitempty"`
	Title     string `json:"title,omitempty"`
	CreatedAt string `json:"createdAt"`
}

// FeedListRecord is the io.planetary.feed.list record.
type FeedListRecord struct {
	Type        string          `json:"$type"`
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	IsPublic    bool            `json:"isPublic"`
	Feeds       []FeedListEntry `json:"feeds,omitempty"`
	CreatedAt   string          `json:"createdAt"`
}

// FeedListEntry is a single feed within a FeedListRecord.
type FeedListEntry struct {
	FeedURL string `json:"feedUrl"`
	SiteURL string `json:"siteUrl,omitempty"`
	Title   string `json:"title,omitempty"`
}

// FormatTime formats a time.Time as an AT Proto datetime string (RFC3339).
func FormatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
