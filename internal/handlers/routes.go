package handlers

import (
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Routes builds the HTTP handler tree. staticFS is served under /static/.
func (h *Handler) Routes(staticFS fs.FS) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(h.securityHeaders)
	if h.cfg.Dev {
		r.Use(middleware.Logger)
	}

	// Assets.
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(h.uploadDir))))

	// Public pages.
	r.Get("/", h.Home)
	r.Get("/posts/", h.Posts)
	r.Get("/posts/{slug}/", h.Post)
	r.Post("/posts/{slug}/comments", h.PostComment)
	r.Get("/projects/", h.Projects)
	r.Get("/about/", h.About)
	r.Get("/tags/", h.TagsIndex)
	r.Get("/tags/{slug}/", h.TagPage)
	r.Get("/search", h.Search)
	r.Get("/healthz", h.Health)

	// Feeds & SEO.
	r.Get("/index.xml", h.Feed)
	r.Get("/sitemap.xml", h.Sitemap)
	r.Get("/robots.txt", h.Robots)

	// Bare-path redirects to canonical trailing-slash URLs.
	for _, p := range []string{"/posts", "/projects", "/about", "/tags"} {
		r.Get(p, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, p+"/", http.StatusMovedPermanently)
		})
	}

	// Commenter OAuth.
	r.Get("/auth/github/login", h.GitHubLogin)
	r.Get("/auth/github/callback", h.GitHubCallback)
	r.Post("/auth/logout", h.Logout)

	// Admin.
	r.Route("/admin", func(r chi.Router) {
		r.Get("/login", h.AdminLoginForm)
		r.Post("/login", h.AdminLogin)
		r.Group(func(r chi.Router) {
			r.Use(h.RequireAdmin)
			r.Get("/", h.AdminDashboard)
			r.Post("/logout", h.AdminLogout)

			r.Get("/posts", h.AdminPosts)
			r.Get("/posts/new", h.AdminNewPost)
			r.Post("/posts", h.AdminCreatePost)
			r.Get("/posts/{id}/edit", h.AdminEditPost)
			r.Post("/posts/{id}", h.AdminUpdatePost)
			r.Post("/posts/{id}/delete", h.AdminDeletePost)
			r.Post("/preview", h.AdminPreview)
			r.Post("/upload", h.AdminUpload)

			r.Get("/projects", h.AdminProjects)
			r.Get("/projects/new", h.AdminNewProject)
			r.Post("/projects", h.AdminCreateProject)
			r.Get("/projects/{id}/edit", h.AdminEditProject)
			r.Post("/projects/{id}", h.AdminUpdateProject)
			r.Post("/projects/{id}/delete", h.AdminDeleteProject)

			r.Get("/comments", h.AdminComments)
			r.Post("/comments/{id}/delete", h.AdminDeleteComment)
			r.Post("/comments/{id}/block", h.AdminBlockUser)
		})
	})

	r.NotFound(h.notFound)
	return r
}
