---
title: What's running in my homelab
slug: my-homelab
date: 2026-08-10T00:00:00Z
tags: homelab, Proxmox, Linux, self-hosting
status: published
summary: One i5-8500T box running Proxmox, one LXC container per service, and a cheap VPS doing the public-facing part over frp.
---
My homelab is one machine. It is an Intel i5-8500T with 6 cores and 23 GB of RAM running Proxmox VE, sitting on a shelf at home. It is quiet, it draws very little power, and it does more than I expected when I bought it.

Every service gets its own LXC container. Containers on Proxmox cost almost nothing, snapshots take a second, and when I break something during an upgrade the damage is limited to one service. This is what is on it at the moment:

- AdGuard Home for DNS and ad blocking for the whole house
- Immich for photo backup from my phone, instead of paying Google for storage
- Vaultwarden for passwords, which works with the Bitwarden clients
- Calibre for the ebook library
- PocketID, a small OIDC provider, for single sign-on to the apps above
- Nginx Proxy Manager for routing and certificates inside the LAN
- mediastack, which is the media server and the downloaders
- downly, a Telegram download bot I wrote
- a few hermes agents, one container per profile

Storage is in two pools. The OS and most of the containers live on an LVM-thin pool on the fast disk. A second pool of about 490 GB, which I named slowbird, holds the big stuff like photos and media.

The box has no ports open to the internet. For the handful of services I want to reach from outside, it makes an outbound connection to a cheap VPS using frp, and the VPS reverse proxies those hostnames back down the tunnel. I wrote that up in [its own post](/posts/frp-reverse-tunnel/). DNS is on Cloudflare. If the VPS disappears I rent another one and point frp at it, and nothing at home has to change.

That is all of it. I would rather the machine that holds my photos and passwords be boring.
