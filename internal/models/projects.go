package models

import (
	"context"
	"database/sql"
)

const projectCols = `id, slug, title, desc_md, desc_html, url, repo_url,
	sort_order, status, created_at, updated_at`

func scanProject(sc interface{ Scan(...any) error }) (Project, error) {
	var p Project
	err := sc.Scan(&p.ID, &p.Slug, &p.Title, &p.DescMD, &p.DescHTML, &p.URL,
		&p.RepoURL, &p.SortOrder, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

// ProjectInput carries the editable fields of a project.
type ProjectInput struct {
	Slug      string
	Title     string
	DescMD    string
	DescHTML  string
	URL       string
	RepoURL   string
	SortOrder int
	Status    string
	Tags      []string
}

// CreateProject inserts a project and its tags.
func (s *Store) CreateProject(ctx context.Context, in ProjectInput) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `INSERT INTO projects
		(slug, title, desc_md, desc_html, url, repo_url, sort_order, status)
		VALUES (?,?,?,?,?,?,?,?)`,
		in.Slug, in.Title, in.DescMD, in.DescHTML, in.URL, in.RepoURL, in.SortOrder, in.Status)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := setProjectTags(ctx, tx, id, in.Tags); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

// UpdateProject updates a project and its tags.
func (s *Store) UpdateProject(ctx context.Context, id int64, in ProjectInput) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `UPDATE projects SET
		slug=?, title=?, desc_md=?, desc_html=?, url=?, repo_url=?, sort_order=?, status=?,
		updated_at=strftime('%Y-%m-%dT%H:%M:%SZ','now')
		WHERE id=?`,
		in.Slug, in.Title, in.DescMD, in.DescHTML, in.URL, in.RepoURL, in.SortOrder, in.Status, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	if err := setProjectTags(ctx, tx, id, in.Tags); err != nil {
		return err
	}
	return tx.Commit()
}

// DeleteProject removes a project.
func (s *Store) DeleteProject(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	return err
}

// ProjectByID loads a project by id, with tags.
func (s *Store) ProjectByID(ctx context.Context, id int64) (Project, error) {
	p, err := scanProject(s.db.QueryRowContext(ctx, `SELECT `+projectCols+` FROM projects WHERE id = ?`, id))
	if err == sql.ErrNoRows {
		return Project{}, ErrNotFound
	}
	if err != nil {
		return Project{}, err
	}
	p.Tags, err = s.tagsForProject(ctx, p.ID)
	return p, err
}

// ListProjects returns projects ordered by sort_order then title. When
// publishedOnly is true, drafts are excluded (public view).
func (s *Store) ListProjects(ctx context.Context, publishedOnly bool) ([]Project, error) {
	where := ""
	if publishedOnly {
		where = ` WHERE status = 'published'`
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+projectCols+` FROM projects`+where+` ORDER BY sort_order ASC, title ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		tags, err := s.tagsForProject(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Tags = tags
	}
	return out, nil
}

// ProjectSlugExists reports whether a slug is taken by a different project.
func (s *Store) ProjectSlugExists(ctx context.Context, slug string, excludeID int64) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM projects WHERE slug = ? AND id <> ?`, slug, excludeID).Scan(&n)
	return n > 0, err
}
