#!/usr/bin/env bash
# Consistent SQLite + uploads backup. Schedule via cron, e.g.:
#   0 3 * * *  /opt/aniicrite/deploy/backup.sh >> /var/log/aniicrite-backup.log 2>&1
set -euo pipefail

DATA_DIR="${DATA_DIR:-/opt/aniicrite/data}"
DEST="${BACKUP_DIR:-/opt/aniicrite/backups}"
KEEP="${BACKUP_KEEP:-14}"
DB="$DATA_DIR/site.db"

mkdir -p "$DEST"
ts="$(date +%Y%m%d-%H%M%S)"

# .backup takes a live-consistent snapshot even while the server is running.
sqlite3 "$DB" ".backup '$DEST/site-$ts.db'"

# Snapshot uploaded images if present.
if [ -d "$DATA_DIR/uploads" ]; then
	tar czf "$DEST/uploads-$ts.tar.gz" -C "$DATA_DIR" uploads
fi

# Retain only the most recent $KEEP of each.
ls -1t "$DEST"/site-*.db 2>/dev/null      | tail -n +$((KEEP + 1)) | xargs -r rm -f
ls -1t "$DEST"/uploads-*.tar.gz 2>/dev/null | tail -n +$((KEEP + 1)) | xargs -r rm -f

echo "backup complete: site-$ts.db"
