---
title: I rebuilt this site from scratch in Go
slug: building-this-site-in-go
date: 2026-08-14T00:00:00Z
tags: Go, HTMX, SQLite, self-hosting
status: published
summary: This site ran on Hugo for years. Now it is a small Go server with templ, HTMX and SQLite in one binary, with an admin page where I write posts in the browser.
---
This site ran on Hugo for a long time. Hugo was never the problem. The problem was that publishing a post meant opening the repo, writing Markdown, running the build and pushing the output, and I write less when writing has steps. I wanted to open a page in a browser, type, press save and be done. So I wrote a small blog engine for myself. This post is running on it.

## What it is made of

It is a Go server with no web framework. The pieces:

[templ](https://templ.guide) for templates. They compile to Go, so a typo in a template fails the build instead of showing up as a blank page in production.

[HTMX](https://htmx.org) for the two or three places that need interaction, like search as you type and the comment form. I did not want to ship a front end bundle for a blog.

SQLite for storage, using the pure Go driver so there is no cgo. Posts, comments and projects are all in one file, and FTS5 gives me full text search without running anything extra. Backup is copying one file.

goldmark renders the Markdown and chroma highlights code, both on the server at save time, so the rendered HTML is stored next to the source.

Templates, CSS and the HTMX script are embedded in the binary. The container image is the distroless base plus one file.

## The admin page

This is the part that justified writing a backend. There is a login, a Markdown editor with a live preview beside it, image upload and a publish button. Projects on the home page are edited the same way. Comments go through GitHub login, which has kept the spam at zero so far without me moderating anything.

## How it runs

Caddy sits in front and handles TLS. The binary runs as a container on the same VPS that hosts everything else I run. A GitHub Actions workflow runs the tests on every push to main and then SSHes into the server, pulls, rebuilds the image and restarts the container. That is the entire release process, and it has been reliable enough that I stopped thinking about it.

## The code

It is on GitHub at [meanii/aniicrite.dev](https://github.com/meanii/aniicrite.dev) under the MIT licence. Site name, author, social links and so on come from environment variables, so you should be able to run your own copy without editing any Go. If you try it and find something hardcoded that should not be, open an issue.
