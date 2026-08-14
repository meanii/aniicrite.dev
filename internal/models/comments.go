package models

import "context"

// CommentInput carries a new comment's fields.
type CommentInput struct {
	PostID   int64
	GHUserID int64
	GHLogin  string
	GHAvatar string
	Body     string
}

// IsBlocked reports whether a GitHub user is banned from commenting.
func (s *Store) IsBlocked(ctx context.Context, ghUserID int64) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM blocked_users WHERE gh_user_id = ?`, ghUserID).Scan(&n)
	return n > 0, err
}

// CreateComment inserts a visible comment.
func (s *Store) CreateComment(ctx context.Context, in CommentInput) (int64, error) {
	res, err := s.db.ExecContext(ctx, `INSERT INTO comments
		(post_id, gh_user_id, gh_login, gh_avatar, body) VALUES (?,?,?,?,?)`,
		in.PostID, in.GHUserID, in.GHLogin, in.GHAvatar, in.Body)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// CommentsForPost returns visible comments oldest-first.
func (s *Store) CommentsForPost(ctx context.Context, postID int64) ([]Comment, error) {
	return s.comments(ctx, `WHERE post_id = ? AND hidden = 0 ORDER BY created_at ASC`, postID)
}

// AllComments returns every comment newest-first for the admin.
func (s *Store) AllComments(ctx context.Context) ([]Comment, error) {
	return s.comments(ctx, `ORDER BY created_at DESC`)
}

func (s *Store) comments(ctx context.Context, clause string, args ...any) ([]Comment, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, post_id, gh_user_id, gh_login, gh_avatar, body, hidden, created_at FROM comments `+clause, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.PostID, &c.GHUserID, &c.GHLogin, &c.GHAvatar, &c.Body, &c.Hidden, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// CommentCount returns the number of visible comments on a post.
func (s *Store) CommentCount(ctx context.Context, postID int64) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM comments WHERE post_id = ? AND hidden = 0`, postID).Scan(&n)
	return n, err
}

// DeleteComment removes a comment by id.
func (s *Store) DeleteComment(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM comments WHERE id = ?`, id)
	return err
}

// BlockUser bans a GitHub user and deletes their existing comments.
func (s *Store) BlockUser(ctx context.Context, id int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var login string
	var ghUserID int64
	if err := tx.QueryRowContext(ctx, `SELECT gh_user_id, gh_login FROM comments WHERE id = ?`, id).Scan(&ghUserID, &login); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO blocked_users(gh_user_id, gh_login) VALUES (?, ?)`, ghUserID, login); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM comments WHERE gh_user_id = ?`, ghUserID); err != nil {
		return err
	}
	return tx.Commit()
}
