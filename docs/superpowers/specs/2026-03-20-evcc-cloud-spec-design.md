# evcc-cloud — System Specification

## Overview

evcc-cloud is a cloud-based EV charging management system that allows users to monitor and control their local evcc instances remotely. It provides multi-tenant, multi-site support with real-time data via MQTT.

**Audience:** Developers contributing to the project.
**Status:** Proof-of-concept — core pieces work but are not production-ready.

---

## Current System (As-Is)

### Architecture

Three-tier cloud application:

- **Frontend:** Vue 3 SPA (forked from upstream evcc UI) in `web/assets/js/`, communicates via MQTT over WebSocket
- **Backend:** Go/Gin HTTP server in `backend/`, handling auth (JWT) and Mosquitto auth/ACL plugins
- **Broker:** Eclipse Mosquitto for real-time pub/sub between evcc instances and the frontend

### Data Model (Current)

Single `users` table in SQLite:

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| email | string | User email |
| password_hash | string | bcrypt hash |
| mqtt_username | string | MQTT credentials |
| mqtt_password | string | MQTT credentials (stored plaintext) |
| topic_prefix | string | `user/{id}/evcc/` |
| created_at | datetime | Row creation timestamp |

No concept of "sites" — one evcc instance per user. MQTT passwords are stored in plaintext for Mosquitto auth plugin comparison.

### Auth Flow

1. User registers/logs in via REST — receives JWT (24h, no refresh) + MQTT credentials
2. Frontend stores everything in localStorage
3. Frontend connects to Mosquitto via WebSocket using MQTT credentials
4. Mosquitto validates credentials via backend HTTP plugin (`/api/mqtt/auth`, `/api/mqtt/acl`)
5. Frontend subscribes to `{topicPrefix}/#` and maps incoming messages to the Vue store

### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/api/auth/register` | User registration |
| POST | `/api/auth/login` | User login |
| POST | `/api/mqtt/auth` | Mosquitto auth plugin |
| POST | `/api/mqtt/acl` | Mosquitto ACL plugin |

### Known Limitations

- No token refresh — JWT expires after 24h, user must re-login
- No error boundaries or reconnection logic in the MQTT client
- Simulator-only — no validation against real evcc MQTT payloads
- Single-site per user — no multi-instance support
- No HTTPS enforcement, no rate limiting, no input validation beyond basic checks
- Secrets (JWT_SECRET) managed via env vars with no rotation strategy
- MQTT passwords stored in plaintext (required for Mosquitto auth plugin comparison)
- ACL only allows writes to `/set` topics — incompatible with real evcc instances that publish to arbitrary data topics
- SQLite is fine for PoC but may need evaluation at scale
- No auth middleware on any endpoint — all routes are either public or called by Mosquitto

### Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Vue 3, TypeScript, Vite, Bootstrap 5, mqtt.js |
| Backend | Go 1.22, Gin, SQLite, golang-jwt |
| Broker | Eclipse Mosquitto v2 |
| Infra | Docker, Nginx, GitHub Actions |

---

## Target Data Model

> **Note:** This section describes the **end-state schema** across all phases. Phase 1 creates the `sites` table and migrates user data. The `refresh_tokens` table is created in Phase 2.

### Core Entities

```
User (1) ──→ (N) Site ──→ (1) MQTT credential set (for evcc instance)
User (1) ──→ (1) MQTT credential set (for frontend/read access)
```

**Users** — account identity with a dedicated read-only MQTT credential for the frontend.

**Sites** — represents a single evcc instance (e.g., "Home", "Vacation House"):
- Belongs to one user
- Has its own MQTT credentials used by the local evcc instance to publish data
- Has a display name and optional metadata (location, timezone)

### MQTT Credential Model

Two types of MQTT credentials exist in the target system:

- **Site credentials** (on `sites` table): Used by the local evcc instance to publish data to `user/{userId}/site/{siteId}/evcc/#`. Has **write** access to its site topic subtree.
- **User credentials** (on `users` table): Used by the frontend to connect to the broker. Has **read** access across all the user's sites (`user/{userId}/site/+/evcc/#`) and **write** access only to `/set` topics (for control commands).

### Topic Structure

Current: `user/{userId}/evcc/{path}`
Target: `user/{userId}/site/{siteId}/evcc/{path}`

Each evcc instance publishes to its own isolated topic namespace.

### ACL Rules (Target)

