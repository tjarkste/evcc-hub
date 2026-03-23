#!/bin/bash
set -euo pipefail

DB_PATH="/opt/evcc-cloud/data/evcc.db"
BACKUP_DIR="/tmp/evcc-backups"
REMOTE="hetzner:evcc-cloud-backups"
DATE=$(date +%Y-%m-%d)
DAY_OF_WEEK=$(date +%u)  # 1=Mon, 7=Sun

mkdir -p "$BACKUP_DIR"

BACKUP_FILE="$BACKUP_DIR/evcc-backup-$DATE.tar.gz"
tar -czf "$BACKUP_FILE" -C "$(dirname "$DB_PATH")" "$(basename "$DB_PATH")"

# Daily backups — keep 7 days
rclone copy "$BACKUP_FILE" "$REMOTE/daily/"
rclone delete --min-age 8d "$REMOTE/daily/"

# Weekly backups on Sunday — keep 4 weeks
if [ "$DAY_OF_WEEK" = "7" ]; then
  rclone copy "$BACKUP_FILE" "$REMOTE/weekly/"
  rclone delete --min-age 29d "$REMOTE/weekly/"
fi

rm -f "$BACKUP_FILE"
echo "Backup completed: $DATE"
