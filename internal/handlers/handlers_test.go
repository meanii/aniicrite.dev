package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"aniicrite.dev/internal/auth"
	"aniicrite.dev/internal/config"
	"aniicrite.dev/internal/db"
	"aniicrite.dev/internal/handlers"
	"aniicrite.dev/internal/models"
)

const testPassword = "secret-password"

func setup(t *testing.T) (*httptest.Server, *models.Store) {
	t.Helper()
	dir := t.TempDir()
	sqldb, err := db.Open(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { sqldb.Close() })
	store := models.New(sqldb)

	hash, _ := auth.HashPassword(testPassword)
	cfg := &config.Config{
		Dev: true, BaseURL: "http://test", DataDir: dir,
		SiteTitle: "T", AuthorName: "A", AuthorBio: "bio",
		AdminPasswordHash: hash,
	}
	sess := auth.NewManager(bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{2}, 32), false)
	h := handlers.New(cfg, store, sess, nil, "<p>about</p>", dir)

	srv := httptest.NewServer(h.Routes(fstest.MapFS{}))
	t.Cleanup(srv.Close)
	return srv, store
}

// noRedirect returns a client that surfaces redirects instead of following.
func noRedirect() *http.Client {
	return &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
}

func TestPublicRoutesAndDraftHiding(t *testing.T) {
	srv, store := setup(t)
	ctx := context.Background()
	if _, err := store.CreatePost(ctx, models.PostInput{
		Slug: "hello-world", Title: "Hello World", BodyMD: "a searchable kumquat body",
		Status: models.StatusPublished, Tags: []string{"Go"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreatePost(ctx, models.PostInput{
		Slug: "secret", Title: "Secret", BodyMD: "hidden", Status: models.StatusDraft,
	}); err != nil {
		t.Fatal(err)
	}

	client := noRedirect()
	cases := []struct {
		path string
		want int
	}{
		{"/", 200},
		{"/posts/", 200},
		{"/projects/", 200},
		{"/about/", 200},
		{"/index.xml", 200},
		{"/sitemap.xml", 200},
		{"/posts/hello-world/", 200},
		{"/posts/secret/", 404},  // draft hidden from the public
		{"/posts/missing/", 404}, // unknown slug
		{"/posts", 301},          // bare path redirects to trailing slash
		{"/tags/", 200},          // tag index
		{"/tags/go/", 200},       // tag with a published post
		{"/tags/nope/", 404},     // unknown tag
		{"/healthz", 200},        // liveness probe
		{"/admin", 303},          // gated: redirect to login
	}
	for _, c := range cases {
		resp, err := client.Get(srv.URL + c.path)
		if err != nil {
			t.Fatalf("GET %s: %v", c.path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != c.want {
			t.Errorf("GET %s = %d, want %d", c.path, resp.StatusCode, c.want)
		}
	}

	// Security headers present on a normal response.
	resp, _ := client.Get(srv.URL + "/")
	resp.Body.Close()
	if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing X-Content-Type-Options header")
	}
	if resp.Header.Get("Content-Security-Policy") == "" {
		t.Error("missing Content-Security-Policy header")
	}

	// Full-text search returns the published post as an HTMX fragment.
	resp, _ = client.Get(srv.URL + "/search?q=kumquat")
	body := readBody(t, resp)
	if !strings.Contains(body, "hello-world") {
		t.Errorf("search fragment missing hit: %q", body)
	}

	// Comments are disabled (gh == nil): posting 404s.
	resp, _ = client.PostForm(srv.URL+"/posts/hello-world/comments", url.Values{"body": {"hi"}})
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("comment POST with comments disabled = %d, want 404", resp.StatusCode)
	}
}

func TestAdminLoginRateLimit(t *testing.T) {
	srv, _ := setup(t)
	client := noRedirect()
	// First 5 wrong attempts are rejected as 401; the 6th trips the limiter.
	for i := 1; i <= 6; i++ {
		resp, err := client.PostForm(srv.URL+"/admin/login", url.Values{"password": {"wrong"}})
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		want := http.StatusUnauthorized
		if i == 6 {
			want = http.StatusTooManyRequests
		}
		if resp.StatusCode != want {
			t.Errorf("attempt %d = %d, want %d", i, resp.StatusCode, want)
		}
	}
}

func TestAdminLoginAndCreatePost(t *testing.T) {
	srv, _ := setup(t)
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}

	// Wrong password: 401, no session.
	resp, _ := client.PostForm(srv.URL+"/admin/login", url.Values{"password": {"nope"}})
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong login = %d, want 401", resp.StatusCode)
	}

	// Correct password: 303 + admin cookie.
	resp, _ = client.PostForm(srv.URL+"/admin/login", url.Values{"password": {testPassword}})
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("correct login = %d, want 303", resp.StatusCode)
	}

	// Authenticated dashboard now reachable.
	resp, _ = client.Get(srv.URL + "/admin")
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("dashboard = %d, want 200", resp.StatusCode)
	}

	// Create a published post through the admin form.
	resp, _ = client.PostForm(srv.URL+"/admin/posts", url.Values{
		"title":   {"Made In Admin"},
		"body_md": {"body text"},
		"status":  {"published"},
		"tags":    {"go, test"},
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("create post = %d, want 303", resp.StatusCode)
	}

	// It is now publicly visible at its derived slug.
	resp, _ = client.Get(srv.URL + "/posts/made-in-admin/")
	body := readBody(t, resp)
	if resp.StatusCode != http.StatusOK || !strings.Contains(body, "Made In Admin") {
		t.Fatalf("published post not visible: status=%d", resp.StatusCode)
	}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	var b bytes.Buffer
	if _, err := b.ReadFrom(resp.Body); err != nil {
		t.Fatal(err)
	}
	return b.String()
}
