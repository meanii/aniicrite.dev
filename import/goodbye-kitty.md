---
title: Goodbye kitty
slug: goodbye-kitty
date: 2024-07-24T00:00:00Z
tags: tmux, terminal, Linux
status: published
summary: Why I dropped the Kitty terminal and went back to tmux.
---
I used Kitty for about two years. It was fast and it mostly stayed out of my way, so I stuck with it. A couple of things finally made me switch back.

### The bugs

Version 0.24.2 broke a few things I hit every day:

- Text stops rendering when the terminal is resized.
- The cursor flips to an I-beam even though I've set it to a block.
- New text draws in the wrong place.

I could have waited for a fix. What actually pushed me out was reading the author's take on tmux in the FAQ.

### I need tmux

My whole setup runs on tmux. It has saved my work more times than I can count — close a terminal by accident and the session is still there when I come back. Kitty just exits. For me that's the difference between an annoyance and losing an afternoon.

### The FAQ

Kitty's author is openly dismissive of multiplexers like tmux, down to telling users to "go soak your head." I don't want to fight my terminal's author about how I work.

### A few more things

- Scrollback is limited because Kitty won't cache to disk.
- Same reason, it uses more RAM than I'd like.
- Bugs, including security ones, sit unacknowledged.

So I moved back to tmux with a plain terminal. tmux gets called a "hack" in that same FAQ, but it does things Kitty can't — like keeping remote sessions alive — and it hasn't lost my work once.

Goodbye, Kitty.

---

Related: [Gavin Howard's post on the same decision](https://gavinhoward.com/2022/02/goodbye-kitty/).
