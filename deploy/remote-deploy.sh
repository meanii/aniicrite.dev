#!/usr/bin/env bash
# Run on the server (piped via SSH by the deploy workflow). Pulls the latest
# main, rebuilds the container, and restarts it. The server-side compose.yaml
# (build: ./repo) and .env live in APP_DIR, outside the repo checkout.
set -euo pipefail

APP_DIR="/root/personal/aniicrite.dev"
cd "$APP_DIR"

git -C repo fetch --quiet origin main
git -C repo reset --quiet --hard origin/main

docker compose up -d --build
docker image prune -f >/dev/null 2>&1 || true

echo "deployed $(git -C repo rev-parse --short HEAD)"
