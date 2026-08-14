package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"aniicrite.dev/internal/auth"
	"aniicrite.dev/internal/markdown"
	"aniicrite.dev/internal/models"
	"aniicrite.dev/web/templates"
)

// RequireAdmin gates a route behind a valid admin session.
func (h *Handler) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.sess.IsAdmin(r) {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// --- Auth ---

// AdminLoginForm shows the login page.
func (h *Handler) AdminLoginForm(w http.ResponseWriter, r *http.Request) {
	if h.sess.IsAdmin(r) {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	b := h.adminBase(r, templates.Meta{Title: "Login"})
	b.Admin = false // no admin nav on the login page
	h.render(w, r, http.StatusOK, templates.AdminLogin(b, ""))
}

// AdminLogin verifies the password and starts a session.
func (h *Handler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if !h.loginLimiter.Allow(ip) {
		b := h.adminBase(r, templates.Meta{Title: "Login"})
		b.Admin = false
		h.render(w, r, http.StatusTooManyRequests, templates.AdminLogin(b, "Too many attempts. Try again later."))
		return
	}
	pw := r.FormValue("password")
	if h.cfg.AdminPasswordHash == "" || !auth.VerifyPassword(h.cfg.AdminPasswordHash, pw) {
		h.loginLimiter.Fail(ip)
		b := h.adminBase(r, templates.Meta{Title: "Login"})
		b.Admin = false
		h.render(w, r, http.StatusUnauthorized, templates.AdminLogin(b, "Incorrect password."))
		return
	}
	h.loginLimiter.Reset(ip)
	h.sess.SetAdmin(w)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// AdminLogout ends the admin session.
func (h *Handler) AdminLogout(w http.ResponseWriter, r *http.Request) {
	h.sess.ClearAdmin(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// --- Dashboard ---

// AdminDashboard shows tallies and quick actions.
func (h *Handler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	c, err := h.store.DashboardCounts(r.Context())
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	b := h.adminBase(r, templates.Meta{Title: "Dashboard"})
	h.render(w, r, http.StatusOK, templates.AdminDashboard(b, templates.Stats(c)))
}

// --- Posts ---

// AdminPosts lists all posts including drafts.
func (h *Handler) AdminPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.store.ListAllPosts(r.Context())
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	b := h.adminBase(r, templates.Meta{Title: "Posts"})
	h.render(w, r, http.StatusOK, templates.AdminPosts(b, posts))
}

// AdminNewPost shows an empty post form.
func (h *Handler) AdminNewPost(w http.ResponseWriter, r *http.Request) {
	b := h.adminBase(r, templates.Meta{Title: "New post"})
	h.render(w, r, http.StatusOK, templates.AdminPostForm(b, models.Post{Status: models.StatusDraft}, true, ""))
}

// AdminCreatePost persists a new post.
func (h *Handler) AdminCreatePost(w http.ResponseWriter, r *http.Request) {
	in := h.postInputFromForm(r)
	slug, err := h.store.UniqueSlug(r.Context(), in.Slug, 0)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	in.Slug = slug
	if _, err := h.store.CreatePost(r.Context(), in); err != nil {
		h.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
}

// AdminEditPost shows the edit form.
func (h *Handler) AdminEditPost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	p, err := h.store.PostByID(r.Context(), id)
	if err == models.ErrNotFound {
		h.notFound(w, r)
		return
	}
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	b := h.adminBase(r, templates.Meta{Title: "Edit post"})
	h.render(w, r, http.StatusOK, templates.AdminPostForm(b, p, false, ""))
}

// AdminUpdatePost saves edits.
func (h *Handler) AdminUpdatePost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	in := h.postInputFromForm(r)
	slug, err := h.store.UniqueSlug(r.Context(), in.Slug, id)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	in.Slug = slug
	if err := h.store.UpdatePost(r.Context(), id, in); err != nil {
		h.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
}

// AdminDeletePost removes a post.
func (h *Handler) AdminDeletePost(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.store.DeletePost(r.Context(), id); err != nil {
		h.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
}

func (h *Handler) postInputFromForm(r *http.Request) models.PostInput {
	title := strings.TrimSpace(r.FormValue("title"))
	slug := models.Slugify(r.FormValue("slug"))
	if slug == "" {
		slug = models.Slugify(title)
	}
	bodyMD := r.FormValue("body_md")
	bodyHTML, _ := markdown.Render(bodyMD)
	summary := strings.TrimSpace(r.FormValue("summary"))
	if summary == "" {
		summary = markdown.Summary(bodyHTML, 200)
	}
	status := models.StatusDraft
	if r.FormValue("status") == models.StatusPublished {
		status = models.StatusPublished
	}
	return models.PostInput{
		Slug:       slug,
		Title:      title,
		Summary:    summary,
		BodyMD:     bodyMD,
		BodyHTML:   bodyHTML,
		CoverImage: strings.TrimSpace(r.FormValue("cover_image")),
		Status:     status,
		Reading:    markdown.ReadingMinutes(bodyMD),
		Tags:       models.ParseTags(r.FormValue("tags")),
	}
}

// AdminPreview renders a Markdown fragment to HTML for the live editor.
func (h *Handler) AdminPreview(w http.ResponseWriter, r *http.Request) {
	html, _ := markdown.Render(r.FormValue("body_md"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, html)
}

// AdminUpload saves an uploaded image and returns a Markdown snippet.
func (h *Handler) AdminUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(16 << 20); err != nil { // 16 MiB
		http.Error(w, "upload too large", http.StatusRequestEntityTooLarge)
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "no file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	// SVG is intentionally excluded: it can carry inline scripts (stored XSS).
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
	default:
		http.Error(w, "unsupported image type", http.StatusUnsupportedMediaType)
		return
	}

	// Sniff the real content type from the first bytes and require an image,
	// so a renamed non-image can't be stored.
	sniff := make([]byte, 512)
	n, _ := io.ReadFull(file, sniff)
	if !strings.HasPrefix(http.DetectContentType(sniff[:n]), "image/") {
		http.Error(w, "file is not a valid image", http.StatusUnsupportedMediaType)
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		h.serverError(w, r, err)
		return
	}
	name := randomHex(8) + ext
	dst, err := os.Create(filepath.Join(h.uploadDir, name))
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		h.serverError(w, r, err)
		return
	}
	url := "/uploads/" + name
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = templates.UploadResult(url).Render(r.Context(), w)
}

// --- Projects ---

// AdminProjects lists projects.
func (h *Handler) AdminProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.store.ListProjects(r.Context(), false)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	b := h.adminBase(r, templates.Meta{Title: "Projects"})
	h.render(w, r, http.StatusOK, templates.AdminProjects(b, projects))
}

// AdminNewProject shows an empty project form.
func (h *Handler) AdminNewProject(w http.ResponseWriter, r *http.Request) {
	b := h.adminBase(r, templates.Meta{Title: "New project"})
	h.render(w, r, http.StatusOK, templates.AdminProjectForm(b, models.Project{Status: models.StatusPublished}, true, ""))
}

// AdminCreateProject persists a new project.
func (h *Handler) AdminCreateProject(w http.ResponseWriter, r *http.Request) {
	in := h.projectInputFromForm(r)
	slug, err := h.uniqueProjectSlug(r, in.Slug, 0)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	in.Slug = slug
	if _, err := h.store.CreateProject(r.Context(), in); err != nil {
		h.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/admin/projects", http.StatusSeeOther)
}

// AdminEditProject shows the edit form.
func (h *Handler) AdminEditProject(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	p, err := h.store.ProjectByID(r.Context(), id)
	if err == models.ErrNotFound {
		h.notFound(w, r)
		return
	}
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	b := h.adminBase(r, templates.Meta{Title: "Edit project"})
	h.render(w, r, http.StatusOK, templates.AdminProjectForm(b, p, false, ""))
}

// AdminUpdateProject saves project edits.
func (h *Handler) AdminUpdateProject(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	in := h.projectInputFromForm(r)
	slug, err := h.uniqueProjectSlug(r, in.Slug, id)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	in.Slug = slug
	if err := h.store.UpdateProject(r.Context(), id, in); err != nil {
		h.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/admin/projects", http.StatusSeeOther)
}

// AdminDeleteProject removes a project.
func (h *Handler) AdminDeleteProject(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.store.DeleteProject(r.Context(), id); err != nil {
		h.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/admin/projects", http.StatusSeeOther)
}

func (h *Handler) projectInputFromForm(r *http.Request) models.ProjectInput {
	title := strings.TrimSpace(r.FormValue("title"))
	slug := models.Slugify(r.FormValue("slug"))
	if slug == "" {
		slug = models.Slugify(title)
	}
	descMD := r.FormValue("desc_md")
	descHTML, _ := markdown.Render(descMD)
	sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))
	status := models.StatusPublished
	if r.FormValue("status") == models.StatusDraft {
		status = models.StatusDraft
	}
	return models.ProjectInput{
		Slug:      slug,
		Title:     title,
		DescMD:    descMD,
		DescHTML:  descHTML,
		URL:       strings.TrimSpace(r.FormValue("url")),
		RepoURL:   strings.TrimSpace(r.FormValue("repo_url")),
		SortOrder: sortOrder,
		Status:    status,
		Tags:      models.ParseTags(r.FormValue("tags")),
	}
}

func (h *Handler) uniqueProjectSlug(r *http.Request, base string, excludeID int64) (string, error) {
	slug := base
	for i := 2; ; i++ {
		exists, err := h.store.ProjectSlugExists(r.Context(), slug, excludeID)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = base + "-" + strconv.Itoa(i)
	}
}

// --- Comments ---

// AdminComments lists all comments.
func (h *Handler) AdminComments(w http.ResponseWriter, r *http.Request) {
	rows, err := h.store.AllCommentsWithPost(r.Context())
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	view := make([]templates.AdminComment, len(rows))
	for i, c := range rows {
		view[i] = templates.AdminComment{Comment: c.Comment, PostSlug: c.PostSlug, PostTitle: c.PostTitle}
	}
	b := h.adminBase(r, templates.Meta{Title: "Comments"})
	h.render(w, r, http.StatusOK, templates.AdminComments(b, view))
}

// AdminDeleteComment removes a single comment.
func (h *Handler) AdminDeleteComment(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.store.DeleteComment(r.Context(), id); err != nil {
		h.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/admin/comments", http.StatusSeeOther)
}

// AdminBlockUser bans the comment's author and deletes their comments.
func (h *Handler) AdminBlockUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.store.BlockUser(r.Context(), id); err != nil {
		h.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/admin/comments", http.StatusSeeOther)
}

// adminBase builds a layout context with the admin chrome enabled.
func (h *Handler) adminBase(r *http.Request, meta templates.Meta) templates.Base {
	b := h.base(r, "", meta)
	b.Admin = true
	return b
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
