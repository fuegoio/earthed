// Package sanitizer provides allowlist-based HTML sanitization for feed content,
// stripping XSS vectors while preserving safe formatting tags.
package sanitizer

import (
	"bytes"
	"fmt"

	"github.com/PuerkitoBio/goquery"
)

// allowedTags is the set of HTML elements preserved in sanitized output.
var allowedTags = map[string]bool{
	"a":          true,
	"abbr":       true,
	"b":          true,
	"blockquote": true,
	"br":         true,
	"cite":       true,
	"code":       true,
	"dd":         true,
	"del":        true,
	"div":        true,
	"dl":         true,
	"dt":         true,
	"em":         true,
	"figcaption": true,
	"figure":     true,
	"h1":         true,
	"h2":         true,
	"h3":         true,
	"h4":         true,
	"h5":         true,
	"h6":         true,
	"hr":         true,
	"i":          true,
	"img":        true,
	"ins":        true,
	"kbd":        true,
	"li":         true,
	"mark":       true,
	"ol":         true,
	"p":          true,
	"pre":        true,
	"q":          true,
	"s":          true,
	"samp":       true,
	"small":      true,
	"span":       true,
	"strong":     true,
	"sub":        true,
	"sup":        true,
	"table":      true,
	"tbody":      true,
	"td":         true,
	"tfoot":      true,
	"th":         true,
	"thead":      true,
	"tr":         true,
	"u":          true,
	"ul":         true,
	"var":        true,
	"video":      true,
	"source":     true,
}

// allowedAttributes lists the attributes preserved per tag.
var allowedAttributes = map[string]map[string]bool{
	"a":   {"href": true, "title": true, "rel": true},
	"img": {"src": true, "alt": true, "title": true, "width": true, "height": true},
	"video":      {"src": true, "width": true, "height": true, "controls": true},
	"source":     {"src": true, "type": true},
}

// Sanitize strips disallowed tags and attributes from an HTML string.
func Sanitize(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(html)))
	if err != nil {
		return "", fmt.Errorf("parse html: %w", err)
	}

	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		if !allowedTags[tag] {
			s.Remove()
			return
		}
		attrs, ok := allowedAttributes[tag]
		if !ok {
			s.RemoveFiltered("*")
			return
		}
		for _, attr := range s.Nodes[0].Attr {
			if !attrs[attr.Key] {
				s.RemoveAttr(attr.Key)
			}
		}
		if tag == "a" {
			href, exists := s.Attr("href")
			if exists {
				s.SetAttr("href", sanitizeURL(href))
				s.SetAttr("rel", "noopener noreferrer")
			}
		}
	})

	out, err := doc.Find("body").Html()
	if err != nil {
		return "", fmt.Errorf("render html: %w", err)
	}
	return out, nil
}

func sanitizeURL(u string) string {
	// Block javascript: and data: URLs
	lower := ""
	for i, c := range u {
		if i >= 20 {
			break
		}
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		lower += string(c)
	}
	if len(lower) >= 11 && lower[:11] == "javascript:" {
		return ""
	}
	if len(lower) >= 5 && lower[:5] == "data:" {
		return ""
	}
	return u
}
