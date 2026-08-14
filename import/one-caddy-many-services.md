---
title: A dozen services behind one Caddy
slug: one-caddy-many-services
date: 2026-08-06T00:00:00Z
tags: Caddy, self-hosting, VPS, networking
status: published
summary: One Caddy instance on a small VPS routes every service I host, with automatic TLS and a short config.
---
One VPS fronts almost everything I run in public, and Caddy is the only thing listening on 80 and 443. Every service is a block in one Caddyfile.

The reason I keep coming back to Caddy is TLS. I don't think about certificates. A site block is a hostname and a `reverse_proxy` line, and the cert just appears and renews itself. For wildcard hosts I use the Cloudflare DNS plugin so certs issue over DNS without exposing anything.

### Two kinds of backend

The blocks fall into two groups.

Most are local Docker containers. Each app publishes a port on `127.0.0.1`, and Caddy proxies to it:

```
memos.example.dev {
    reverse_proxy 127.0.0.1:5230
}
```

That covers the things that live on the VPS itself — a couple of Ghost blogs, NocoDB, Memos, Zennotes, a PDF tool, a NetBird control plane, this site, and some demo apps.

The rest live at home and reach the VPS through an frp tunnel. Those blocks proxy into frp's vhost port instead of a local container, so `*.home` hostnames resolve to services on the home box without any of them being exposed directly.

### Why one config

Everything being in one file sounds fragile but it's the opposite. I can read the entire public surface of my setup in one screen: every hostname, where it goes, nothing hidden. Adding a service is three lines and a reload. Caddy validates the config on reload and keeps the old one if the new one is broken, so a typo doesn't take everything down.

It's the least clever part of my setup and the one I worry about least.
