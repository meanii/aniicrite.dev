---
title: Single sign-on for my homelab with PocketID
slug: pocketid-sso
date: 2026-08-08T00:00:00Z
tags: homelab, SSO, OIDC, self-hosting
status: published
summary: One small OIDC provider with passkeys, so I stop keeping a separate login for every self-hosted app.
---
Every self-hosted app wants its own account. A dozen apps means a dozen logins, a dozen password resets, and a dozen places to get security wrong. I put [PocketID](https://github.com/pocket-id/pocket-id) in front of them to fix that.

PocketID is a small OIDC provider built around passkeys. It runs in its own container on port `1411`, and it's the only thing holding identity in the homelab. Any app that speaks OpenID Connect points at it as the identity provider instead of managing its own users.

### Why PocketID specifically

- It's tiny and does one job — OIDC, nothing else.
- It's passkey-first, so there's no master password to phish; I log in with the device I'm already holding.
- It's easy to self-host: a single container and a database.

### How it fits together

Each OIDC-capable app gets a client entry in PocketID — a client ID, a secret, and a redirect URL. When I open the app, it bounces me to PocketID, I approve with a passkey, and I'm back in the app signed in. New device? I enrol one passkey in PocketID and every app follows.

For apps that don't support OIDC, I keep them behind the reverse proxy and don't expose them further. Not everything needs to be on the internet.

It's a small change but it's the one that made running this many services feel manageable instead of like a pile of separate accounts I'm slowly losing track of.
