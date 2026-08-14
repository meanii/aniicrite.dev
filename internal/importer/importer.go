// Package importer performs the one-time migration of Hugo-style Markdown
// files into the SQLite store. Files carry a minimal front matter block:
//
//	---
//	title: Goodbye kitty
//	slug: goodbye-kitty
//	date: 2024-07-24T00:00:00Z
//	tags: tmux, terminal, Linux
//	status: published
//	---
//	<markdown body>
package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aniicrite.dev/internal/markdown"
	"aniicrite.dev/internal/models"
)

// ImportDir imports every *.md file in dir, returning the number imported.
func ImportDir(ctx context.Context, store *models.Store, dir string) (int, error) {
	entries, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return 0, err
	}
	n := 0
	for _, path := range entries {
		if err := importFile(ctx, store, path); err != nil {
			return n, fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
		n++
	}
	return n, nil
}

func importFile(ctx context.Context, store *models.Store, path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	meta, body, err := split(string(raw))
	if err != nil {
		return err
	}

	slug := meta["slug"]
	if slug == "" {
		slug = models.Slugify(meta["title"])
	}
	date := time.Now().UTC()
	if d := meta["date"]; d != "" {
		if parsed, err := time.Parse(time.RFC3339, d); err == nil {
			date = parsed
		} else {
			return fmt.Errorf("bad date %q: %w", d, err)
		}
	}
	status := models.StatusPublished
	if meta["status"] == models.StatusDraft {
		status = models.StatusDraft
	}
	bodyHTML, err := markdown.Render(body)
	if err != nil {
		return err
	}
	summary := meta["summary"]
	if summary == "" {
		summary = markdown.Summary(bodyHTML, 200)
	}

	in := models.PostInput{
		Slug:     slug,
		Title:    meta["title"],
		Summary:  summary,
		BodyMD:   body,
		BodyHTML: bodyHTML,
		Status:   status,
		Reading:  markdown.ReadingMinutes(body),
		Tags:     models.ParseTags(meta["tags"]),
	}
	return store.ImportPost(ctx, in, date)
}

// split separates the front matter map from the Markdown body.
func split(content string) (map[string]string, string, error) {
	content = strings.TrimLeft(content, "\ufeff \t\r\n")
	if !strings.HasPrefix(content, "---") {
		return nil, "", fmt.Errorf("missing front matter")
	}
	rest := content[3:]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return nil, "", fmt.Errorf("unterminated front matter")
	}
	head := rest[:end]
	body := rest[end+4:]
	if i := strings.IndexByte(body, '\n'); i >= 0 {
		body = body[i+1:]
	}

	meta := map[string]string{}
	for _, line := range strings.Split(head, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		meta[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return meta, strings.TrimSpace(body), nil
}

// projectFile is one entry in the projects JSON import file.
type projectFile struct {
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	RepoURL     string   `json:"repo_url"`
	Tags        []string `json:"tags"`
	SortOrder   int      `json:"sort_order"`
	Status      string   `json:"status"`
}

// ImportProjectsFile imports projects from a JSON array file, returning the count.
func ImportProjectsFile(ctx context.Context, store *models.Store, path string) (int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var entries []projectFile
	if err := json.Unmarshal(raw, &entries); err != nil {
		return 0, fmt.Errorf("parse %s: %w", filepath.Base(path), err)
	}
	for i, e := range entries {
		slug := e.Slug
		if slug == "" {
			slug = models.Slugify(e.Title)
		}
		descHTML, err := markdown.Render(e.Description)
		if err != nil {
			return i, err
		}
		status := models.StatusPublished
		if e.Status == models.StatusDraft {
			status = models.StatusDraft
		}
		in := models.ProjectInput{
			Slug: slug, Title: e.Title, DescMD: e.Description, DescHTML: descHTML,
			URL: e.URL, RepoURL: e.RepoURL, SortOrder: e.SortOrder, Status: status, Tags: e.Tags,
		}
		if err := store.ImportProject(ctx, in); err != nil {
			return i, fmt.Errorf("import %q: %w", e.Title, err)
		}
	}
	return len(entries), nil
}
