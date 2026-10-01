---
title: Persisting tmux sessions across reboots
slug: persistent-tmux-environments
date: 2024-01-26T00:00:00Z
tags: tmux, Linux, DevOps
status: published
summary: tmux-resurrect saves your tmux windows and panes to disk and brings them back after a reboot.
---
I run everything inside tmux, and for a long time a reboot wiped all of it. Every window, every pane layout, every working directory, gone. tmux-resurrect fixes that. It writes the session layout to disk and restores it after a restart.

I recorded a short video of the setup:

[![tmux persistence tutorial](https://img.youtube.com/vi/4pMxsNanc_g/0.jpg)](https://youtube.com/shorts/4pMxsNanc_g)

The short version. Install the plugin and add it to your `tmux.conf`. Save with `prefix + Ctrl-s`. After a reboot, restore with `prefix + Ctrl-r`, or configure it to restore on its own when tmux starts.

Links:

- [tmux-resurrect](https://github.com/tmux-plugins/tmux-resurrect)
- [my tmux.conf](https://github.com/meanii/dotfiles/blob/main/dot_config/tmux/tmux.conf)
