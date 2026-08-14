package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrNotFound is returned when a lookup matches no row.
var ErrNotFound = errors.New("not found")

const postCols = `id, slug, title, summary, body_md, body_html, cover_image,
	status, reading_minutes, view_count, published_at, created_at, updated_at`

func scanPost(sc interface{ Scan(...any) error }) (Post, error) {
	var p Post
	err := sc.Scan(&p.ID, &p.Slug, &p.Title, &p.Summary, &p.BodyMD, &p.BodyHTML,
		&p.CoverImage, &p.Status, &p.ReadingMinutes, &p.ViewCount,
		&p.PublishedAt, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

// PostInput carries the editable fields of a post.
type PostInput struct {
	Slug       string
	Title      string
	Summary    string
	BodyMD     string
	BodyHTML   string
	CoverImage string
	Status     string
	Reading    int
	Tags       []string
}

// ImportPost upserts a post by slug, preserving an explicit publish date.
// Used by the one-time migration; idempotent so re-running is safe.
func (s *Store) ImportPost(ctx context.Context, in PostInput, publishedAt time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var id int64
	err = tx.QueryRowContext(ctx, `SELECT id FROM posts WHERE slug = ?`, in.Slug).Scan(&id)
	switch err {
	case nil:
		if _, err = tx.ExecContext(ctx, `UPDATE posts SET
			title=?, summary=?, body_md=?, body_html=?, cover_image=?,
			status=?, reading_minutes=?, published_at=?,
			updated_at=strftime('%Y-%m-%dT%H:%M:%SZ','now')
			WHERE id=?`,
			in.Title, in.Summary, in.BodyMD, in.BodyHTML, in.CoverImage,
			in.Status, in.Reading, isoTime(publishedAt), id); err != nil {
			return err
		}
	case sql.ErrNoRows:
		res, err := tx.ExecContext(ctx, `INSERT INTO posts
			(slug, title, summary, body_md, body_html, cover_image, status, reading_minutes, published_at, created_at)
			VALUES (?,?,?,?,?,?,?,?,?,?)`,
			in.Slug, in.Title, in.Summary, in.BodyMD, in.BodyHTML, in.CoverImage,
			in.Status, in.Reading, isoTime(publishedAt), isoTime(publishedAt))
		if err != nil {
			return err
		}
		if id, err = res.LastInsertId(); err != nil {
			return err
		}
	default:
		return err
	}
	if err := setPostTags(ctx, tx, id, in.Tags); err != nil {
		return err
	}
	return tx.Commit()
}

// CreatePost inserts a post and its tags, returning the new id.
func (s *Store) CreatePost(ctx context.Context, in PostInput) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var published any
	if in.Status == StatusPublished {
		published = isoTime(time.Now())
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO posts
		(slug, title, summary, body_md, body_html, cover_image, status, reading_minutes, published_at)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		in.Slug, in.Title, in.Summary, in.BodyMD, in.BodyHTML, in.CoverImage, in.Status, in.Reading, published)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := setPostTags(ctx, tx, id, in.Tags); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

// UpdatePost updates an existing post. Transitioning draft→published stamps
// published_at once; it is preserved on subsequent edits.
func (s *Store) UpdatePost(ctx context.Context, id int64, in PostInput) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var curStatus string
	var curPublished sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT status, published_at FROM posts WHERE id = ?`, id).
		Scan(&curStatus, &curPublished); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	var published any
	switch {
	case in.Status == StatusPublished && curPublished.Valid:
		published = isoTime(curPublished.Time) // keep original publish time
	case in.Status == StatusPublished:
		published = isoTime(time.Now()) // first publish
	default:
		published = nil // back to draft
	}

	_, err = tx.ExecContext(ctx, `UPDATE posts SET
		slug=?, title=?, summary=?, body_md=?, body_html=?, cover_image=?,
		status=?, reading_minutes=?, published_at=?,
		updated_at=strftime('%Y-%m-%dT%H:%M:%SZ','now')
		WHERE id=?`,
		in.Slug, in.Title, in.Summary, in.BodyMD, in.BodyHTML, in.CoverImage,
		in.Status, in.Reading, published, id)
	if err != nil {
		return err
	}
	if err := setPostTags(ctx, tx, id, in.Tags); err != nil {
		return err
	}
	return tx.Commit()
}

// DeletePost removes a post (cascading to tags/comments).
func (s *Store) DeletePost(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM posts WHERE id = ?`, id)
	return err
}

