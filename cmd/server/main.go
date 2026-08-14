// Command server runs the aniicrite.dev site: a Go SSR + HTMX monolith backed
// by SQLite. Run `server hash-password` to generate an admin password hash.
package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"aniicrite.dev/internal/auth"
	"aniicrite.dev/internal/config"
	"aniicrite.dev/internal/db"
	"aniicrite.dev/internal/handlers"
	"aniicrite.dev/internal/importer"
	"aniicrite.dev/internal/markdown"
	"aniicrite.dev/internal/models"
	"aniicrite.dev/web"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "hash-password":
			hashPassword(os.Args[2:])
			return
		case "import":
			if err := runImport(os.Args[2:]); err != nil {
				log.Fatal(err)
			}
			return
		case "import-projects":
			if err := runImportProjects(os.Args[2:]); err != nil {
				log.Fatal(err)
			}
			return
		}
	}
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.Load()

	uploadDir := filepath.Join(cfg.DataDir, "uploads")
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return fmt.Errorf("create upload dir: %w", err)
	}

	sqldb, err := db.Open(filepath.Join(cfg.DataDir, "site.db"))
	if err != nil {
		return err
	}
	defer sqldb.Close()
	store := models.New(sqldb)

	// Prefer an operator-supplied data/about.md; fall back to the embedded
	// default. Lets forks edit the About page without rebuilding.
	aboutMD, err := os.ReadFile(filepath.Join(cfg.DataDir, "about.md"))
	if err != nil {
		if aboutMD, err = web.ContentFS.ReadFile("content/about.md"); err != nil {
			return fmt.Errorf("read about.md: %w", err)
		}
	}
	aboutHTML, err := markdown.Render(string(aboutMD))
	if err != nil {
		return fmt.Errorf("render about.md: %w", err)
	}

	sess := auth.NewManager(cfg.SessionHashKey, cfg.SessionBlockKey, !cfg.Dev)

	var gh *auth.GitHub
	if cfg.CommentsEnabled() {
		gh = auth.NewGitHub(cfg.GitHubClientID, cfg.GitHubClientSecret, cfg.BaseURL+"/auth/github/callback")
		log.Println("comments: GitHub OAuth enabled")
	} else {
		log.Println("comments: disabled (set GITHUB_CLIENT_ID/SECRET to enable)")
	}

	staticFS, err := fs.Sub(web.StaticFS, "static")
	if err != nil {
		return err
	}

	h := handlers.New(cfg, store, sess, gh, aboutHTML, uploadDir)
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           h.Routes(staticFS),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("listening on %s (base %s)", cfg.Addr, cfg.BaseURL)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down…")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}

// runImport migrates Markdown files from a directory into the store.
func runImport(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: server import <dir>")
	}
	cfg := config.Load()
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return err
	}
	sqldb, err := db.Open(filepath.Join(cfg.DataDir, "site.db"))
	if err != nil {
		return err
	}
	defer sqldb.Close()
	n, err := importer.ImportDir(context.Background(), models.New(sqldb), args[0])
	if err != nil {
		return err
	}
	log.Printf("imported %d posts from %s", n, args[0])
	return nil
}

// runImportProjects imports projects from a JSON file into the store.
func runImportProjects(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: server import-projects <file.json>")
	}
	cfg := config.Load()
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return err
	}
	sqldb, err := db.Open(filepath.Join(cfg.DataDir, "site.db"))
	if err != nil {
		return err
	}
	defer sqldb.Close()
	n, err := importer.ImportProjectsFile(context.Background(), models.New(sqldb), args[0])
	if err != nil {
		return err
	}
	log.Printf("imported %d projects from %s", n, args[0])
	return nil
}

// hashPassword prints an argon2id hash for the given password (or stdin).
func hashPassword(args []string) {
	var pw string
	if len(args) > 0 {
		pw = args[0]
	} else {
		fmt.Fprint(os.Stderr, "password: ")
		fmt.Scanln(&pw)
	}
	if pw == "" {
		log.Fatal("empty password")
	}
	hash, err := auth.HashPassword(pw)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(hash)
}
