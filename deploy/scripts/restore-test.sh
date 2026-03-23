#!/bin/bash
set -euo pipefail

REMOTE="hetzner:evcc-cloud-backups"
RESTORE_DIR="/tmp/evcc-restore-test"

mkdir -p "$RESTORE_DIR"

echo "=== Available backups ==="
rclone ls "$REMOTE/daily/"

echo ""
echo "=== Downloading latest backup ==="
LATEST=$(rclone ls "$REMOTE/daily/" | sort | tail -1 | awk '{print $2}')
rclone copy "$REMOTE/daily/$LATEST" "$RESTORE_DIR/"

echo ""
echo "=== Extracting ==="
tar -xzf "$RESTORE_DIR/$LATEST" -C "$RESTORE_DIR/"

echo ""
echo "=== Verifying SQLite integrity ==="
sqlite3 "$RESTORE_DIR/evcc.db" "PRAGMA integrity_check;"

echo ""
echo "=== Done. Files in $RESTORE_DIR: ==="
ls -lh "$RESTORE_DIR/"