// PostByID loads a single post by id, including its tags.
func (s *Store) PostByID(ctx context.Context, id int64) (Post, error) {
	p, err := scanPost(s.db.QueryRowContext(ctx, `SELECT `+postCols+` FROM posts WHERE id = ?`, id))
	if err == sql.ErrNoRows {
		return Post{}, ErrNotFound
	}
	if err != nil {
		return Post{}, err
	}
	tags, err := s.tagsForPosts(ctx, []int64{p.ID})
	if err != nil {
		return Post{}, err
	}
	p.Tags = tags[p.ID]
	return p, nil
}

// PostBySlug loads a published post plus its tags. When includeDrafts is true,
// drafts are also returned (for admin preview).
func (s *Store) PostBySlug(ctx context.Context, slug string, includeDrafts bool) (Post, error) {
	q := `SELECT ` + postCols + ` FROM posts WHERE slug = ?`
	if !includeDrafts {
		q += ` AND status = 'published'`
	}
	p, err := scanPost(s.db.QueryRowContext(ctx, q, slug))
	if err == sql.ErrNoRows {
		return Post{}, ErrNotFound
	}
	if err != nil {
		return Post{}, err
	}
	tags, err := s.tagsForPosts(ctx, []int64{p.ID})
	if err != nil {
		return Post{}, err
	}
	p.Tags = tags[p.ID]
	return p, nil
}

// ListPosts returns published posts newest-first, paginated, with tags.
// It also reports the total published count for pagination.
func (s *Store) ListPosts(ctx context.Context, limit, offset int) ([]Post, int, error) {
	return s.listPosts(ctx, true, limit, offset)
}

// ListAllPosts returns every post (drafts included) for the admin, newest-first.
func (s *Store) ListAllPosts(ctx context.Context) ([]Post, error) {
	posts, _, err := s.listPosts(ctx, false, 1000, 0)
	return posts, err
}

func (s *Store) listPosts(ctx context.Context, publishedOnly bool, limit, offset int) ([]Post, int, error) {
	where := ""
	order := ` ORDER BY COALESCE(published_at, created_at) DESC`
	if publishedOnly {
		where = ` WHERE status = 'published'`
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+postCols+` FROM posts`+where+order+` LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var posts []Post
	var ids []int64
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, 0, err
		}
		posts = append(posts, p)
		ids = append(ids, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	tags, err := s.tagsForPosts(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	for i := range posts {
		posts[i].Tags = tags[posts[i].ID]
	}

	var total int
	countQ := `SELECT COUNT(*) FROM posts` + where
	if err := s.db.QueryRowContext(ctx, countQ).Scan(&total); err != nil {
		return nil, 0, err
	}
	return posts, total, nil
}

// IncrementViews bumps a post's view counter.
func (s *Store) IncrementViews(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE posts SET view_count = view_count + 1 WHERE id = ?`, id)
	return err
}

// SlugExists reports whether a slug is taken by a different post.
func (s *Store) SlugExists(ctx context.Context, slug string, excludeID int64) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts WHERE slug = ? AND id <> ?`, slug, excludeID).Scan(&n)
	return n > 0, err
}

// UniqueSlug returns slug or slug-2, slug-3… so it does not collide.
func (s *Store) UniqueSlug(ctx context.Context, base string, excludeID int64) (string, error) {
	slug := base
	for i := 2; ; i++ {
		exists, err := s.SlugExists(ctx, slug, excludeID)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
}

// isoTime formats a time as the same ISO8601-UTC layout used by the schema's
// strftime defaults, keeping all stored timestamps in one comparable format.
func isoTime(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05Z") }
