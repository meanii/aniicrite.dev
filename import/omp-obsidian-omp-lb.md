---
title: How I run AI agents at home — omp, Obsidian, and omp-lb
slug: omp-obsidian-omp-lb
date: 2026-08-13T00:00:00Z
tags: omp, agents, Obsidian, self-hosting, AI
status: published
summary: The three pieces that make my home agent setup work — omp as the runtime, Obsidian as shared memory, and omp-lb as the dashboard.
---
I run a few AI agents on my homelab. Not a chatbot in a browser tab — actual agents that live on a box, hold their own memory, and answer over messaging. Three tools do the heavy lifting: **omp** runs the agents, **Obsidian** is their memory, and **omp-lb** is how I keep an eye on all of it. Here's how they fit together.

## omp: the runtime

[omp](https://github.com/) (oh-my-pi) is the agent runtime. It's the thing that turns "an LLM plus some API keys" into an agent that can use tools, run shell commands, spawn sub-agents, and follow skills.

I run one agent per **profile**. On my homelab those are separate LXC containers — one per person the agent works for — each with its own memory, its own credentials, and its own personality file. A profile isn't just a system prompt; it's a whole workspace: cron jobs, hooks, a kanban board, cached media, and a `SOUL.md` that sets its voice.

Each profile also has a **gateway** — the channel it talks to me over. Mine run a WhatsApp bridge, so I message the agent the same way I'd message a person. The runtime handles the model calls, tool use, and streaming; the gateway just carries the conversation.

The important part is that omp is boring infrastructure. It starts, it stays up, it restarts if it falls over. The interesting behaviour comes from the profile and the skills, not from babysitting the process.

## Obsidian: the agent's memory

An agent is only as good as what it remembers. I use **Obsidian** for that, and omp ships a note-taking skill that reads and writes an Obsidian vault directly. So the agent's notes and my notes are the *same* notes — plain Markdown files in one vault, not some opaque memory blob I can't inspect.

To keep that vault in sync across my phone, my laptop, and the agent, I self-host **Obsidian LiveSync** on CouchDB. Every device — including the box the agent runs on — points at the same CouchDB, and changes replicate around in near real time. I capture something on my phone, the agent sees it; the agent writes a summary, it shows up in my vault.

Two things I like about this:

- **It's plain Markdown.** If I turn every agent off tomorrow, my notes are still just files. Nothing is locked inside a model.
- **I can read and edit the agent's memory by hand.** When it gets something wrong, I open the note and fix it, and that's the correction.

CouchDB stays inside the homelab and is reached the same way as everything else here — through the tunnel, never exposed directly.

## omp-lb: the dashboard

Once you're running more than one agent, you lose track of them. Which account is a profile using? How much quota is left? What did it actually do last night? [omp-lb](https://github.com/inhumanxd/omp-lb) answers that.

It's a small local control panel for omp — zero dependencies, no build step, a single Node server you point at omp's data directory. It reads omp's own databases (`agent.db`, and read-only `history.db` / `models.db`) and gives you a browser view:

- **Accounts** — every credential, enable/disable, "use only this one", and the real per-account quota pulled from omp's cached usage reports.
- **Sessions** — full prompt history by session, with full-text search. This is the one I use most: I can go back and read exactly what an agent was asked and how it answered.
- **Models** — the model catalogue per provider and which role each model is wired to.
- **Usage** — a prompt-activity chart and per-account quota windows, so I can see which profile is burning through what.
- **System** — theme, cache inspector, and backups.

A few details that matter:

- It **binds to `127.0.0.1` only**. Nothing is exposed; each person runs their own instance. Start it with `node server.mjs` and open `http://127.0.0.1:8787`.
- It **does not proxy your requests.** It's a read/manage layer over omp's local data, not a gateway in the request path. Quota comes from omp's own cached reports.
- Every write takes a `VACUUM INTO` snapshot next to `agent.db`, keeping the last 20 — so a bad toggle is one file-copy away from undone.
- omp loads credentials at session start, so account changes apply to **new** sessions; restart the agent to pick them up mid-flight.

If your omp data isn't in the default `~/.omp/agent`, point `OMP_DB` at the right `agent.db` and it picks up the siblings from the same folder.

## The loop

Put together, it's a tidy loop. I capture into Obsidian. The **omp** agent reads and writes that same vault through its Obsidian skill, and answers me over WhatsApp. **LiveSync** keeps the vault identical everywhere. And **omp-lb** sits off to the side, read-only on the request path, showing me what the agents are doing — sessions, quota, usage — so nothing is a black box.

None of the three is doing anything flashy on its own. It's the combination that turns "some API keys" into agents I actually trust to run unattended at home.
