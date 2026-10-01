---
title: Keeping SSH sessions alive with autossh
slug: persistent-ssh-session
date: 2023-11-04T00:00:00Z
tags: SSH, autossh, Linux
status: published
summary: assh is a small wrapper around autossh that reconnects an SSH session when the network drops.
---
![ssh-session](https://miro.medium.com/v2/resize:fit:1400/format:webp/1*iVYccCsXIgpYvmC3zMnLBA.png)

I spend a lot of the day in SSH sessions and my network is not reliable. The connection drops, the terminal hangs, and I reconnect by hand. I got tired of that and wrote assh, a small wrapper script that uses autossh to reconnect automatically when the link goes down.

There is a short [demo video](https://www.youtube.com/watch?v=EAjosu4AVGQ).

## Install

```bash
curl --silent -o- https://raw.githubusercontent.com/meanii/assh/main/install.sh | sudo bash
```

It asks for sudo because it copies the script to `/usr/local/bin/assh`. The script is short, so read it first if you prefer.

## Use

```bash
assh <ssh-connection-string>
```

Pass it the same connection string you would give `ssh`.

The source is at [github.com/meanii/assh](https://github.com/meanii/assh).
