// Package discoverer discovers RSS/Atom/JSON feed URLs from a given URL.
// If the URL is already a feed, it is returned as-is. If the URL is an
// HTML page, the discoverer parses <link rel="alternate"> tags to find
// feed URLs.
package discoverer

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/html"

	"github.com/fuegoio/sunred/go/api/internal/reader/fetcher"
	"github.com/fuegoio/sunred/go/api/internal/reader/parser"
)

// Result holds the outcome of feed discovery.
type Result struct {
	// SiteURL is the URL of the website (the page that was fetched when
	// discovery started from an HTML page). Empty when the input URL was
	// already a feed.
	SiteURL string

	// FeedURL is the discovered feed URL. When the input URL is already a
	// feed, this equals the input URL.
	FeedURL string
}

// Discover fetches the given URL. If the response is a feed, it returns
// the URL directly. If the response is HTML, it parses <link
// rel="alternate"> tags to find feed URLs and returns the first one
// found.
func Discover(ctx context.Context, f *fetcher.Fetcher, inputURL string) (*Result, error) {
	result, err := f.Fetch(ctx, inputURL, "", "")
	if err != nil {
		return nil, fmt.Errorf("discover %s: %w", inputURL, err)
	}
	if result.NotModified {
		return nil, fmt.Errorf("discover %s: 304 Not Modified", inputURL)
	}

	// If the content parses as a feed, the URL is already a feed URL.
	if _, err := parser.Parse(result.Body, result.ContentType); err == nil {
		return &Result{FeedURL: inputURL}, nil
	}

	// Otherwise, treat the response as HTML and look for <link
	// rel="alternate"> tags pointing to feeds.
	feedURLs := extractFeedLinks(result.Body, inputURL)
	if len(feedURLs) == 0 {
		return nil, fmt.Errorf("discover %s: no feed links found in HTML", inputURL)
	}

	return &Result{
		SiteURL: inputURL,
		FeedURL: feedURLs[0],
	}, nil
}

// feedMimeTypes lists MIME types and substrings that indicate a feed
// link in an HTML <link> tag.
var feedMimeTypes = []string{
	"application/rss+xml",
	"application/atom+xml",
	"application/json",
	"application/feed+json",
	"application/xml",
	"text/xml",
}

// isFeedType returns true if the given type attribute value looks like
// a feed MIME type.
func isFeedType(t string) bool {
	t = strings.ToLower(strings.TrimSpace(t))
	for _, mt := range feedMimeTypes {
		if strings.Contains(t, mt) {
			return true
		}
	}
	return false
}

// extractFeedLinks parses an HTML document and returns all feed URLs
// found in <link rel="alternate"> tags, ordered by preference (RSS
// before Atom before JSON).
func extractFeedLinks(body []byte, baseURL string) []string {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil
	}

	type linkCandidate struct {
		href string
		rank int // lower is better: rss=0, atom=1, json=2
	}

	var candidates []linkCandidate
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "link" {
			var rel, href, linkType string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "rel":
					rel = strings.ToLower(strings.TrimSpace(attr.Val))
				case "href":
					href = attr.Val
				case "type":
					linkType = strings.ToLower(strings.TrimSpace(attr.Val))
				}
			}
			if strings.Contains(rel, "alternate") && href != "" && isFeedType(linkType) {
				resolved := resolveURL(href, baseURL)
				if resolved == "" {
					return
				}
				rank := 3
				switch {
				case strings.Contains(linkType, "rss"):
					rank = 0
				case strings.Contains(linkType, "atom"):
					rank = 1
				case strings.Contains(linkType, "json"):
					rank = 2
				}
				candidates = append(candidates, linkCandidate{href: resolved, rank: rank})
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	// Sort by rank (stable to preserve document order within same rank).
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].rank < candidates[i].rank {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	var urls []string
	seen := make(map[string]bool)
	for _, c := range candidates {
		if !seen[c.href] {
			seen[c.href] = true
			urls = append(urls, c.href)
		}
	}
	return urls
}

// resolveURL resolves a possibly-relative href against the base URL.
func resolveURL(href, base string) string {
	if href == "" {
		return ""
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return href
	}
	hrefURL, err := url.Parse(href)
	if err != nil {
		return href
	}
	return baseURL.ResolveReference(hrefURL).String()
}
