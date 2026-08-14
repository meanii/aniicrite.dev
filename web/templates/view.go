// Package templates holds the templ components and their view models.
package templates

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"aniicrite.dev/internal/auth"
	"aniicrite.dev/internal/models"
)

func postFormAction(p models.Post, isNew bool) string {
	if isNew {
		return "/admin/posts"
	}
	return fmt.Sprintf("/admin/posts/%d", p.ID)
}

func projectFormAction(pr models.Project, isNew bool) string {
	if isNew {
		return "/admin/projects"
	}
	return fmt.Sprintf("/admin/projects/%d", pr.ID)
}

// Social is a labeled profile link.
type Social struct {
	Label string
	URL   string
}

// Site carries process-wide identity shared by every page.
type Site struct {
	Title           string
	Author          string
	Bio             string
	BaseURL         string
	AvatarURL       string
	DefaultOGImage  string // absolute URL used when a page has no image
	Socials         []Social
	CommentsEnabled bool
}

// Meta is the per-page <head> metadata.
type Meta struct {
	Title       string // page title without the site suffix ("" = home)
	Description string
	Canonical   string    // absolute URL
	OGImage     string    // absolute URL, optional (falls back to site default)
	Article     bool      // og:type article vs website
	Published   time.Time // article: publish time (zero = omit)
	Modified    time.Time // article: last-modified time (zero = omit)
	Tags        []string  // article: keywords
}

// Base is the layout context threaded into every page.
type Base struct {
	Site  Site
	Meta  Meta
	Nav   string // active nav key: home|posts|projects|about
	Admin bool
	User  *auth.User
}

// Stats populates the admin dashboard.
type Stats struct {
	Posts    int
	Drafts   int
	Projects int
	Comments int
}

// AdminComment is a comment plus the post it belongs to, for the admin list.
type AdminComment struct {
	models.Comment
	PostSlug  string
	PostTitle string
}

// FullTitle is the <title> text.
func (b Base) FullTitle() string {
	if b.Meta.Title == "" || b.Meta.Title == b.Site.Title {
		return b.Site.Title
	}
	return b.Meta.Title + " · " + b.Site.Title
}

// Description falls back to the site bio.
func (b Base) Description() string {
	if b.Meta.Description != "" {
		return b.Meta.Description
	}
	return b.Site.Bio
}

// OGType returns the OpenGraph object type.
func (b Base) OGType() string {
	if b.Meta.Article {
		return "article"
	}
	return "website"
}

func isActive(current, key string) string {
	if current == key {
		return "active"
	}
	return ""
}

// fmtDate renders a display date like "2 Jan 2006".
func fmtDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2 Jan 2006")
}

// rfc3339 renders a machine date for <time datetime>.
func rfc3339(t time.Time) string { return t.UTC().Format(time.RFC3339) }

// tagList renders tags as a comma string for form fields.
func tagList(tags []models.Tag) string {
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	return strings.Join(names, ", ")
}

// sameAs returns the http(s) social URLs for JSON-LD (mailto: excluded).
func (b Base) sameAs() []string {
	var out []string
	for _, s := range b.Site.Socials {
		if strings.HasPrefix(s.URL, "http://") || strings.HasPrefix(s.URL, "https://") {
			out = append(out, s.URL)
		}
	}
	return out
}

// OGImageURL returns the page's share image, falling back to the site default.
func (b Base) OGImageURL() string {
	if b.Meta.OGImage != "" {
		return b.Meta.OGImage
	}
	return b.Site.DefaultOGImage
}

// PublishedISO / ModifiedISO render article timestamps, or "" when unset.
func (b Base) PublishedISO() string { return isoOrEmpty(b.Meta.Published) }
func (b Base) ModifiedISO() string  { return isoOrEmpty(b.Meta.Modified) }

func isoOrEmpty(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// LDJSONScript returns the full <script> element for the page's JSON-LD.
// Emitted via templ.Raw at head level because templ does not evaluate
// expressions inside a literal <script> element.
func (b Base) LDJSONScript() string {
	return `<script type="application/ld+json">` + b.LDJSON() + `</script>`
}

// LDJSON builds the schema.org JSON-LD document for the current page.
// Output is HTML-escaped by encoding/json, so it is safe inside <script>.
func (b Base) LDJSON() string {
	var doc any
	switch {
	case b.Meta.Article:
		article := map[string]any{
			"@context":         "https://schema.org",
			"@type":            "BlogPosting",
			"headline":         b.Meta.Title,
			"description":      b.Description(),
			"url":              b.Meta.Canonical,
			"mainEntityOfPage": b.Meta.Canonical,
			"image":            b.OGImageURL(),
			"author":           map[string]any{"@type": "Person", "name": b.Site.Author, "url": b.Site.BaseURL},
			"publisher":        map[string]any{"@type": "Person", "name": b.Site.Author},
		}
		if s := b.PublishedISO(); s != "" {
			article["datePublished"] = s
		}
		if s := b.ModifiedISO(); s != "" {
			article["dateModified"] = s
		}
		if len(b.Meta.Tags) > 0 {
			article["keywords"] = strings.Join(b.Meta.Tags, ", ")
		}
		doc = article
	case b.Nav == "home":
		doc = map[string]any{
			"@context": "https://schema.org",
			"@graph": []any{
				map[string]any{"@type": "WebSite", "@id": b.Site.BaseURL + "#website", "name": b.Site.Title, "url": b.Site.BaseURL},
				map[string]any{"@type": "Person", "name": b.Site.Author, "url": b.Site.BaseURL, "description": b.Site.Bio, "sameAs": b.sameAs()},
			},
		}
	default:
		doc = map[string]any{
			"@context": "https://schema.org",
			"@type":    "WebSite",
			"name":     b.Site.Title,
			"url":      b.Site.BaseURL,
		}
	}
	out, err := json.Marshal(doc)
	if err != nil {
		return ""
	}
	return string(out)
}
