package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"aniicrite.dev/internal/models"
	"aniicrite.dev/web/templates"
)

const postsPerPage = 10

// Home renders the landing page with recent posts.
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	posts, _, err := h.store.ListPosts(r.Context(), 5, 0)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	b := h.base(r, "home", templates.Meta{Canonical: h.cfg.BaseURL + "/"})
	h.render(w, r, http.StatusOK, templates.Home(b, posts))
}

// Posts renders the paginated post list.
func (h *Handler) Posts(w http.ResponseWriter, r *http.Request) {
	page := 1
	if v, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && v > 1 {
		page = v
	}
	posts, total, err := h.store.ListPosts(r.Context(), postsPerPage, (page-1)*postsPerPage)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	totalPages := (total + postsPerPage - 1) / postsPerPage
	if totalPages < 1 {
		totalPages = 1
	}
	b := h.base(r, "posts", templates.Meta{Title: "Posts"})
	h.render(w, r, http.StatusOK, templates.PostsPage(b, posts, page, totalPages))
}

// Post renders a single post and increments its view count once per session.
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	admin := h.sess.IsAdmin(r)
	p, err := h.store.PostBySlug(r.Context(), slug, admin)
	if err == models.ErrNotFound {
		h.notFound(w, r)
		return
	}
	if err != nil {
		h.serverError(w, r, err)
		return
	}

	if !admin && h.markViewed(w, r, p.ID) {
		if err := h.store.IncrementViews(r.Context(), p.ID); err == nil {
			p.ViewCount++
		}
	}

	var comments []models.Comment
	if h.site.CommentsEnabled {
		if comments, err = h.store.CommentsForPost(r.Context(), p.ID); err != nil {
			h.serverError(w, r, err)
			return
		}
	}

	meta := templates.Meta{
		Title:       p.Title,
		Description: p.Summary,
		Article:     true,
		Canonical:   h.cfg.BaseURL + "/posts/" + p.Slug + "/",
		OGImage:     h.absoluteURL(p.CoverImage),
		Published:   p.Date(),
		Modified:    p.UpdatedAt,
		Tags:        tagNames(p.Tags),
	}
	b := h.base(r, "posts", meta)
	h.render(w, r, http.StatusOK, templates.PostPage(b, p, comments, h.loginURL(r)))
}

// Projects renders the projects page.
func (h *Handler) Projects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.store.ListProjects(r.Context(), true)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	b := h.base(r, "projects", templates.Meta{Title: "Projects"})
	h.render(w, r, http.StatusOK, templates.ProjectsPage(b, projects))
}

// About renders the static about page.
func (h *Handler) About(w http.ResponseWriter, r *http.Request) {
	b := h.base(r, "about", templates.Meta{Title: "About"})
	h.render(w, r, http.StatusOK, templates.AboutPage(b, h.aboutHTML))
}

// Search returns an HTMX fragment of full-text results.
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	var results []models.SearchResult
	if q != "" {
		var err error
		if results, err = h.store.Search(r.Context(), q, 15); err != nil {
			h.serverError(w, r, err)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = templates.SearchResults(results, q).Render(r.Context(), w)
}

// absoluteURL turns a possibly-relative asset path into an absolute URL.
func (h *Handler) absoluteURL(u string) string { return absURL(h.cfg.BaseURL, u) }

// absURL joins base and u unless u is empty or already absolute.
func absURL(base, u string) string {
	if u == "" {
		return ""
	}
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return u
	}
	return base + u
}

// tagNames extracts tag display names.
func tagNames(tags []models.Tag) []string {
	out := make([]string, len(tags))
	for i, t := range tags {
		out[i] = t.Name
	}
	return out
}

// TagsIndex lists every tag that labels a published post.
func (h *Handler) TagsIndex(w http.ResponseWriter, r *http.Request) {
	tags, err := h.store.AllTags(r.Context())
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	b := h.base(r, "", templates.Meta{Title: "Tags"})
	h.render(w, r, http.StatusOK, templates.TagsIndex(b, tags))
}

// TagPage lists published posts carrying a tag.
func (h *Handler) TagPage(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	tag, err := h.store.TagBySlug(r.Context(), slug)
	if err == models.ErrNotFound {
		h.notFound(w, r)
		return
	}
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	posts, err := h.store.PostsByTag(r.Context(), slug)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	meta := templates.Meta{
		Title:       "#" + tag.Name,
		Description: "Posts tagged " + tag.Name + ".",
	}
	b := h.base(r, "", meta)
	h.render(w, r, http.StatusOK, templates.TagPage(b, tag, posts))
}

// Health is a liveness probe that also verifies the database is reachable.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Ping(r.Context()); err != nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
