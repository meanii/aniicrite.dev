package models_test

import (
	"context"
	"path/filepath"
	"testing"

	"aniicrite.dev/internal/db"
	"aniicrite.dev/internal/models"
)

func newStore(t *testing.T) *models.Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	sqldb, err := db.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { sqldb.Close() })
	return models.New(sqldb)
}

func TestPostLifecycleAndSearch(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)

	id, err := s.CreatePost(ctx, models.PostInput{
		Slug:    "hello-golang",
		Title:   "Hello Golang",
		BodyMD:  "Configuration management with Viper in Go.",
		Status:  models.StatusPublished,
		Reading: 3,
		Tags:    []string{"Go", "Config"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// A draft must not appear in public listings or search.
	if _, err := s.CreatePost(ctx, models.PostInput{
		Slug: "secret", Title: "Secret Draft", BodyMD: "golang draft", Status: models.StatusDraft,
	}); err != nil {
		t.Fatalf("create draft: %v", err)
	}

	posts, total, err := s.ListPosts(ctx, 10, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(posts) != 1 {
		t.Fatalf("want 1 published post, got total=%d len=%d", total, len(posts))
	}
	if got := posts[0]; !got.PublishedAt.Valid || got.Date().IsZero() {
		t.Fatalf("published_at not scanned as time: %+v", got.PublishedAt)
	}
	if len(posts[0].Tags) != 2 {
		t.Fatalf("want 2 tags, got %d", len(posts[0].Tags))
	}

	// FTS: prefix match on "gol" should hit the published post only.
	hits, err := s.Search(ctx, "gol", 10)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(hits) != 1 || hits[0].Slug != "hello-golang" {
		t.Fatalf("search want [hello-golang], got %+v", hits)
	}

	// Update body so FTS index must re-sync via trigger.
	if err := s.UpdatePost(ctx, id, models.PostInput{
		Slug: "hello-golang", Title: "Hello Golang", BodyMD: "now about kubernetes", Status: models.StatusPublished,
	}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if hits, _ := s.Search(ctx, "kubernetes", 10); len(hits) != 1 {
		t.Fatalf("search after update: want 1 kubernetes hit, got %d", len(hits))
	}
	if hits, _ := s.Search(ctx, "viper", 10); len(hits) != 0 {
		t.Fatalf("stale FTS: 'viper' should no longer match, got %d", len(hits))
	}

	// Publish→republish keeps the original publish time.
	p0, _ := s.PostBySlug(ctx, "hello-golang", false)
	if err := s.UpdatePost(ctx, id, models.PostInput{Slug: "hello-golang", Title: "x", BodyMD: "y", Status: models.StatusPublished}); err != nil {
		t.Fatalf("republish: %v", err)
	}
	p1, _ := s.PostBySlug(ctx, "hello-golang", false)
	if !p0.PublishedAt.Time.Equal(p1.PublishedAt.Time) {
		t.Fatalf("publish time changed on re-save: %v != %v", p0.PublishedAt.Time, p1.PublishedAt.Time)
	}
}

func TestCommentsAndBlocking(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)
	pid, err := s.CreatePost(ctx, models.PostInput{Slug: "p", Title: "P", BodyMD: "b", Status: models.StatusPublished})
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	cid, err := s.CreateComment(ctx, models.CommentInput{PostID: pid, GHUserID: 42, GHLogin: "spammer", Body: "hi"})
	if err != nil {
		t.Fatalf("comment: %v", err)
	}
	if n, _ := s.CommentCount(ctx, pid); n != 1 {
		t.Fatalf("want 1 comment, got %d", n)
	}
	if err := s.BlockUser(ctx, cid); err != nil {
		t.Fatalf("block: %v", err)
	}
	if n, _ := s.CommentCount(ctx, pid); n != 0 {
		t.Fatalf("block should delete comments, got %d", n)
	}
	if blocked, _ := s.IsBlocked(ctx, 42); !blocked {
		t.Fatalf("user 42 should be blocked")
	}
}
