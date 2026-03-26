#!/bin/bash
set -euo pipefail

BACKUP_DIR="/opt/evcc-hub/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
DAY_OF_WEEK=$(date +%u)

mkdir -p "$BACKUP_DIR/daily" "$BACKUP_DIR/weekly"

# Dump database via Docker
docker exec evcc-hub-postgres-1 pg_dump -U evcc evcc_hub > "$BACKUP_DIR/daily/evcc_hub_${TIMESTAMP}.sql"

# Keep daily backups for 7 days
find "$BACKUP_DIR/daily" -name "*.sql" -mtime +7 -delete

# Weekly backup on Sunday (keep 4 weeks)
if [ "$DAY_OF_WEEK" -eq 7 ]; then
    cp "$BACKUP_DIR/daily/evcc_hub_${TIMESTAMP}.sql" "$BACKUP_DIR/weekly/"
    find "$BACKUP_DIR/weekly" -name "*.sql" -mtime +28 -delete
fi

echo "Backup completed: evcc_hub_${TIMESTAMP}.sql"
