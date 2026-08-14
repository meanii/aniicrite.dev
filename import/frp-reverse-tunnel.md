---
title: Reaching my homelab through a cheap VPS with frp
slug: frp-reverse-tunnel
date: 2026-08-07T00:00:00Z
tags: homelab, frp, networking, self-hosting
status: published
summary: My home box has no open ports. It dials out to a VPS with frp, and the VPS handles everything public-facing.
---
My homelab sits behind a normal home router with no port forwarding and no static IP. I still want to reach some of it from outside. I do that with [frp](https://github.com/fatedier/frp) instead of opening ports.

The idea is a reverse tunnel. Nothing at home listens on the public internet. Instead the home box makes an outbound connection to a server I control, and that server accepts the public traffic and passes it back down the tunnel.

### The two sides

On a cheap VPS I run `frps`, the frp server:

- port `7000` — the control connection the client dials into
- ports `8009` / `8010` — the HTTP and HTTPS vhost ports frp multiplexes by hostname

At home, `frpc` runs as a small service and connects out to that `7000`. Once it's connected, it registers the hostnames I want to expose. A request for one of those hostnames hits the VPS, frp forwards it down the existing tunnel, and the home box answers — no inbound firewall hole anywhere.

### Where Caddy fits

I don't point DNS straight at frp. Caddy on the VPS is the public edge and terminates TLS, then reverse-proxies the home hostnames into frp's vhost port. Wildcard certs come from Cloudflare DNS, so a whole `*.home` space resolves without me touching the config each time I add a service. At home, everything sits behind a local reverse proxy, so from the tunnel's point of view there's just one backend to talk to.

### Why not Tailscale or Cloudflare Tunnel

Both are fine and I use a mesh VPN for other things. For this I wanted the relay to be a box I own, running software I can read, with one moving part on each side. frp is a single static binary at both ends and a short TOML config. If the VPS dies I rent another, drop the same `frps.toml` on it, and the home client reconnects — nothing on the home box changes.

That last part is the whole point. The home machine holds the data and stays sealed; the VPS is disposable plumbing.
