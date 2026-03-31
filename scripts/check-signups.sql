-- Run daily to find new signups in the last 24 hours.
-- Usage: psql $DATABASE_URL -f scripts/check-signups.sql
-- The email column is included so you can send the personal follow-up directly.
SELECT
    id,
    email,
    created_at AT TIME ZONE 'Europe/Berlin' AS created_local,
    mqtt_username
FROM users
WHERE created_at > NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC;
