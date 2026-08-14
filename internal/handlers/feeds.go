package handlers

import (
	"net/http"

	"aniicrite.dev/internal/rss"
)

// Feed serves the RSS 2.0 feed at /index.xml.
func (h *Handler) Feed(w http.ResponseWriter, r *http.Request) {
	posts, _, err := h.store.ListPosts(r.Context(), 50, 0)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	body, err := rss.Feed(h.cfg.BaseURL, h.site.Title, h.site.Bio, h.site.Author, posts)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	_, _ = w.Write(body)
}

// Sitemap serves sitemap.xml.
func (h *Handler) Sitemap(w http.ResponseWriter, r *http.Request) {
	posts, _, err := h.store.ListPosts(r.Context(), 1000, 0)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	tags, err := h.store.AllTags(r.Context())
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	body, err := rss.Sitemap(h.cfg.BaseURL, posts, tags)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = w.Write(body)
}

// Robots serves a permissive robots.txt pointing at the sitemap.
func (h *Handler) Robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("User-agent: *\nAllow: /\nDisallow: /admin\nSitemap: " + h.cfg.BaseURL + "/sitemap.xml\n"))
}
