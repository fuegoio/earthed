// Package parser detects and parses feed formats: RSS 2.0, Atom 1.0, RDF/RSS 1.0,
// and JSON Feed. It returns a normalized Feed representation.
package parser

import (
	"encoding/xml"
	"fmt"
	"strings"

	"encoding/json"
)

// Feed is the normalized representation of a parsed feed, regardless of format.
type Feed struct {
	Title       string
	SiteURL     string
	FeedURL     string
	Description string
	Items       []Item
}

// Item is the normalized representation of a single feed entry.
type Item struct {
	Title       string
	Link        string
	Description string
	Content     string
	Author      string
	PublishedAt string
	Tags        []string
	Enclosures  []Enclosure
}

// Enclosure represents a media attachment on an item.
type Enclosure struct {
	URL      string
	MimeType string
	Size     int64
}

// Parse detects the feed format from body and contentType, then delegates
// to the appropriate parser.
func Parse(body []byte, contentType string) (*Feed, error) {
	ct := strings.ToLower(contentType)
	trimmed := strings.TrimSpace(string(body))

	if strings.Contains(ct, "json") || strings.HasPrefix(trimmed, "{") {
		return parseJSONFeed(body)
	}

	return parseXMLFeed(body)
}

func parseXMLFeed(body []byte) (*Feed, error) {
	var generic struct {
		XMLName xml.Name `xml:"rss"`
		Channel struct {
			Title       string   `xml:"title"`
			Links       []string `xml:"link"`
			Description string   `xml:"description"`
			Items       []struct {
				Title       string   `xml:"title"`
				Links       []string `xml:"link"`
				Description string   `xml:"description"`
				Content     string   `xml:"encoded"`
				PubDate     string   `xml:"pubDate"`
				Author      string   `xml:"author"`
				Categories  []string `xml:"category"`
				Enclosures  []struct {
					URL  string `xml:"url,attr"`
					Type string `xml:"type,attr"`
					Length int64 `xml:"length,attr"`
				} `xml:"enclosure"`
			} `xml:"item"`
		} `xml:"channel"`
	}

	// Try RSS 2.0 first
	if err := xml.Unmarshal(body, &generic); err == nil && generic.XMLName.Local == "rss" {
		feed := &Feed{
			Title:       generic.Channel.Title,
			SiteURL:     firstNonEmpty(generic.Channel.Links),
			Description: generic.Channel.Description,
		}
		for _, item := range generic.Channel.Items {
			feed.Items = append(feed.Items, Item{
				Title:       item.Title,
				Link:        firstNonEmpty(item.Links),
				Description: item.Description,
				Content:     item.Content,
				Author:      item.Author,
				PublishedAt: item.PubDate,
				Tags:        item.Categories,
			})
		}
		return feed, nil
	}

	// Fallback: try Atom
	return parseAtom(body)
}

// firstNonEmpty returns the first non-empty string in s, or "" if none.
// This handles RSS feeds where <atom:link> (a self-closing element with
// no text content) appears alongside <link> and overrides it in Go's
// encoding/xml, which matches by local name regardless of namespace.
func firstNonEmpty(s []string) string {
	for _, v := range s {
		if v != "" {
			return v
		}
	}
	return ""
}

func parseAtom(body []byte) (*Feed, error) {
	var atom struct {
		XMLName xml.Name `xml:"http://www.w3.org/2005/Atom feed"`
		Title   string   `xml:"title"`
		Links   []struct {
			Href string `xml:"href,attr"`
			Rel  string `xml:"rel,attr"`
			Type string `xml:"type,attr"`
		} `xml:"link"`
		Entries []struct {
			Title     string `xml:"title"`
			Links     []struct {
				Href string `xml:"href,attr"`
				Rel  string `xml:"rel,attr"`
			} `xml:"link"`
			Summary   string `xml:"summary"`
			Content   string `xml:"content"`
			Published string `xml:"published"`
			Updated   string `xml:"updated"`
			Author    struct {
				Name string `xml:"name"`
			} `xml:"author"`
			Categories []struct {
				Term string `xml:"term,attr"`
			} `xml:"category"`
		} `xml:"entry"`
	}

	if err := xml.Unmarshal(body, &atom); err != nil {
		return nil, fmt.Errorf("parse atom/rss: %w", err)
	}

	feed := &Feed{Title: atom.Title}
	for _, link := range atom.Links {
		if link.Rel == "" || link.Rel == "alternate" {
			feed.SiteURL = link.Href
			break
		}
	}

	for _, entry := range atom.Entries {
		item := Item{
			Title:       entry.Title,
			Description: entry.Summary,
			Content:     entry.Content,
			PublishedAt: entry.Published,
			Author:      entry.Author.Name,
		}
		if item.PublishedAt == "" {
			item.PublishedAt = entry.Updated
		}
		for _, link := range entry.Links {
			if link.Rel == "" || link.Rel == "alternate" {
				item.Link = link.Href
				break
			}
		}
		for _, cat := range entry.Categories {
			item.Tags = append(item.Tags, cat.Term)
		}
		feed.Items = append(feed.Items, item)
	}

	return feed, nil
}

func parseJSONFeed(body []byte) (*Feed, error) {
	var jf struct {
		Title       string `json:"title"`
		HomePageURL string `json:"home_page_url"`
		FeedURL     string `json:"feed_url"`
		Description string `json:"description"`
		Items       []struct {
			Title         string `json:"title"`
			URL           string `json:"url"`
			ContentHTML   string `json:"content_html"`
			ContentText   string `json:"content_text"`
			Summary       string `json:"summary"`
			DatePublished string `json:"date_published"`
			Authors       []struct {
				Name string `json:"name"`
			} `json:"authors"`
			Tags       []string `json:"tags"`
			Enclosures []struct {
				URL      string `json:"url"`
				MimeType string `json:"mime_type"`
				Size     int64   `json:"size"`
			} `json:"attachments"`
		} `json:"items"`
	}

	if err := json.Unmarshal(body, &jf); err != nil {
		return nil, fmt.Errorf("parse json feed: %w", err)
	}

	feed := &Feed{
		Title:       jf.Title,
		SiteURL:     jf.HomePageURL,
		FeedURL:     jf.FeedURL,
		Description: jf.Description,
	}

	for _, item := range jf.Items {
		content := item.ContentHTML
		if content == "" {
			content = item.ContentText
		}
		author := ""
		if len(item.Authors) > 0 {
			author = item.Authors[0].Name
		}
		var encs []Enclosure
		for _, enc := range item.Enclosures {
			encs = append(encs, Enclosure{
				URL:      enc.URL,
				MimeType: enc.MimeType,
				Size:     enc.Size,
			})
		}
		feed.Items = append(feed.Items, Item{
			Title:       item.Title,
			Link:        item.URL,
			Description: item.Summary,
			Content:     content,
			Author:      author,
			PublishedAt: item.DatePublished,
			Tags:        item.Tags,
			Enclosures:  encs,
		})
	}

	return feed, nil
}
