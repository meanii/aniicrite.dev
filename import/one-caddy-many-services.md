---
title: A dozen services behind one Caddy
slug: one-caddy-many-services
date: 2026-08-06T00:00:00Z
tags: Caddy, self-hosting, VPS, networking
status: published
summary: One Caddy on a small VPS is the only thing listening on 80 and 443. Every service I host, local or tunnelled from home, is a few lines in a single Caddyfile.
---
One VPS fronts almost everything I run in public, and Caddy is the only process on it listening on ports 80 and 443. Every service is a block in a single Caddyfile.

I keep choosing Caddy because of certificates. I have not thought about TLS in years. A site block is a hostname and a `reverse_proxy` line, and the certificate gets issued and renewed without me doing anything. For the wildcard hostnames I use the Cloudflare DNS plugin, so those are issued over DNS and nothing has to be exposed for the challenge.

## Two kinds of backend

Most of the blocks point at Docker containers on the VPS itself. Each container publishes a port on 127.0.0.1 and Caddy proxies to it:

```
memos.example.dev {
    reverse_proxy 127.0.0.1:5230
}
```

That covers a couple of Ghost blogs, NocoDB, Memos, Zennotes, a PDF tool, a NetBird control plane, this site and some demo apps.

The other blocks are for services that live on the homelab and reach the VPS through an [frp tunnel](/posts/frp-reverse-tunnel/). Those proxy to frp's vhost port instead of a local container, so a `*.home` hostname resolves to something on the box at home without that box being exposed to anything.

## One file

Having the whole thing in one file sounds like a liability and turns out to be the opposite. I can read my entire public surface on one screen: every hostname and where it goes, nothing hidden in a dozen config directories. Adding a service is three lines and a reload. Caddy validates the new config before switching to it and keeps the old one running if the new one is broken, so a typo does not take everything down.

It is the least clever part of my setup and the part I have had the fewest problems with.
