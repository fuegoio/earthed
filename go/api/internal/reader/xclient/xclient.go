// Package xclient fetches a person's X (Twitter) timeline via the official
// X API v2 and normalizes the response into the same Feed/Item shapes the
// RSS parser produces, so the feed processor can store X posts alongside
// RSS articles without a separate pipeline.
//
// It uses the user Tweet timeline endpoint:
//
//	GET /2/users/{id}/tweets
//
// and resolves a username to a numeric user ID via:
//
//	GET /2/users/by/username/{username}
package xclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fuegoio/earthed/go/api/internal/reader/parser"
)

const (
	// DefaultBaseURL is the official X API v2 root.
	DefaultBaseURL = "https://api.x.com/2"

	// defaultMaxResults is the number of posts requested per timeline
	// fetch. The X API requires 5 <= max_results <= 100.
	defaultMaxResults = 100

	// tweetFields asks the X API for the fields we need to build entries.
	tweetFields = "created_at,author_id,text,entities,referenced_tweets,public_metrics"

	// expansions pulls the author into the includes block so we can show
	// the display name and @username on each post.
	expansions = "author_id"

	// userFields asks for the author fields we render.
	userFields = "name,username,profile_image_url"

	// excludeReplies keeps the timeline focused on the user's own posts.
	excludeReplies = "replies,retweets"
)

// Client fetches X timelines via the official X API v2 using a bearer token.
type Client struct {
	baseURL   string
	token     string
	userAgent string
	client    *http.Client
}

// New returns a Client authenticated with the given App-only bearer token.
// An empty token yields a client whose methods return an error when called,
// so callers can construct it unconditionally and gate on Enabled().
func New(baseURL, bearerToken, userAgent string, timeout time.Duration) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		token:     bearerToken,
		userAgent: userAgent,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Enabled reports whether the client is configured with a bearer token.
func (c *Client) Enabled() bool {
	return c != nil && c.token != ""
}

// timelineResponse models the subset of the X API v2
// GET /2/users/{id}/tweets response that we consume.
type timelineResponse struct {
	Data []struct {
		ID               string            `json:"id"`
		Text             string            `json:"text"`
		AuthorID         string            `json:"author_id"`
		CreatedAt        string            `json:"created_at"`
		PublicMetrics    publicMetrics     `json:"public_metrics"`
		ReferencedTweets []referencedTweet `json:"referenced_tweets,omitempty"`
		Entities         *entities         `json:"entities,omitempty"`
	} `json:"data"`
	Includes *includes `json:"includes,omitempty"`
	Meta     struct {
		NewestID    string `json:"newest_id"`
		OldestID    string `json:"oldest_id"`
		NextToken   string `json:"next_token"`
		ResultCount int    `json:"result_count"`
	} `json:"meta"`
}

type publicMetrics struct {
	ImpressionCount int `json:"impression_count"`
	LikeCount       int `json:"like_count"`
	ReplyCount      int `json:"reply_count"`
	RepostCount     int `json:"repost_count"`
	QuoteCount      int `json:"quote_count"`
	BookmarkCount   int `json:"bookmark_count"`
}

type referencedTweet struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type entities struct {
	URLs []struct {
		URL         string `json:"url"`
		ExpandedURL string `json:"expanded_url"`
		DisplayURL  string `json:"display_url"`
		UnwoundURL  string `json:"unwound_url,omitempty"`
	} `json:"urls"`
	Hashtags []struct {
		Tag string `json:"tag"`
	} `json:"hashtags"`
	Mentions []struct {
		Username string `json:"username"`
	} `json:"mentions"`
}

type includes struct {
	Users []struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		Username        string `json:"username"`
		ProfileImageURL string `json:"profile_image_url"`
	} `json:"users"`
}

// userResponse models GET /2/users/by/username/{username}.
type userResponse struct {
	Data *struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		Username        string `json:"username"`
		ProfileImageURL string `json:"profile_image_url"`
	} `json:"data"`
	Errors []xAPIError `json:"errors,omitempty"`
}

// xAPIError is the standard error object returned by the X API.
type xAPIError struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Status int    `json:"status"`
	Type   string `json:"type"`
}

// User is a resolved X account.
type User struct {
	ID              string
	Name            string
	Username        string
	ProfileImageURL string
}

// ResolveUsername looks up the numeric user ID and display metadata for an X
// username via GET /2/users/by/username/{username}.
func (c *Client) ResolveUsername(ctx context.Context, username string) (*User, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	username = strings.TrimPrefix(strings.TrimSpace(username), "@")
	if username == "" {
		return nil, fmt.Errorf("xclient: empty username")
	}

	endpoint := fmt.Sprintf("%s/users/by/username/%s", c.baseURL, url.PathEscape(username))
	var resp userResponse
	if err := c.getJSON(ctx, endpoint, url.Values{
		"user.fields": []string{userFields},
	}, &resp); err != nil {
		return nil, err
	}
	if len(resp.Errors) > 0 {
		e := resp.Errors[0]
		return nil, fmt.Errorf("xclient: lookup @%s: %s", username, e.Detail)
	}
	if resp.Data == nil {
		return nil, fmt.Errorf("xclient: user @%s not found", username)
	}
	return &User{
		ID:              resp.Data.ID,
		Name:            resp.Data.Name,
		Username:        resp.Data.Username,
		ProfileImageURL: resp.Data.ProfileImageURL,
	}, nil
}

