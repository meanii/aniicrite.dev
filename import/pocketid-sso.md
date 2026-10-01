---
title: Single sign-on for my homelab with PocketID
slug: pocketid-sso
date: 2026-08-08T00:00:00Z
tags: homelab, SSO, OIDC, self-hosting
status: published
summary: PocketID is a small OIDC provider with passkey login. I put it in front of the self-hosted apps so there is one login instead of one per app.
---
Every self-hosted app wants its own user table. After a dozen apps that is a dozen logins, each with its own password reset flow and its own way of getting security wrong. I put [PocketID](https://github.com/pocket-id/pocket-id) in front of them so there is one login.

PocketID is a small OpenID Connect provider built around passkeys. It runs in its own container on port 1411 and it is the only thing in the homelab that holds identity. Any app that speaks OIDC is pointed at it as the identity provider and stops managing its own users.

## Why this one

It is small and does one job. It is passkey-first, so there is no master password to type and nothing to phish. I log in with the laptop or phone I am already holding. And it self-hosts without fuss: one container and a database.

## How it fits

Each app gets a client entry in PocketID, which is a client ID, a secret and a redirect URL. Opening the app bounces me to PocketID, I approve with a passkey, and I am back in the app signed in. When I get a new device I enrol one passkey in PocketID and every app follows.

Apps that do not support OIDC stay behind the reverse proxy on the LAN and are not exposed any further. Not everything needs to be reachable from outside.

It is a small change. It is also the one that made running this many services feel like one system instead of a pile of accounts I was slowly losing track of.
