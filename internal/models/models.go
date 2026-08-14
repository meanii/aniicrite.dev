// Package models defines the domain types and their SQLite-backed data access.
package models

import (
	"context"
	"database/sql"
	"time"
)

// Status values for posts and projects.
const (
	StatusDraft     = "draft"
	StatusPublished = "published"
)

// Post is a blog entry.
type Post struct {
	ID             int64
	Slug           string
	Title          string
	Summary        string
	BodyMD         string
	BodyHTML       string
	CoverImage     string
	Status         string
	ReadingMinutes int
	ViewCount      int
	PublishedAt    sql.NullTime
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Tags           []Tag
}

// Published reports whether the post is publicly visible.
func (p Post) Published() bool { return p.Status == StatusPublished }

// Date is the display date: publish time when set, else creation time.
func (p Post) Date() time.Time {
	if p.PublishedAt.Valid {
		return p.PublishedAt.Time
	}
	return p.CreatedAt
}

// Project is a portfolio entry.
type Project struct {
	ID        int64
	Slug      string
	Title     string
	DescMD    string
	DescHTML  string
	URL       string
	RepoURL   string
	SortOrder int
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
	Tags      []Tag
}

// Comment is a reader comment tied to a GitHub identity.
type Comment struct {
	ID        int64
	PostID    int64
	GHUserID  int64
	GHLogin   string
	GHAvatar  string
	Body      string
	Hidden    bool
	CreatedAt time.Time
}

// Tag labels posts and projects.
type Tag struct {
	ID   int64
	Name string
	Slug string
}

// Store is the data access layer over a single SQLite database.
type Store struct {
	db *sql.DB
}

// New wraps an open database.
func New(db *sql.DB) *Store { return &Store{db: db} }

// Ping verifies database connectivity.
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
