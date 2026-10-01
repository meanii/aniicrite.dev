// Package rss builds the Atom-less RSS 2.0 feed and the XML sitemap.
package rss

import (
	"encoding/xml"
	"strings"
	"time"

	"aniicrite.dev/internal/models"
)

// Feed renders an RSS 2.0 document for the given published posts.
func Feed(baseURL, title, description, author string, posts []models.Post) ([]byte, error) {
	type encoded struct {
		Text string `xml:",cdata"`
	}
	type item struct {
		Title   string   `xml:"title"`
		Link    string   `xml:"link"`
		GUID    string   `xml:"guid"`
		PubDate string   `xml:"pubDate"`
		Desc    string   `xml:"description"`
		Content *encoded `xml:"content:encoded,omitempty"`
	}
	type channel struct {
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		Description string `xml:"description"`
		Generator   string `xml:"generator"`
		Items       []item `xml:"item"`
	}
	type rss struct {
		XMLName   xml.Name `xml:"rss"`
		Version   string   `xml:"version,attr"`
		ContentNS string   `xml:"xmlns:content,attr"`
		Channel   channel  `xml:"channel"`
	}

	ch := channel{Title: title, Link: baseURL + "/", Description: description, Generator: "aniicrite.dev (Go)"}
	for _, p := range posts {
		link := baseURL + "/posts/" + p.Slug + "/"
		it := item{
			Title:   p.Title,
			Link:    link,
			GUID:    link,
			PubDate: p.Date().UTC().Format(time.RFC1123Z),
			Desc:    p.Summary,
		}
		if p.BodyHTML != "" {
			it.Content = &encoded{Text: absolutize(baseURL, p.BodyHTML)}
		}
		ch.Items = append(ch.Items, it)
	}
	body, err := xml.MarshalIndent(rss{Version: "2.0", ContentNS: "http://purl.org/rss/1.0/modules/content/", Channel: ch}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), body...), nil
}

// absolutize rewrites root-relative href/src attributes so feed readers,
// which have no base URL, can follow links and load images.
func absolutize(baseURL, html string) string {
	r := strings.NewReplacer(`href="/`, `href="`+baseURL+`/`, `src="/`, `src="`+baseURL+`/`)
	return r.Replace(html)
}

// Sitemap renders a sitemap.xml covering the static routes, every post, and
// every tag page.
func Sitemap(baseURL string, posts []models.Post, tags []models.Tag) ([]byte, error) {
	type url struct {
		Loc     string `xml:"loc"`
		LastMod string `xml:"lastmod,omitempty"`
	}
	type urlset struct {
		XMLName xml.Name `xml:"urlset"`
		NS      string   `xml:"xmlns,attr"`
		URLs    []url    `xml:"url"`
	}

	set := urlset{NS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	for _, p := range []string{"/", "/posts/", "/projects/", "/about/", "/tags/"} {
		set.URLs = append(set.URLs, url{Loc: baseURL + p})
	}
	for _, p := range posts {
		set.URLs = append(set.URLs, url{
			Loc:     baseURL + "/posts/" + p.Slug + "/",
			LastMod: p.UpdatedAt.UTC().Format("2006-01-02"),
		})
	}
	for _, t := range tags {
		set.URLs = append(set.URLs, url{Loc: baseURL + "/tags/" + t.Slug + "/"})
	}
	body, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), body...), nil
}

// TrimBase normalizes a base URL (no trailing slash).
func TrimBase(u string) string { return strings.TrimRight(u, "/") }
