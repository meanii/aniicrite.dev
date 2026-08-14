---
title: I rebuilt this site from scratch in Go
slug: building-this-site-in-go
date: 2026-08-14T00:00:00Z
tags: Go, HTMX, SQLite, self-hosting
status: published
summary: Why I dropped Hugo and wrote my own small blog engine — Go, templ, HTMX, and SQLite in a single binary.
---
This site used to run on Hugo. It was fine, but publishing meant editing Markdown, rebuilding, and pushing files. I wanted to write a post from a browser, hit save, and be done. So I wrote my own engine.

### The stack

It's a plain Go server, no framework doing the heavy lifting:

- **[templ](https://templ.guide)** for HTML templates, compiled to Go so mistakes are caught at build time.
- **[HTMX](https://htmx.org)** for the few interactive bits — live search, the comment form — instead of shipping a SPA.
- **SQLite** for storage, with FTS5 for full-text search over posts. One file on disk, trivial to back up.
- **goldmark + chroma** to render Markdown and highlight code on the server.

Everything — templates, CSS, the htmx script — is embedded into one static binary. No cgo, so it cross-compiles cleanly and drops into a `FROM scratch`-style image.

### The admin side

The reason I wrote a backend at all: there's an admin panel. I log in, write a post in Markdown with a live preview next to it, upload images, and publish. Projects and comments are managed the same way. Comments use GitHub OAuth so I'm not moderating spam by hand.

### Running it

The binary sits behind Caddy, which handles TLS. It runs as a container on my VPS, and a GitHub Actions workflow redeploys it on every push to `main` after the tests pass — `git pull`, rebuild, restart. Pushing a commit is the whole deploy.

### It's open source

The code is on GitHub: [meanii/aniicrite.dev](https://github.com/meanii/aniicrite.dev). It's MIT-licensed, and I tried to keep the personal bits (identity, socials, content) in config so anyone can run their own copy without editing Go.
