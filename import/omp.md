---
title: How I run omp on my laptop
slug: omp
date: 2026-08-13T00:00:00Z
tags: omp, AI, Obsidian, tools, workflow
status: published
summary: omp is the terminal agent I do most of my work in. Obsidian is where it keeps its notes, and omp-lb is a small local panel for seeing what it is doing. This is how the three fit together.
---
Most AI coding tools are a chat window with a code block in the reply. You paste the code in, it does not quite work, you go back and forth. omp is not that. It runs in a terminal, inside the repo I am working on, and it does the work itself: reads the files, edits them, runs the build and the tests, reads the output and tries again. Most of my working day now happens inside it.

Three things make up the setup on my laptop. omp itself, Obsidian for its memory, and omp-lb for watching what it is doing.

## omp

omp (oh-my-pi) is a terminal-native agent harness. Harness is the right word. The model is one component. What makes it useful is the set of tools around it and the loop that runs them.

It can read and edit files anywhere in the repo, including structural edits across many files. It runs shell commands and reads the result, so builds, tests, git and migrations are things it does rather than things it tells me to do. It talks to language servers for go-to-definition, rename and find-references, so it does not have to guess at call sites. It can drive a browser when a task needs a rendered page or a login. It has fast grep and glob over the tree. It can fan work out to sub-agents and collect the results. And it can follow skills, which are reusable procedures, and connect to MCP servers for outside tools.

The thing that matters is that it closes the loop. It makes a change, runs the tests, watches them fail, fixes the change, and then tells me what it did and how it verified it. I use it for features, debugging, refactors, deploys and reading code I have never seen before. I also use it for ops work, research and writing.

## Obsidian

An agent that forgets everything between sessions is a lot less useful. I use [Obsidian](https://obsidian.md) as omp's memory. omp has a note-taking skill that reads and writes an Obsidian vault directly, so its notes and my notes are the same Markdown files in the same folder.

I like this for two reasons. The notes are plain files, so if I stop using omp tomorrow nothing is lost inside a model. And when it has remembered something wrong I open the note and fix it, which is the whole correction. No prompt wrangling.

The vault syncs to my other devices, so whatever the agent wrote is there on my phone, and whatever I jotted down on my phone is there for it next time.

## omp-lb

Once omp is doing real work across a few accounts and many sessions, it gets hard to keep track of. Which credential is it using? How much quota is left on it? What did it do in that session two hours ago? [omp-lb](https://github.com/inhumanxd/omp-lb) is a small local panel that answers those.

It is a single Node server with no dependencies and no build step. You run `node server.mjs`, point it at omp's data directory, and open `http://127.0.0.1:8787`. It reads omp's own databases, `agent.db` for accounts and read-only copies of `history.db` and `models.db`, and shows them in a browser. There is a page per thing: accounts, with enable, disable and use-only-this, and the real per-account quota from omp's cached usage reports; sessions, with full text search over the prompt history, which is the page I use most; the model catalogue per provider and which role each model is wired to; a usage chart; and a system page with theme, cache inspector and backups.

Some details I care about. It binds to 127.0.0.1 only, so nothing is on the network. It is not in the request path. It reads omp's data and manages accounts, and it never proxies a request. Every write first takes a `VACUUM INTO` snapshot next to `agent.db` and keeps the last 20, so a bad toggle is one file copy from undone. omp loads credentials at session start, so account changes apply to new sessions, not the one that is already running. And if your data is not in `~/.omp/agent`, set `OMP_DB` to the right `agent.db` and it finds the other files next to it.

## Together

I keep notes in Obsidian. omp reads and writes the same vault while it works in my repos and my shell. omp-lb sits off to the side showing me accounts, quota and every past session. It is not a complicated setup. It is just the one that made me trust an agent with real work.