// FetchTimeline fetches the most recent posts from the X user identified by
// userID and normalizes them into a parser.Feed so the existing feed
// processor can store them. The author name and @username come from the
// includes block; when the author is absent the numeric author_id is used.
func (c *Client) FetchTimeline(ctx context.Context, userID string) (*parser.Feed, error) {
	if !c.Enabled() {
		return nil, ErrNotConfigured
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("xclient: empty user id")
	}

	endpoint := fmt.Sprintf("%s/users/%s/tweets", c.baseURL, url.PathEscape(userID))
	params := url.Values{
		"max_results":  []string{strconv.Itoa(defaultMaxResults)},
		"tweet.fields": []string{tweetFields},
		"expansions":   []string{expansions},
		"user.fields":  []string{userFields},
		"exclude":      []string{excludeReplies},
	}

	var resp timelineResponse
	if err := c.getJSON(ctx, endpoint, params, &resp); err != nil {
		return nil, err
	}

	// Build author_id -> user lookup from the includes block.
	authors := make(map[string]User)
	siteURL := ""
	siteAuthor := ""
	if resp.Includes != nil {
		for _, u := range resp.Includes.Users {
			authors[u.ID] = User{ID: u.ID, Name: u.Name, Username: u.Username, ProfileImageURL: u.ProfileImageURL}
		}
		// Use the first (and usually only) included user as the feed author.
		if len(resp.Includes.Users) > 0 {
			u := resp.Includes.Users[0]
			siteAuthor = u.Name
			siteURL = "https://x.com/" + u.Username
		}
	}

	feed := &parser.Feed{
		Title:       siteAuthor,
		SiteURL:     siteURL,
		Description: "",
	}

	for _, tw := range resp.Data {
		authorName := tw.AuthorID
		if u, ok := authors[tw.AuthorID]; ok {
			authorName = u.Name
		}

		// Replace t.co short URLs with their expanded form so the post
		// body is readable and links resolve correctly.
		text := expandURLs(tw.Text, tw.Entities)

		tags := collectTags(tw.Entities)

		feed.Items = append(feed.Items, parser.Item{
			Title:       truncateForTitle(text),
			Link:        postURL(tw.ID),
			Description: text,
			Content:     text,
			Author:      authorName,
			PublishedAt: tw.CreatedAt,
			Tags:        tags,
		})
	}

	return feed, nil
}

// getJSON performs an authenticated GET against endpoint with the given
// query parameters and decodes the JSON response into out.
func (c *Client) getJSON(ctx context.Context, endpoint string, params url.Values, out interface{}) error {
	if len(params) > 0 {
		endpoint = endpoint + "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("xclient: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("xclient: request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return c.decodeError(resp)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("xclient: decode response: %w", err)
	}
	return nil
}

// decodeError reads a non-200 X API response and returns a descriptive error.
func (c *Client) decodeError(resp *http.Response) error {
	var body struct {
		Title  string      `json:"title"`
		Detail string      `json:"detail"`
		Errors []xAPIError `json:"errors"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	detail := body.Detail
	if detail == "" && len(body.Errors) > 0 {
		detail = body.Errors[0].Detail
	}
	if detail == "" {
		detail = http.StatusText(resp.StatusCode)
	}
	return fmt.Errorf("xclient: HTTP %d: %s", resp.StatusCode, detail)
}

// postURL returns the canonical permalink for an X post.
func postURL(postID string) string {
	return "https://x.com/i/status/" + postID
}

// expandURLs replaces t.co short links in text with their expanded URLs.
func expandURLs(text string, ents *entities) string {
	if ents == nil {
		return text
	}
	out := text
	for _, u := range ents.URLs {
		target := u.UnwoundURL
		if target == "" {
			target = u.ExpandedURL
		}
		if target != "" {
			out = strings.Replace(out, u.URL, target, 1)
		}
	}
	return out
}

// collectTags extracts hashtag tags from entities into a slice for storage.
func collectTags(ents *entities) []string {
	if ents == nil || len(ents.Hashtags) == 0 {
		return nil
	}
	tags := make([]string, 0, len(ents.Hashtags))
	for _, h := range ents.Hashtags {
		if h.Tag != "" {
			tags = append(tags, h.Tag)
		}
	}
	return tags
}

// truncateForTitle produces a short, single-line title from a post's text,
// since X posts often have no separate title field. X does not impose a
// hard limit shorter than 280 characters, but a title is kept readable.
func truncateForTitle(text string) string {
	text = strings.TrimSpace(strings.ReplaceAll(text, "\n", " "))
	if len(text) <= 140 {
		return text
	}
	return text[:139] + "…"
}
