---
title: What's running in my homelab
slug: my-homelab
date: 2026-08-10T00:00:00Z
tags: homelab, Proxmox, Linux, self-hosting
status: published
summary: A single i5-8500T box running Proxmox, one LXC per service, tunnelled out through a cheap VPS.
---
My homelab is one small machine: an Intel i5-8500T, 6 cores, 23 GB of RAM, running Proxmox VE. Nothing exotic. It sits at home, stays quiet, and does a surprising amount of work.

I run one LXC container per service instead of stacking everything on one host. Containers are cheap on Proxmox, snapshots are quick, and if I break one I haven't taken the rest down with it. Right now the list looks like this:

- **AdGuard Home** — DNS and ad-blocking for the whole house.
- **Immich** — photo backup, so I'm not renting space from Google for it.
- **Vaultwarden** — passwords, Bitwarden-compatible.
- **Calibre** — ebook library.
- **PocketID** — a small OIDC provider for single sign-on across the other apps.
- **Nginx Proxy Manager** — routing and certs inside the LAN.
- **mediastack** — media server and downloaders.
- **downly** — my Telegram download bot.
- A few **hermes** agents, one per profile.

Storage is split: the OS and most containers live on the fast LVM-thin pool, and a bigger ~490 GB pool (I named it `slowbird`) holds the bulk stuff like photos and media.

The one piece that isn't at home is public access. The box has no ports open to the internet. Instead it dials out to a cheap VPS over [frp](https://github.com/fatedier/frp), and the VPS reverse-proxies the handful of services I want to reach from outside. DNS runs through Cloudflare. So the VPS is disposable — if it dies I spin up another and point frp at it, and nothing on the home box changes.

That's the whole setup. It's boring on purpose. Boring is what I want from the machine that holds my photos and passwords.
