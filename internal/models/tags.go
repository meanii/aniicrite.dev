package models

import (
	"context"
	"database/sql"
	"strings"
)

// ParseTags splits a comma-separated tag input into unique display names.
func ParseTags(input string) []string {
	seen := map[string]bool{}
	var out []string
	for _, raw := range strings.Split(input, ",") {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, name)
	}
	return out
}

// upsertTag returns the id of a tag, creating it if absent.
func upsertTag(ctx context.Context, tx *sql.Tx, name string) (int64, error) {
	slug := Slugify(name)
	var id int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM tags WHERE slug = ?`, slug).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO tags(name, slug) VALUES (?, ?)`, name, slug)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// setPostTags replaces the tag set of a post within a transaction.
func setPostTags(ctx context.Context, tx *sql.Tx, postID int64, names []string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM post_tags WHERE post_id = ?`, postID); err != nil {
		return err
	}
	for _, name := range names {
		tagID, err := upsertTag(ctx, tx, name)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO post_tags(post_id, tag_id) VALUES (?, ?)`, postID, tagID); err != nil {
			return err
		}
	}
	return nil
}

// setProjectTags replaces the tag set of a project within a transaction.
func setProjectTags(ctx context.Context, tx *sql.Tx, projectID int64, names []string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM project_tags WHERE project_id = ?`, projectID); err != nil {
		return err
	}
	for _, name := range names {
		tagID, err := upsertTag(ctx, tx, name)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO project_tags(project_id, tag_id) VALUES (?, ?)`, projectID, tagID); err != nil {
			return err
		}
	}
	return nil
}

// tagsFor loads the tags attached to a set of post ids, keyed by post id.
func (s *Store) tagsForPosts(ctx context.Context, ids []int64) (map[int64][]Tag, error) {
	out := map[int64][]Tag{}
	if len(ids) == 0 {
		return out, nil
	}
	q := `SELECT pt.post_id, t.id, t.name, t.slug
	      FROM post_tags pt JOIN tags t ON t.id = pt.tag_id
	      WHERE pt.post_id IN (` + placeholders(len(ids)) + `)
	      ORDER BY t.name`
	rows, err := s.db.QueryContext(ctx, q, toAny(ids)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var pid int64
		var t Tag
		if err := rows.Scan(&pid, &t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		out[pid] = append(out[pid], t)
	}
	return out, rows.Err()
}

func (s *Store) tagsForProject(ctx context.Context, projectID int64) ([]Tag, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT t.id, t.name, t.slug
		FROM project_tags pt JOIN tags t ON t.id = pt.tag_id
		WHERE pt.project_id = ? ORDER BY t.name`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

func toAny[T any](in []T) []any {
	out := make([]any, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}

// TagBySlug loads a tag by its slug.
func (s *Store) TagBySlug(ctx context.Context, slug string) (Tag, error) {
	var t Tag
	err := s.db.QueryRowContext(ctx, `SELECT id, name, slug FROM tags WHERE slug = ?`, slug).
		Scan(&t.ID, &t.Name, &t.Slug)
	if err == sql.ErrNoRows {
		return Tag{}, ErrNotFound
	}
	return t, err
}

// PostsByTag returns published posts carrying the given tag, newest-first,
// each with its full tag set.
func (s *Store) PostsByTag(ctx context.Context, tagSlug string) ([]Post, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+qualifyCols(postCols, "p")+`
		FROM posts p
		JOIN post_tags pt ON pt.post_id = p.id
		JOIN tags t ON t.id = pt.tag_id
		WHERE t.slug = ? AND p.status = 'published'
		ORDER BY COALESCE(p.published_at, p.created_at) DESC`, tagSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []Post
	var ids []int64
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
		ids = append(ids, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	tags, err := s.tagsForPosts(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range posts {
		posts[i].Tags = tags[posts[i].ID]
	}
	return posts, nil
}

// AllTags returns tags that label at least one published post, with counts,
// alphabetically. The count rides along in Tag via a parallel slice is avoided
// by returning name/slug only; callers needing counts can extend later.
func (s *Store) AllTags(ctx context.Context) ([]Tag, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT t.id, t.name, t.slug
		FROM tags t
		JOIN post_tags pt ON pt.tag_id = t.id
		JOIN posts p ON p.id = pt.post_id AND p.status = 'published'
		ORDER BY t.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// qualifyCols prefixes each comma-separated column name with alias+".".
func qualifyCols(cols, alias string) string {
	parts := strings.Split(cols, ",")
	for i, p := range parts {
		parts[i] = alias + "." + strings.TrimSpace(p)
	}
	return strings.Join(parts, ", ")
}
