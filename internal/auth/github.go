package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// GitHub wraps the OAuth2 config for the commenter login flow.
type GitHub struct {
	cfg *oauth2.Config
}

// NewGitHub builds the OAuth client. redirectURL must match the app's
// registered callback, e.g. https://aniicrite.dev/auth/github/callback.
func NewGitHub(clientID, clientSecret, redirectURL string) *GitHub {
	return &GitHub{cfg: &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"read:user"},
		Endpoint:     github.Endpoint,
	}}
}

// AuthCodeURL returns the URL to redirect the user to, for the given state.
func (g *GitHub) AuthCodeURL(state string) string {
	return g.cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// Exchange trades an OAuth code for the authenticated GitHub user.
func (g *GitHub) Exchange(ctx context.Context, code string) (User, error) {
	tok, err := g.cfg.Exchange(ctx, code)
	if err != nil {
		return User{}, fmt.Errorf("oauth exchange: %w", err)
	}
	client := g.cfg.Client(ctx, tok)
	client.Timeout = 10 * time.Second

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return User{}, fmt.Errorf("fetch github user: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return User{}, fmt.Errorf("github user endpoint: status %d", resp.StatusCode)
	}
	var payload struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return User{}, fmt.Errorf("decode github user: %w", err)
	}
	if payload.ID == 0 || payload.Login == "" {
		return User{}, fmt.Errorf("github user missing id/login")
	}
	return User{ID: payload.ID, Login: payload.Login, Avatar: payload.AvatarURL}, nil
}

// RandomState returns a random hex string for CSRF protection.
func RandomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
