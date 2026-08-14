---
title: omp — the AI agent I actually work in
slug: omp
date: 2026-08-13T00:00:00Z
tags: omp, AI, Obsidian, tools, workflow
status: published
summary: My laptop setup around omp — the agent runtime itself, Obsidian as its memory, and omp-lb to see what it's doing.
---
Most "AI coding" tools are a chat window. You describe a problem, you get a block of code back, you paste it in, and you fix whatever's wrong. omp isn't that. It runs on my laptop, in my actual repos, and does the work itself — reads the files, makes the edits, runs the commands, checks the result. I spend most of my working day in it.

There are three pieces to how I run it on my laptop: **omp** itself, **Obsidian** as its memory, and **omp-lb** to keep an eye on it. Here's each one.

## omp: the runtime

omp (oh-my-pi) is a terminal-native agent harness. I call it a harness because the model is only one part; the useful bit is everything wrapped around it. The difference from a chat box is that omp has real tools and uses them:

- **Files** — reads and edits across the whole repo, structural edits included, not just suggestions in a reply.
- **Shell** — runs commands and reads the output: build, test, git, migrations, whatever the task needs.
- **Code intelligence** — language servers for real go-to-definition, rename, and find-references; a debugger; and it runs the test suite itself.
- **Browser** — drives a real browser for anything that needs a rendered page or a login.
- **Search** — fast grep and glob over the entire tree instead of guessing where things are.
- **Sub-agents** — fans work out in parallel and pulls the results back together.
- **Skills and MCP** — reusable procedures it follows, plus connections to outside tools and servers.

The point of all that: it closes the loop on its own. It doesn't just propose a change, it makes the change, runs the tests, sees them fail, fixes them, and then tells me what it did and how it checked. I use it all day — features, debugging, refactoring, deploys, reading unfamiliar code — and for ops, research, and writing too.

## Obsidian: its memory

An agent is only as good as what it remembers between sessions. I use **[Obsidian](https://obsidian.md)** for that. omp ships a note-taking skill that reads and writes an Obsidian vault directly, so the agent's notes and my notes are the *same* notes — plain Markdown files in one vault on my laptop, not some opaque memory blob I can't see.

Two things I like about this:

- **It's plain Markdown.** If I turn omp off tomorrow, the notes are still just files. Nothing is locked inside a model.
- **I can read and edit its memory by hand.** When it remembers something wrong, I open the note and fix it, and that's the correction — no prompt-wrangling.

The vault lives on my laptop and syncs to my other devices, so what the agent writes shows up wherever I am, and anything I jot down is there for it next time.

## omp-lb: seeing what it's doing

Once omp is running real work across multiple accounts and sessions, you lose track of it. Which credential is it using? How much quota is left? What did it actually do in that session an hour ago? **[omp-lb](https://github.com/inhumanxd/omp-lb)** answers that.

It's a small local control panel for omp — zero dependencies, no build step, a single Node server (`node server.mjs`) that you point at omp's data directory and open at `http://127.0.0.1:8787`. It reads omp's own databases (`agent.db`, plus read-only `history.db` and `models.db`) and gives you a browser view over them:

- **Accounts** — every credential, with enable / disable / "use only this one", and the real per-account quota pulled from omp's own cached usage reports.
- **Sessions** — full prompt history by session, with full-text search. This is the tab I use most: I can go back and read exactly what I asked and how the agent answered.
- **Models** — the model catalogue per provider and which role each model is wired to.
- **Usage** — a prompt-activity chart and per-account quota windows, so I can see what's burning through what.
- **System** — theme, cache inspector, and backups.

A few details that matter:

- It **binds to `127.0.0.1` only** — nothing is exposed to the network, everyone runs their own local instance.
- It **does not proxy your requests.** It's a read-and-manage layer over omp's local data, not something in the request path; quota comes from omp's cached reports.
- Every write takes a `VACUUM INTO` snapshot next to `agent.db` and keeps the last 20, so a bad toggle is one file-copy away from undone.
- omp loads credentials at session start, so account changes apply to **new** sessions — start a fresh session to pick them up.
- If your data isn't in the default `~/.omp/agent`, point `OMP_DB` at the right `agent.db` and it finds the siblings alongside it.

## How they fit

On the laptop it's a tidy loop. I keep notes in **Obsidian**; **omp** reads and writes that same vault while it works in my repos and shell; and **omp-lb** sits off to the side — read-only on the request path — showing me accounts, quota, and every past session. omp does the work, Obsidian is the memory, and omp-lb is the window into both. None of the three is flashy alone; together they turn "an LLM plus some API keys" into an agent I actually trust to run real work.