> **Note:** These are logical rules enforced by the Go ACL plugin (`/api/mqtt/acl`), not literal Mosquitto ACL file patterns. The plugin uses prefix matching and suffix checks in code.

| Client | Rule | Access |
|--------|------|--------|
| evcc instance (site creds) | Topic starts with `user/{userId}/site/{siteId}/evcc/` | Read + Write (full subtree) |
| Frontend (user creds) | Topic starts with `user/{userId}/site/` and contains `/evcc/` | Read |
| Frontend (user creds) | Topic ends with `/set` under user's site prefix | Write |

### Database Schema

```sql
-- Users table (Phase 1 — migrated from current)
users (
  id              TEXT PRIMARY KEY,  -- UUID
  email           TEXT UNIQUE NOT NULL,
  password_hash   TEXT NOT NULL,
  mqtt_username   TEXT UNIQUE NOT NULL,  -- frontend MQTT credentials
  mqtt_password   TEXT NOT NULL,         -- frontend MQTT credentials (plaintext)
  created_at      DATETIME NOT NULL,
  updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)

-- Sites table (Phase 1 — new)
sites (
  id            TEXT PRIMARY KEY,  -- UUID
  user_id       TEXT NOT NULL REFERENCES users(id),
  name          TEXT NOT NULL,
  mqtt_username TEXT UNIQUE NOT NULL,  -- evcc instance credentials
  mqtt_password TEXT NOT NULL,         -- evcc instance credentials (plaintext)
  topic_prefix  TEXT UNIQUE NOT NULL,
  timezone      TEXT,
  created_at    DATETIME NOT NULL,
  updated_at    DATETIME NOT NULL
)

-- Refresh tokens table (Phase 2 — new)
refresh_tokens (
  id          TEXT PRIMARY KEY,  -- UUID
  user_id     TEXT NOT NULL REFERENCES users(id),
  token_hash  TEXT NOT NULL,
  expires_at  DATETIME NOT NULL,
  created_at  DATETIME NOT NULL
)
```

### Migration Strategy

SQLite does not support `ALTER TABLE DROP COLUMN` reliably. The migration uses the existing `CREATE TABLE IF NOT EXISTS` pattern with these steps:

1. Create the `sites` table
2. For each existing user: create a default site ("My Home") with a new site-specific MQTT credential set, copying the user's current `topic_prefix` restructured to the new format
3. The user's existing `mqtt_username`/`mqtt_password` remain on the `users` table and become the frontend read-only credential
4. The old `topic_prefix` column on `users` is kept but ignored (SQLite column drops are destructive)
5. Add `updated_at` column to `users`: `ALTER TABLE users ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP` (existing rows get current timestamp; application sets `updated_at = created_at` for backfilled rows where possible)
6. Migration runs automatically on application startup (same pattern as existing `migrate()` function)

---

## Phased Delivery

### Phase 1: Data Model & Real evcc Integration

**Goals:** Implement the multi-site data model, connect real evcc instances, validate with real MQTT payloads.

#### Backend Changes

**JWT auth middleware (new):** All `/api/sites/*` endpoints require a valid JWT in the `Authorization: Bearer <token>` header. This middleware is new work — the current codebase has no protected endpoints.

New API endpoints (all require JWT auth):

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/sites` | Create a site |
| GET | `/api/sites` | List user's sites |
| PUT | `/api/sites/:id` | Update site name/metadata |
| DELETE | `/api/sites/:id` | Remove a site |

No single-site GET endpoint — site count per user is small enough to filter from the list response.

**`POST /api/sites` response:**
```json
{
  "site": {
    "id": "uuid",
    "name": "My Home",
    "topicPrefix": "user/{userId}/site/{siteId}/evcc",
    "mqttUsername": "site-generated-username",
    "mqttPassword": "site-generated-password",
    "timezone": null,
    "createdAt": "2026-03-20T12:00:00Z"
  }
}
```

MQTT credentials are only returned on creation. `GET /api/sites` returns sites without `mqttPassword`.

**Login/register response (updated for multi-site):**
```json
{
  "token": "jwt-access-token",
  "mqttUsername": "user-level-mqtt-username",
  "mqttPassword": "user-level-mqtt-password",
  "userId": "uuid"
}
```

The `topicPrefix` is removed from the auth response. The frontend fetches the user's sites via `GET /api/sites` after login and subscribes to the selected site's topic prefix.

**ACL plugin update:** Rewritten to support two credential types:
- Site credentials: full read/write to `user/{userId}/site/{siteId}/evcc/#`
- User credentials: read across all user's sites, write only to `/set` topics

