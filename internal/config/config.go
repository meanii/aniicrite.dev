// Package config loads runtime configuration from the environment.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strings"
)

// Config holds all runtime settings. Secrets come from the environment; the
// site's descriptive fields have sane defaults so a bare `go run` still boots.
type Config struct {
	Addr    string // listen address, e.g. ":8080"
	BaseURL string // canonical origin, e.g. "https://aniicrite.dev"
	DataDir string // writable dir for the sqlite file + uploads
	Dev     bool   // relaxes cookie Secure flag, enables verbose errors

	// Admin auth: argon2id-encoded hash of the single admin password.
	AdminPasswordHash string

	// securecookie keys (hex). Regenerated per-process if unset (dev only).
	SessionHashKey  []byte
	SessionBlockKey []byte

	// GitHub OAuth for commenters (Phase 1). Optional; comments disabled if empty.
	GitHubClientID     string
	GitHubClientSecret string

	// Site identity, surfaced in templates + feeds.
	SiteTitle      string
	SiteDesc       string
	AuthorName     string
	AuthorBio      string
	AvatarURL      string
	DefaultOGImage string
	Socials        []Social
}

// Social is a labeled profile link shown in the footer and used for JSON-LD.
type Social struct {
	Label string
	URL   string
}

// Load reads configuration from the environment, applying defaults.
func Load() *Config {
	c := &Config{
		Addr:               env("APP_ADDR", ":8080"),
		BaseURL:            strings.TrimRight(env("BASE_URL", "http://localhost:8080"), "/"),
		DataDir:            env("DATA_DIR", "./data"),
		Dev:                env("APP_ENV", "dev") != "production",
		AdminPasswordHash:  os.Getenv("ADMIN_PASSWORD_HASH"),
		GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		SiteTitle:          env("SITE_TITLE", "My Site"),
		SiteDesc:           env("SITE_DESC", "A personal site and blog."),
		AuthorName:         env("AUTHOR_NAME", "Your Name"),
		AuthorBio:          env("AUTHOR_BIO", "Short bio — set AUTHOR_BIO to describe yourself."),
		AvatarURL:          env("AVATAR_URL", ""),
		DefaultOGImage:     env("OG_IMAGE", "/static/og-default.png"),
		Socials:            parseSocials(os.Getenv("SOCIAL_LINKS")),
	}

	c.SessionHashKey = keyOrRandom("SESSION_HASH_KEY", 32)
	c.SessionBlockKey = keyOrRandom("SESSION_BLOCK_KEY", 32)

	if c.AdminPasswordHash == "" {
		log.Println("config: ADMIN_PASSWORD_HASH unset — admin login is disabled until it is set")
	}
	return c
}

// CommentsEnabled reports whether GitHub OAuth is configured.
func (c *Config) CommentsEnabled() bool {
	return c.GitHubClientID != "" && c.GitHubClientSecret != ""
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// keyOrRandom decodes a hex key from the environment or, in its absence,
// generates an ephemeral one so dev boots without ceremony. An ephemeral key
// invalidates sessions on restart — production MUST set it.
func keyOrRandom(key string, n int) []byte {
	if v := os.Getenv(key); v != "" {
		b, err := hex.DecodeString(v)
		if err != nil {
			log.Fatalf("config: %s must be hex: %v", key, err)
		}
		return b
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("config: generating %s: %v", key, err)
	}
	log.Printf("config: %s unset — using an ephemeral key (sessions reset on restart)", key)
	return b
}

// parseSocials reads SOCIAL_LINKS, a comma-separated list of "Label|URL"
// pairs, e.g. "GitHub|https://github.com/you,Email|mailto:you@example.com".
func parseSocials(raw string) []Social {
	var out []Social
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		label, url, ok := strings.Cut(part, "|")
		label, url = strings.TrimSpace(label), strings.TrimSpace(url)
		if !ok || label == "" || url == "" {
			continue
		}
		out = append(out, Social{Label: label, URL: url})
	}
	return out
}
