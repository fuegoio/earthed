package parser

import "testing"

func TestParseMitchellHFeed(t *testing.T) {
	body := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>Mitchell Hashimoto</title>
    <link>https://mitchellh.com</link>
    <description>Mitchell Hashimoto's personal website.</description>
    <language>en-us</language>
    <atom:link href="https://mitchellh.com/feed.xml" rel="self" type="application/rss+xml"/>
    <item>
      <title>Superlogical</title>
      <link>https://mitchellh.com/writing/superlogical</link>
      <guid>https://mitchellh.com/writing/superlogical</guid>
      <pubDate>Wed, 29 Jul 2026 00:00:00 GMT</pubDate>
    </item>
  </channel>
</rss>`)

	feed, err := Parse(body, "application/rss+xml; charset=utf-8")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if feed.Title != "Mitchell Hashimoto" {
		t.Errorf("Title = %q, want %q", feed.Title, "Mitchell Hashimoto")
	}
	if feed.SiteURL != "https://mitchellh.com" {
		t.Errorf("SiteURL = %q, want %q", feed.SiteURL, "https://mitchellh.com")
	}
	if len(feed.Items) != 1 {
		t.Fatalf("Items = %d, want 1", len(feed.Items))
	}
	if feed.Items[0].Link != "https://mitchellh.com/writing/superlogical" {
		t.Errorf("Item Link = %q, want %q", feed.Items[0].Link, "https://mitchellh.com/writing/superlogical")
	}
}

func TestParseRSSAtomLinkBeforeLink(t *testing.T) {
	// Some feeds put <atom:link> before <link>. The parser should still
	// pick up the real <link> value, not the empty <atom:link>.
	body := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>Test Feed</title>
    <atom:link href="https://example.com/feed.xml" rel="self" type="application/rss+xml"/>
    <link>https://example.com</link>
    <description>Test</description>
  </channel>
</rss>`)

	feed, err := Parse(body, "application/rss+xml")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if feed.SiteURL != "https://example.com" {
		t.Errorf("SiteURL = %q, want %q", feed.SiteURL, "https://example.com")
	}
}
