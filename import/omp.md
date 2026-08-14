---
title: omp — the AI agent I actually work in
slug: omp
date: 2026-08-13T00:00:00Z
tags: omp, AI, tools, workflow
status: published
summary: omp runs on my laptop all day. It edits real code and runs real commands instead of handing me snippets to paste.
---
Most "AI coding" tools are a chat window. You describe a problem, you get a block of code back, you paste it in, and you fix whatever's wrong. omp isn't that. It runs on my laptop, in my actual repos, and does the work itself — reads the files, makes the edits, runs the commands, checks the result. I spend most of my working day in it.

omp (oh-my-pi) is a terminal-native agent harness. I call it a harness because the model is only one part; the useful bit is everything wrapped around it.

## What it can actually touch

The difference from a chat box is that omp has real tools and uses them:

- **Files** — it reads and edits across the whole repo, structural edits included, not just suggestions in a reply.
- **Shell** — it runs commands and reads the output: build, test, git, migrations, whatever the task needs.
- **Code intelligence** — language servers for real go-to-definition, rename, and find-references; a debugger; and it runs the test suite itself.
- **Browser** — it drives a real browser for anything that needs a rendered page or a login.
- **Search** — fast grep and glob over the entire tree instead of guessing where things are.
- **Sub-agents** — it fans work out in parallel and pulls the results back together.
- **Skills and MCP** — reusable procedures it follows, plus connections to outside tools and servers.

The point of all that: it closes the loop on its own. It doesn't just propose a change, it makes the change, runs the tests, sees them fail, fixes them, and then tells me what it did and how it checked.

## How I use it

It's my daily driver for work — building features, debugging, refactoring, wiring up deploys, and reading through code I've never seen before. Because it works in the real repo, it has context a chat tab never gets: the actual files, the actual errors, the actual command output.

And it's not just code. I lean on it for ops and one-off automation, for research and reading, and for writing. The pattern is the same every time: I give it the goal, it does the legwork, and it shows me the evidence instead of asking me to trust it.

## Why the terminal, not a chat tab

Two reasons. First, the work already lives there — the repo, the shell, the tools are all one command away. Second, it does the boring part. A chat tab makes me the integration layer, shuttling code back and forth by hand. omp runs the commands and iterates, so I review results instead of assembling them.

It isn't magic and it isn't fully autonomous — I'm still steering, and I still read what it does. But it moved my job from "write this code" to "here's the goal, go," and most days that's exactly the trade I want.
