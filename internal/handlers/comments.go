package handlers

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"aniicrite.dev/internal/auth"
	"aniicrite.dev/internal/models"
	"aniicrite.dev/web/templates"
)

const seenCookie = "seen"

// markViewed records a post id in the visitor's "seen" cookie, returning true
// when this is the first view (so the caller should increment the counter).
func (h *Handler) markViewed(w http.ResponseWriter, r *http.Request, postID int64) bool {
	id := strconv.FormatInt(postID, 10)
	var seen []string
	if c, err := r.Cookie(seenCookie); err == nil {
		seen = strings.Split(c.Value, ".")
		for _, s := range seen {
			if s == id {
				return false
			}
		}
	}
	seen = append(seen, id)
	if len(seen) > 200 { // cap cookie size
		seen = seen[len(seen)-200:]
	}
	http.SetCookie(w, &http.Cookie{
		Name: seenCookie, Value: strings.Join(seen, "."), Path: "/",
		MaxAge: 400 * 24 * 60 * 60, HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
	return true
}

// loginURL returns the GitHub sign-in URL that returns to the current path.
func (h *Handler) loginURL(r *http.Request) string {
	if h.gh == nil {
		return "#"
	}
	return "/auth/github/login?return=" + url.QueryEscape(r.URL.Path)
}

// PostComment stores a comment from a signed-in GitHub user. On an HTMX
// request it returns the rendered comment fragment; otherwise it redirects
// back to the post.
func (h *Handler) PostComment(w http.ResponseWriter, r *http.Request) {
	if h.gh == nil {
		http.Error(w, "comments are disabled", http.StatusNotFound)
		return
	}
	user := h.sess.CurrentUser(r)
	if user == nil {
		http.Redirect(w, r, h.loginURL(r), http.StatusSeeOther)
		return
	}
	slug := chi.URLParam(r, "slug")
	post, err := h.store.PostBySlug(r.Context(), slug, false)
	if err == models.ErrNotFound {
		h.notFound(w, r)
		return
	}
	if err != nil {
		h.serverError(w, r, err)
		return
	}

	if blocked, _ := h.store.IsBlocked(r.Context(), user.ID); blocked {
		http.Error(w, "you are not permitted to comment", http.StatusForbidden)
		return
	}

	body := strings.TrimSpace(r.FormValue("body"))
	if body == "" {
		http.Error(w, "empty comment", http.StatusBadRequest)
		return
	}
	if len([]rune(body)) > 4000 {
		body = string([]rune(body)[:4000])
	}

	in := models.CommentInput{
		PostID: post.ID, GHUserID: user.ID, GHLogin: user.Login, GHAvatar: user.Avatar, Body: body,
	}
	id, err := h.store.CreateComment(r.Context(), in)
	if err != nil {
		h.serverError(w, r, err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		c := models.Comment{ID: id, PostID: post.ID, GHUserID: user.ID, GHLogin: user.Login, GHAvatar: user.Avatar, Body: body}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = templates.CommentItem(c).Render(r.Context(), w)
		return
	}
	http.Redirect(w, r, "/posts/"+slug+"/#comments-list", http.StatusSeeOther)
}

// GitHubLogin starts the OAuth flow.
func (h *Handler) GitHubLogin(w http.ResponseWriter, r *http.Request) {
	if h.gh == nil {
		h.notFound(w, r)
		return
	}
	returnTo := r.URL.Query().Get("return")
	if !strings.HasPrefix(returnTo, "/") {
		returnTo = "/"
	}
	state := auth.RandomState()
	h.sess.SetState(w, state, returnTo)
	http.Redirect(w, r, h.gh.AuthCodeURL(state), http.StatusSeeOther)
}

// GitHubCallback completes the OAuth flow and stores the commenter identity.
func (h *Handler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	if h.gh == nil {
		h.notFound(w, r)
		return
	}
	wantState, returnTo := h.sess.State(w, r)
	if wantState == "" || r.URL.Query().Get("state") != wantState {
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}
	user, err := h.gh.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	h.sess.SetUser(w, user)
	if returnTo == "" {
		returnTo = "/"
	}
	http.Redirect(w, r, returnTo, http.StatusSeeOther)
}

// Logout clears the commenter session.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.sess.ClearUser(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
