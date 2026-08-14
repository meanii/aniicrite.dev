// Package handlers wires HTTP routes to the store, auth, and templates.
package handlers

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/a-h/templ"

	"aniicrite.dev/internal/auth"
	"aniicrite.dev/internal/config"
	"aniicrite.dev/internal/models"
	"aniicrite.dev/internal/ratelimit"
	"aniicrite.dev/web/templates"
)

// Handler holds the dependencies shared by every route.
type Handler struct {
	cfg          *config.Config
	store        *models.Store
	sess         *auth.Manager
	gh           *auth.GitHub // nil when comments are disabled
	site         templates.Site
	aboutHTML    string
	uploadDir    string
	loginLimiter *ratelimit.Limiter
}

// New builds a Handler.
func New(cfg *config.Config, store *models.Store, sess *auth.Manager, gh *auth.GitHub, aboutHTML, uploadDir string) *Handler {
	return &Handler{
		cfg:   cfg,
		store: store,
		sess:  sess,
		gh:    gh,
		site: templates.Site{
			Title:           cfg.SiteTitle,
			Author:          cfg.AuthorName,
			Bio:             cfg.AuthorBio,
			BaseURL:         cfg.BaseURL,
			AvatarURL:       cfg.AvatarURL,
			DefaultOGImage:  absURL(cfg.BaseURL, cfg.DefaultOGImage),
			Socials:         mapSocials(cfg.Socials),
			CommentsEnabled: cfg.CommentsEnabled(),
		},
		aboutHTML:    aboutHTML,
		uploadDir:    uploadDir,
		loginLimiter: ratelimit.New(5, 15*time.Minute),
	}
}

// base assembles the layout context for a request.
func (h *Handler) base(r *http.Request, nav string, meta templates.Meta) templates.Base {
	if meta.Canonical == "" {
		meta.Canonical = h.cfg.BaseURL + r.URL.Path
	}
	return templates.Base{
		Site: h.site,
		Meta: meta,
		Nav:  nav,
		// Admin chrome is scoped to /admin/* (set by adminBase); public pages
		// always render the public nav, even for a logged-in admin.
		Admin: false,
		User:  h.sess.CurrentUser(r),
	}
}

// render writes a templ component with the given status.
func (h *Handler) render(w http.ResponseWriter, r *http.Request, status int, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := c.Render(r.Context(), w); err != nil {
		log.Printf("render: %v", err)
	}
}

func (h *Handler) notFound(w http.ResponseWriter, r *http.Request) {
	b := h.base(r, "", templates.Meta{Title: "Not found"})
	h.render(w, r, http.StatusNotFound, templates.NotFound(b))
}

func (h *Handler) serverError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("server error: %s %s: %v", r.Method, r.URL.Path, err)
	b := h.base(r, "", templates.Meta{Title: "Error"})
	h.render(w, r, http.StatusInternalServerError, templates.ServerError(b))
}

// clientIP extracts the request's client IP (RealIP middleware normalizes
// r.RemoteAddr upstream), stripping any port.
func clientIP(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// securityHeaders sets baseline security response headers. HSTS is only
// emitted outside dev (behind TLS at the Caddy edge).
func (h *Handler) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		head := w.Header()
		head.Set("X-Content-Type-Options", "nosniff")
		head.Set("X-Frame-Options", "DENY")
		head.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Inline theme script + htmx inline handlers require 'unsafe-inline';
		// posts embed remote images, so img-src is permissive over https/data.
		head.Set("Content-Security-Policy",
			"default-src 'self'; "+
				"img-src 'self' https: data:; "+
				"style-src 'self' 'unsafe-inline'; "+
				"script-src 'self' 'unsafe-inline'; "+
				"frame-ancestors 'none'; "+
				"base-uri 'self'")
		if !h.cfg.Dev {
			head.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

// mapSocials converts config socials to the template type.
func mapSocials(in []config.Social) []templates.Social {
	out := make([]templates.Social, len(in))
	for i, s := range in {
		out[i] = templates.Social{Label: s.Label, URL: s.URL}
	}
	return out
}
