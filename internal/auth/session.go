package auth

import (
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

// Cookie names.
const (
	adminCookie = "adm"
	userCookie  = "usr"
	stateCookie = "oauthstate"
)

// User is a commenter's GitHub identity, persisted in a signed cookie.
type User struct {
	ID     int64
	Login  string
	Avatar string
}

// Manager encodes/decodes signed, encrypted cookies for admin and commenter
// sessions. It is safe for concurrent use.
type Manager struct {
	sc     *securecookie.SecureCookie
	secure bool // set Secure flag on cookies (false in dev over http)
}

// NewManager builds a session manager from hash/block keys.
func NewManager(hashKey, blockKey []byte, secure bool) *Manager {
	sc := securecookie.New(hashKey, blockKey)
	sc.MaxAge(int((30 * 24 * time.Hour).Seconds()))
	return &Manager{sc: sc, secure: secure}
}

func (m *Manager) set(w http.ResponseWriter, name string, value any, maxAge int) {
	encoded, err := m.sc.Encode(name, value)
	if err != nil {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    encoded,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (m *Manager) clear(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: m.secure, SameSite: http.SameSiteLaxMode,
	})
}

const sessionMaxAge = 30 * 24 * 60 * 60 // 30 days

// SetAdmin marks the response's session as authenticated admin.
func (m *Manager) SetAdmin(w http.ResponseWriter) { m.set(w, adminCookie, true, sessionMaxAge) }

// IsAdmin reports whether the request carries a valid admin session.
func (m *Manager) IsAdmin(r *http.Request) bool {
	c, err := r.Cookie(adminCookie)
	if err != nil {
		return false
	}
	var ok bool
	return m.sc.Decode(adminCookie, c.Value, &ok) == nil && ok
}

// ClearAdmin logs the admin out.
func (m *Manager) ClearAdmin(w http.ResponseWriter) { m.clear(w, adminCookie) }

// SetUser persists a commenter identity.
func (m *Manager) SetUser(w http.ResponseWriter, u User) { m.set(w, userCookie, u, sessionMaxAge) }

// CurrentUser returns the signed-in commenter, or nil.
func (m *Manager) CurrentUser(r *http.Request) *User {
	c, err := r.Cookie(userCookie)
	if err != nil {
		return nil
	}
	var u User
	if m.sc.Decode(userCookie, c.Value, &u) != nil {
		return nil
	}
	return &u
}

// ClearUser signs the commenter out.
func (m *Manager) ClearUser(w http.ResponseWriter) { m.clear(w, userCookie) }

// SetState stores a short-lived OAuth CSRF state and the post-login redirect.
func (m *Manager) SetState(w http.ResponseWriter, state, returnTo string) {
	m.set(w, stateCookie, []string{state, returnTo}, 600)
}

// State returns the stored (state, returnTo) pair and clears it.
func (m *Manager) State(w http.ResponseWriter, r *http.Request) (state, returnTo string) {
	c, err := r.Cookie(stateCookie)
	if err != nil {
		return "", ""
	}
	var pair []string
	if m.sc.Decode(stateCookie, c.Value, &pair) != nil || len(pair) != 2 {
		return "", ""
	}
	m.clear(w, stateCookie)
	return pair[0], pair[1]
}
