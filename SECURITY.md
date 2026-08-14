# Security Policy

## Reporting a vulnerability

Please **do not** open a public issue for security problems. Instead, use
GitHub's [private vulnerability reporting](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing-information-about-vulnerabilities/privately-reporting-a-security-vulnerability)
(Security → Report a vulnerability) on this repository, or email the maintainer.

You'll get an acknowledgement within a few days. Please include reproduction
steps and the affected version/commit.

## Security posture

The application ships with these defenses:

- **Admin auth** — single password stored only as an argon2id hash; login is
  rate-limited per IP (5 failures / 15 min).
- **Sessions** — signed + encrypted cookies (`gorilla/securecookie`),
  `HttpOnly`, `SameSite=Lax`, and `Secure` when `APP_ENV=production`.
- **Comment identity** — GitHub OAuth with CSRF state validation; a block-list
  removes and bans abusive users.
- **Response headers** — `Content-Security-Policy`, `X-Content-Type-Options`,
  `X-Frame-Options`, `Referrer-Policy`, and HSTS in production.
- **Uploads** — extension allow-list (no SVG) plus content-type sniffing.

## Operator responsibilities

- Set strong `SESSION_HASH_KEY` / `SESSION_BLOCK_KEY` (32 random bytes each).
- Serve behind TLS (e.g. Caddy) and set `APP_ENV=production`.
- Keep the admin panel reachable only as needed.
