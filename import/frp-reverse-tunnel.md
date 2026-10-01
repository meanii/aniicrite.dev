---
title: Reaching my homelab through a cheap VPS with frp
slug: frp-reverse-tunnel
date: 2026-08-07T00:00:00Z
tags: homelab, frp, networking, self-hosting
status: published
summary: My home server has no open ports and no static IP. It dials out to a small VPS running frp, and the VPS is the only thing on the public internet.
---
My homelab is behind an ordinary home router. No port forwarding, no static IP, and the ISP is not going to give me one. I still want to reach a few things on it from outside. Instead of opening ports I use [frp](https://github.com/fatedier/frp), which is a reverse tunnel.

Reverse here means the direction of the connection is flipped. Nothing at home listens on the internet. The home box makes an outbound connection to a VPS I rent, and the VPS accepts public traffic and sends it back down that connection.

## The two halves

The VPS runs `frps`, the server side. It listens on port 7000 for the control connection from home, and on 8009 and 8010 for HTTP and HTTPS, where it routes by hostname.

At home `frpc` runs as a small service. It connects out to port 7000 and registers the hostnames it wants to serve. When a request for one of those hostnames arrives at the VPS, frp pushes it down the existing tunnel and the home box answers. The router at home never sees an inbound connection.

## Caddy in front

DNS does not point at frp directly. Caddy on the VPS is the public edge. It terminates TLS and reverse proxies the home hostnames into frp's vhost port, the same way it proxies to the containers that run on the VPS itself. Wildcard certificates are issued through the Cloudflare DNS challenge, so adding a service at home is a line in the Caddyfile and nothing else.

At home there is one local reverse proxy in front of everything, so frp only ever talks to a single backend.

## Why not Tailscale or Cloudflare Tunnel

I use a mesh VPN for getting into the machine myself. For public services I wanted the relay to be a box I own, running software I can read, with one process on each side. frp is one static binary and a short TOML file at both ends.

The part I care about is that the VPS is disposable. If it dies I rent another, copy the same `frps.toml` onto it, point the DNS name at it, and the home client reconnects on its own. The home machine holds the data and the VPS holds nothing.
