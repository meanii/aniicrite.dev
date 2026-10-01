---
title: Goodbye kitty
slug: goodbye-kitty
date: 2024-07-24T00:00:00Z
tags: tmux, terminal, Linux
status: published
summary: After two years on Kitty I went back to tmux in a plain terminal. Some rendering bugs started it and the author's FAQ finished it.
---
I used Kitty for about two years. It was fast, the font rendering was good, and it stayed out of my way, which is most of what I want from a terminal. I have now gone back to a plain terminal with tmux, and this is why.

## The bugs

Version 0.24.2 broke a few things I hit every single day. Text stopped rendering after a resize until I pressed a key. The cursor turned into an I-beam even though my config says block. New output sometimes drew in the wrong place. None of these is fatal on its own and I could have waited for a fix. They were enough to make me go and read the FAQ, which is where the real reason is.

## The FAQ

My whole workflow runs inside tmux. I close terminals by accident, my laptop goes to sleep, SSH connections drop, and the session is still there when I come back. That has saved me real work more times than I can count. Kitty's author thinks multiplexers are a bad idea and says so in the FAQ in fairly blunt terms, down to telling people who want tmux-style behaviour to "go soak your head". Kitty has its own windows and tabs that cover some of what tmux does, but not the thing I need most, which is a session that survives the terminal closing or the remote connection dropping.

I do not want to use a tool whose author is this openly opposed to how I work. It means the bugs that affect my setup are not going to be a priority, and it already felt that way.

## Smaller things

Scrollback is capped because Kitty will not spill it to disk, and for the same reason it uses more memory than I would like with a lot of tabs open. A few bugs, including security reports, had been sitting without a response for a while.

So I am back on tmux. The same FAQ calls it a hack. It is a hack that keeps my remote sessions alive and has not lost my work once.

Gavin Howard wrote about [making the same switch](https://gavinhoward.com/2022/02/goodbye-kitty/) for similar reasons.
