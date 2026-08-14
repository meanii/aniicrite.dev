package models

import (
	"context"
	"strings"
)

// SearchResult is a lightweight post hit with a highlighted snippet.
type SearchResult struct {
	Slug    string
	Title   string
	Snippet string // safe HTML: FTS5 snippet with <mark> around matches
}

// Search runs a full-text query over published posts, returning ranked hits.
// The user query is tokenized and each term suffixed with '*' for prefix
// matching, so "gol" matches "golang".
func (s *Store) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	match := buildMatch(query)
	if match == "" {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.slug, p.title,
		       snippet(posts_fts, 2, '<mark>', '</mark>', '…', 12) AS snip
		FROM posts_fts
		JOIN posts p ON p.id = posts_fts.rowid
		WHERE posts_fts MATCH ? AND p.status = 'published'
		ORDER BY rank
		LIMIT ?`, match, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.Slug, &r.Title, &r.Snippet); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// buildMatch turns free text into a safe FTS5 MATCH expression: each
// alphanumeric token is quoted and given a prefix wildcard, joined by AND.
func buildMatch(query string) string {
	fields := strings.FieldsFunc(query, func(r rune) bool {
		return !(r == '_' ||
			('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') ||
			('0' <= r && r <= '9'))
	})
	var terms []string
	for _, f := range fields {
		if f == "" {
			continue
		}
		terms = append(terms, `"`+f+`"*`)
	}
	return strings.Join(terms, " AND ")
}
