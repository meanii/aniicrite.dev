---
title: I built a webhook debugger for local development
slug: webhooklocal
date: 2026-10-02T00:00:00Z
tags: Go, webhooks, Stripe, HTMX, side project
status: published
summary: webhooklocal gives you a private URL that never changes, stores every webhook, forwards it to localhost, and lets you edit and replay any request. It is in private beta and I am looking for testers.
---
Every time I wired up Stripe or GitHub webhooks on a new project I went through the same loop: start ngrok, copy the new URL into the provider's dashboard, trigger an event, watch it fail, fix the handler, and then wait for the provider to send the event again. Restart the laptop and the URL changes, so the first step repeats. After doing this enough times I built the tool I kept wishing for. It is live at [webhooklocal.com](https://webhooklocal.com).

## What it does

- **A private URL that never changes.** Paste it into Stripe once. It survives restarts, reboots and new laptops. You can rotate it if it leaks.
- **Every request is stored.** Method, path, headers and body, with the body pretty-printed and secrets like `Authorization` masked by default.
- **Forwarded to localhost as it arrives.** A small CLI, `wl listen 3000`, connects out from your machine over a WebSocket, so it works behind NAT, VPNs and office firewalls. Nothing to open, no tunnel URL to copy around.
- **Edit and replay.** Change a field, drop a header, send the request to your app again. If your laptop was closed when the event came in, it is kept and marked not sent, ready to replay later.
- **Your app's answer goes back to the provider.** The status and body your handler returns are passed through, so the provider's retry behaviour is the same as in production.

Setup is three commands:

```sh
curl -fsSL https://webhooklocal.com/install.sh | sh
wl login
wl listen 3000
```

The install script checks the download against a published checksum and never asks for `sudo`.

## How it is built

The whole thing is one Go binary: `net/http`, `html/template` and HTMX for the dashboard, Postgres for storage, and the same binary serves the CLI's WebSocket. No JavaScript framework, no build step beyond the Tailwind CLI. It runs in a container on a single Hetzner server behind Caddy and Cloudflare, configured only through an Ansible playbook so the repository is the source of truth for the machine. Deploys go out from GitHub Actions after the tests pass.

Sign-in is a GitHub button or an emailed link, no passwords. The sign-in form sits behind rate limits, a honeypot field and Cloudflare Turnstile, because every sign-in costs an email and I did not want the sender address burned by a bot before the first real user showed up.

Two things I spent more time on than I expected:

- **Passing localhost's response through safely.** The first version copied your app's response headers back to the provider. That meant a `Set-Cookie` or `Location` from localhost could land on the webhooklocal origin. Now only the status and body go through, with a safe content type and a sandboxed content security policy, so a webhook URL can never serve a page that runs in a browser.
- **Making the UI not look generated.** The first design had the usual blurred glass panels and purple glows. I redid it flat: near-black, hairline borders, two greys of text, a serif only for page titles, and a real slice of the dashboard as the hero instead of an illustration.

## What it is not, yet

The Free plan has one URL, 100 stored requests and 24 hours of history, which is enough for one integration at a time. Paid plans with longer history and more URLs are listed but checkout is not open. Signature checks for Stripe and GitHub, replay to a public URL and team accounts are planned, not built. The dashboard says so wherever that matters.

## I am looking for testers

If you integrate Stripe, GitHub or Shopify webhooks and want to try this on one real integration this week, [sign up](https://webhooklocal.com/signin) and then tell me what was annoying. The account can be deleted in one click from settings. Email [hello@webhooklocal.com](mailto:hello@webhooklocal.com) or leave a comment below, and if you have fifteen minutes for a call about how you debug webhooks today, that is worth more to me than any sign-up.
