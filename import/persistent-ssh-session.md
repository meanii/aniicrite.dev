---
title: Keeping SSH sessions alive with autossh
slug: persistent-ssh-session
date: 2023-11-04T00:00:00Z
tags: SSH, autossh, Linux
status: published
summary: assh — a small wrapper that reconnects your SSH session when the network drops.
---
![ssh-session](https://miro.medium.com/v2/resize:fit:1400/format:webp/1*iVYccCsXIgpYvmC3zMnLBA.png)

I work over SSH a lot, and flaky networks drop the connection at the worst possible time. `assh` is a small wrapper I use that reconnects automatically when the link goes down, so I don't have to babysit the session.

[Watch the demo](https://www.youtube.com/watch?v=EAjosu4AVGQ)

## Installation

```bash
curl --silent -o- https://raw.githubusercontent.com/meanii/assh/main/install.sh | sudo bash
```

It needs sudo because it installs the script to `/usr/local/bin/assh`.

## Usage

```bash
assh <ssh-connection-string>
```

Source: <https://github.com/meanii/assh>
