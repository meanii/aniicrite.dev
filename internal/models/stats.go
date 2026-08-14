package models

import "context"

// Counts holds dashboard tallies.
type Counts struct {
	Posts    int
	Drafts   int
	Projects int
	Comments int
}

// DashboardCounts returns published/draft post counts plus project and comment totals.
func (s *Store) DashboardCounts(ctx context.Context) (Counts, error) {
	var c Counts
	row := s.db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM posts WHERE status='published'),
			(SELECT COUNT(*) FROM posts WHERE status='draft'),
			(SELECT COUNT(*) FROM projects),
			(SELECT COUNT(*) FROM comments)`)
	err := row.Scan(&c.Posts, &c.Drafts, &c.Projects, &c.Comments)
	return c, err
}

// CommentWithPost is a comment plus the post it belongs to.
type CommentWithPost struct {
	Comment
	PostSlug  string
	PostTitle string
}

// AllCommentsWithPost returns every comment newest-first, joined to its post.
func (s *Store) AllCommentsWithPost(ctx context.Context) ([]CommentWithPost, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.id, c.post_id, c.gh_user_id, c.gh_login, c.gh_avatar, c.body, c.hidden, c.created_at,
		       p.slug, p.title
		FROM comments c JOIN posts p ON p.id = c.post_id
		ORDER BY c.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CommentWithPost
	for rows.Next() {
		var c CommentWithPost
		if err := rows.Scan(&c.ID, &c.PostID, &c.GHUserID, &c.GHLogin, &c.GHAvatar,
			&c.Body, &c.Hidden, &c.CreatedAt, &c.PostSlug, &c.PostTitle); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
