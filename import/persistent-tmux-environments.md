---
title: Persisting tmux sessions across reboots
slug: persistent-tmux-environments
date: 2024-01-26T00:00:00Z
tags: tmux, Linux, DevOps
status: published
summary: Restore your tmux windows and panes after a restart with tmux-resurrect.
---
tmux keeps my terminal sessions organized, but a reboot used to wipe all of it — every window and pane gone. `tmux-resurrect` fixes that: it saves the session to disk and restores it after a restart, layout and all.

I made a short video walking through the setup:

[![tmux persistence tutorial](https://img.youtube.com/vi/4pMxsNanc_g/0.jpg)](https://youtube.com/shorts/4pMxsNanc_g)

The gist:

- Install the plugin and add it to your `tmux.conf`.
- Save a session with the resurrect keybind (`prefix + Ctrl-s`).
- Restore it after a reboot (`prefix + Ctrl-r`), or set it to restore automatically.

Links:

- Plugin: [tmux-resurrect](https://github.com/tmux-plugins/tmux-resurrect)
- My config: [tmux.conf](https://github.com/meanii/dotfiles/blob/main/dot_config/tmux/tmux.conf)