The current restriction that only allows writes to `/set` topics is removed for site credentials, since real evcc instances publish to arbitrary data topics.

- Auto-migrate existing users to the new schema on startup (see Migration Strategy above)

#### Frontend Changes

- Site list/selector in the UI (dropdown or sidebar)
- "Add site" flow showing the user what to configure in their local evcc
- MQTT client subscribes to `user/{userId}/site/{siteId}/evcc/#` based on selected site
- Re-subscribe when switching sites

#### Real evcc Connection

The user configures their local evcc's MQTT section:

```yaml
mqtt:
  broker: mqtts://cloud-broker:8883
  user: <site_mqtt_username>
  password: <site_mqtt_password>
  topic: user/{userId}/site/{siteId}/evcc
```

No bridge or agent needed — evcc publishes natively.

#### Validation

- Test with at least one real evcc instance
- Document gaps between expected and actual MQTT payloads
- Ensure store mapping handles real evcc topic paths

---

### Phase 2: Security Hardening

**Goals:** Proper auth lifecycle, input validation, transport security, security headers.

#### Token Refresh

- Access token: 15-minute expiry
- Refresh token: 30-day expiry, stored hashed in `refresh_tokens` table
- New endpoint: `POST /api/auth/refresh`
- Refresh token rotation: each use invalidates the old token
- Frontend silently refreshes before access token expires

#### Logout

- `POST /api/auth/logout` — invalidates refresh token server-side
- Frontend clears localStorage and disconnects MQTT

#### Input Validation & Rate Limiting

- Validate all API inputs (email format, password strength, site name length)
- Rate limit auth endpoints (5 login attempts per minute per IP)
- Rate limit site creation (10 sites per user max)

#### Transport Security

- Enforce TLS on all connections — HTTP redirects to HTTPS
- MQTT over TLS only (port 8883 for native, WSS for WebSocket)
- Remove any plaintext fallbacks

#### Secret Management

- JWT secret loaded from environment, documented rotation procedure
- MQTT credentials generated with sufficient entropy
- No secrets in code or git history

#### CORS & Headers

- Restrict CORS to known origins in production
- Security headers: `Strict-Transport-Security`, `X-Content-Type-Options`, `X-Frame-Options`

---

### Phase 3: Reliability

**Goals:** Resilient MQTT client, error boundaries, graceful degradation, backend robustness.

#### MQTT Client Resilience

- Auto-reconnect with exponential backoff + jitter (1s, 2s, 4s... up to 30s)
- Re-subscribe to active site's topics after reconnect
- Connection status indicator (connected / reconnecting / offline)

#### Error Boundaries

- Vue error boundaries around major UI sections — a broken chart doesn't crash the dashboard
- Consistent backend error format: `{error: string, code: string}`
- Gracefully ignore malformed MQTT payloads, log warnings

#### Offline & Degraded Mode

- Stale data indicator: "last updated X ago" when MQTT data hasn't arrived recently
- Cache last known state in localStorage, show on page reload while reconnecting
- If backend is unreachable but tokens are valid, still attempt MQTT connection

#### Backend Resilience

- `/health` verifies DB connectivity and MQTT broker reachability
- Graceful shutdown: drain active connections on SIGTERM
- Request timeouts to prevent resource exhaustion

---

### Phase 4: UI Polish

**Goals:** Polished multi-site experience, onboarding, responsive design, extensibility.

#### Multi-Site Experience

- Site switcher dropdown in the header bar
- Site overview page: summary cards for all sites (status, power, active loadpoints)
- Single-site users skip the overview, go straight to the dashboard

#### Onboarding

- First-time flow: guide through creating a site and configuring local evcc (step-by-step with copyable config snippets)
- Connection verification: "waiting for data..." state that transitions to dashboard on first MQTT message
- Meaningful empty states instead of blank screens

#### Responsive Design

- Mobile-first audit of cloud-specific additions (site switcher, onboarding, site management)
- Touch-friendly control sizing

#### Extensibility Foundation

- Keep Vue components modular and well-bounded
- Identify insertion points for future cloud-only features (dashboard widgets, notification settings) — don't build them, just ensure the structure supports them

#### Visual Consistency

- Stay close to upstream evcc look and feel (color palette, typography, component style)
- Cloud-specific additions should feel native, not bolted on
