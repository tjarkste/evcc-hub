# Production Launch Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deploy evcc Cloud to a production Hetzner CX22 VPS with SSL, backups, monitoring, legal pages, and a soft-launch-ready frontend.

**Architecture:** The existing Docker Compose stack (Mosquitto + Go backend + Nginx) is deployed unchanged on a Hetzner CX22. SSL is handled by Let's Encrypt via Certbot standalone (initial issue) + webroot renewal. Legal pages are static Vue views added to the existing router. Backups run via a host-level cron script to Hetzner Object Storage. Sentry is integrated in both the Vue frontend and the Go backend (EU region, PII scrubbing enabled).

**Tech Stack:** Ubuntu 24.04, Docker Compose, Certbot, UFW, Nginx, Vue 3 + Vue Router, @sentry/vue, sentry-go, rclone

---

## File Map

### New files to create

| File | Responsibility |
|---|---|
| `deploy/nginx/nginx.prod.conf` | Production nginx config with Let's Encrypt cert paths + ACME webroot |
| `deploy/scripts/backup.sh` | Daily SQLite backup → Hetzner Object Storage |
| `deploy/scripts/restore-test.sh` | Manual restore verification script |
| `web/assets/js/views/Impressum.vue` | Impressum page (content filled in manually post-generation) |
| `web/assets/js/views/Datenschutz.vue` | Datenschutz page (content filled in manually post-generation) |
| `web/assets/js/views/Nutzungsbedingungen.vue` | Nutzungsbedingungen page (content written manually) |

### Files to modify

| File | Change |
|---|---|
| `docker-compose.yml` | Add certbot webroot renewal service, mount letsencrypt certs into nginx, fix DB volume |
| `web/assets/js/router.ts` | Add 3 legal routes with `noAuth: true` |
| `web/assets/js/components/Footer/Footer.vue` | Add legal links row |
| `web/assets/js/app.ts` | Initialize Sentry (frontend) |
| `web/package.json` | Add @sentry/vue dependency |
| `backend/cmd/server/main.go` | Initialize Sentry (backend) |
| `backend/go.mod` | Add github.com/getsentry/sentry-go |

---

## Phase 1: Pre-requisites (Manual — No Code)

### Task 1: Register Domain and Provision Hetzner Server

These steps are manual. Complete them before any code tasks.

- [ ] **Step 1: Register a domain**

  Go to [inwx.de](https://www.inwx.de) or [Hetzner Domains](https://www.hetzner.com/domainregistration).
  Register your chosen domain (e.g., `evcc-hub.de`).

- [ ] **Step 2: Create Hetzner Cloud account and project**

  Go to [console.hetzner.cloud](https://console.hetzner.cloud).
  Create a new project, e.g., `evcc-cloud`.

- [ ] **Step 3: Create a CX22 server**

  - Image: **Ubuntu 24.04**
  - Location: **Falkenstein** or **Nuremberg**
  - Enable backups: **No** (handled manually — cheaper)
  - Add your SSH public key during setup

  Note down the server's public IPv4 address.

- [ ] **Step 4: Point domain to server IP**

  In your domain registrar's DNS panel, create:
  ```
  A    @    <server-ipv4>    TTL 300
  A    www  <server-ipv4>    TTL 300
  ```
  Verify propagation:
  ```bash
  dig +short yourdomain.de
  ```
  Expected: your server's IP address.

---

### Task 2: Initial Server Setup

SSH into the server as root: `ssh root@<server-ip>`

- [ ] **Step 1: Update system and install Docker**

  ```bash
  apt update && apt upgrade -y
  apt install -y ca-certificates curl gnupg ufw sqlite3

  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
  chmod a+r /etc/apt/keyrings/docker.gpg
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo $VERSION_CODENAME) stable" | tee /etc/apt/sources.list.d/docker.list
  apt update && apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
  ```

  Verify: `docker --version` → prints `Docker version 26.x.x`

- [ ] **Step 2: Configure UFW firewall**

  ```bash
  ufw default deny incoming
  ufw default allow outgoing
  ufw allow 22/tcp comment 'SSH'
  ufw allow 80/tcp comment 'HTTP (redirect to HTTPS)'
  ufw allow 443/tcp comment 'HTTPS'
  ufw --force enable
  ufw status
  ```

  Expected: rules for 22, 80, 443 shown as ALLOW.

- [ ] **Step 3: Create deploy user**

  ```bash
  adduser deploy
  usermod -aG docker deploy
  mkdir -p /home/deploy/.ssh
  cp /root/.ssh/authorized_keys /home/deploy/.ssh/
  chown -R deploy:deploy /home/deploy/.ssh
  chmod 700 /home/deploy/.ssh
  chmod 600 /home/deploy/.ssh/authorized_keys
  ```

  Test: `ssh deploy@<server-ip>` — should log in without a password prompt.

- [ ] **Step 4: Create app directory**

  ```bash
  mkdir -p /opt/evcc-cloud/data
  chown -R deploy:deploy /opt/evcc-cloud
  ```

---

## Phase 2: SSL — Let's Encrypt

### Task 3: Update Nginx Config for Let's Encrypt

- [ ] **Step 1: Create production nginx config**

  Create `deploy/nginx/nginx.prod.conf`.
  Replace `YOUR_DOMAIN_HERE` with your actual domain in both places:

  ```nginx
  events {
    worker_connections 1024;
  }

  http {
    include       mime.types;
    default_type  application/octet-stream;
    sendfile      on;

    # HTTP → HTTPS redirect + ACME challenge for cert renewal
    server {
      listen 80;
      server_name _;

      location /.well-known/acme-challenge/ {
        root /var/www/certbot;
      }

      location / {
        return 301 https://$host$request_uri;
      }
    }

    server {
      listen 443 ssl;
      server_name YOUR_DOMAIN_HERE;

      ssl_certificate     /etc/letsencrypt/live/YOUR_DOMAIN_HERE/fullchain.pem;
      ssl_certificate_key /etc/letsencrypt/live/YOUR_DOMAIN_HERE/privkey.pem;
      ssl_protocols       TLSv1.2 TLSv1.3;
      ssl_session_cache   shared:SSL:10m;
      ssl_session_timeout 10m;

      add_header Strict-Transport-Security "max-age=63072000; includeSubDomains" always;
      add_header X-Content-Type-Options "nosniff" always;
      add_header X-Frame-Options "DENY" always;
      add_header Referrer-Policy "strict-origin-when-cross-origin" always;

      root /usr/share/nginx/html;
      index index.html;

      location / {
        try_files $uri $uri/ /index.html;
      }

      location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
      }

      location /mqtt {
        proxy_pass http://mosquitto:9001;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
      }
    }
  }
  ```

- [ ] **Step 2: Update docker-compose.yml**

  Replace the full contents of `docker-compose.yml`:

  ```yaml
  version: "3.8"
  services:
    mosquitto:
      image: eclipse-mosquitto:2
      ports:
        - "8883:8883"
        - "9001:9001"
      volumes:
        - ./deploy/mosquitto/mosquitto.conf:/mosquitto/config/mosquitto.conf
        - ./deploy/mosquitto/certs:/mosquitto/certs

    backend:
      build: ./backend
      ports:
        - "8080:8080"
      env_file: .env
      environment:
        MQTT_BROKER_ADDR: "mosquitto:1883"
        DB_PATH: "/data/evcc.db"
      volumes:
        - ./data:/data
      depends_on:
        - mosquitto
      restart: unless-stopped

    nginx:
      image: nginx:alpine
      ports:
        - "443:443"
        - "80:80"
      volumes:
        - ./deploy/nginx/nginx.prod.conf:/etc/nginx/nginx.conf
        - ./web/dist:/usr/share/nginx/html
        - /etc/letsencrypt:/etc/letsencrypt:ro
        - /var/www/certbot:/var/www/certbot:ro
      depends_on:
        - backend
        - mosquitto
      restart: unless-stopped

    certbot:
      image: certbot/certbot
      volumes:
        - /etc/letsencrypt:/etc/letsencrypt
        - /var/www/certbot:/var/www/certbot
      entrypoint: "/bin/sh -c 'trap exit TERM; while :; do certbot renew --webroot -w /var/www/certbot --quiet; sleep 12h & wait $${!}; done'"
      restart: unless-stopped
  ```

  Key changes vs. the original:
  - `nginx` now mounts `nginx.prod.conf` and Let's Encrypt cert directories
  - `backend` has explicit `DB_PATH=/data/evcc.db` and a `./data:/data` volume
  - `certbot` handles automatic renewal every 12h

- [ ] **Step 3: Commit**

  ```bash
  git add deploy/nginx/nginx.prod.conf docker-compose.yml
  git commit -m "feat: add Let's Encrypt SSL config and fix DB volume for production"
  ```

---

### Task 4: Issue First Certificate on Server

Run **on the server** after deploying code (during Task 10). The first certificate is issued via `--standalone` (Certbot binds directly to port 80 — no nginx needed yet).

- [ ] **Step 1: Create required directories**

  ```bash
  mkdir -p /var/www/certbot /etc/letsencrypt
  ```

- [ ] **Step 2: Issue certificate via standalone mode**

  ```bash
  docker run --rm \
    -p 80:80 \
    -v /etc/letsencrypt:/etc/letsencrypt \
    certbot/certbot certonly \
    --standalone \
    -d yourdomain.de \
    --email your@email.de \
    --agree-tos \
    --no-eff-email
  ```

  Expected: `Successfully received certificate. Certificate is saved at: /etc/letsencrypt/live/yourdomain.de/fullchain.pem`

- [ ] **Step 3: Start full stack**

  ```bash
  cd /opt/evcc-cloud
  docker compose up -d
  docker compose ps
  ```

  Expected: 4 services running (mosquitto, backend, nginx, certbot).

- [ ] **Step 4: Verify SSL**

  ```bash
  curl -I https://yourdomain.de
  ```
  Expected: `HTTP/2 200`

---

## Phase 3: Backup Script

### Task 5: SQLite Backup to Hetzner Object Storage

- [ ] **Step 1: Create Hetzner Object Storage bucket**

  In [Hetzner Cloud Console](https://console.hetzner.cloud):
  - Go to **Object Storage** → Create bucket, e.g., `evcc-cloud-backups`
  - Region: same as your server (Falkenstein or Nuremberg)
  - Under **Security → S3 Credentials**: create an access key + secret key
  - Note the S3 endpoint URL shown on the bucket page (e.g., `fsn1.your-objectstorage.com`)

- [ ] **Step 2: Install rclone on server**

  ```bash
  curl https://rclone.org/install.sh | sudo bash

  rclone config
  # choose: n (new remote)
  # name: hetzner
  # storage type: s3
  # provider: Other
  # access_key_id: <your key>
  # secret_access_key: <your secret>
  # endpoint: fsn1.your-objectstorage.com  (use your exact endpoint URL)
  # leave everything else as default
  ```

  Test: `rclone ls hetzner:evcc-cloud-backups` → empty list, no error.

- [ ] **Step 3: Create `deploy/scripts/` directory**

  ```bash
  mkdir -p deploy/scripts
  ```

- [ ] **Step 4: Create backup script**

  Create `deploy/scripts/backup.sh`:

  ```bash
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
  ```

- [ ] **Step 5: Create restore test script**

  Create `deploy/scripts/restore-test.sh`:

  ```bash
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
  ```

- [ ] **Step 6: Commit**

  ```bash
  git add deploy/scripts/backup.sh deploy/scripts/restore-test.sh
  git commit -m "feat: add SQLite backup and restore-test scripts"
  ```

- [ ] **Step 7: Install scripts and cron on server**

  After deploying (Task 10):
  ```bash
  chmod +x /opt/evcc-cloud/deploy/scripts/backup.sh

  # 2:00 AM daily
  (crontab -l 2>/dev/null; echo "0 2 * * * /opt/evcc-cloud/deploy/scripts/backup.sh >> /var/log/evcc-backup.log 2>&1") | crontab -
  crontab -l  # verify it appears
  ```

---

## Phase 4: Sentry Integration

### Task 6: Sentry Frontend (Vue)

- [ ] **Step 1: Create Sentry project**

  Go to [sentry.io](https://sentry.io). Sign up. During signup, select **EU region (Frankfurt)**.
  Create a project of type **Vue**.
  Copy the DSN from: Project Settings → Client Keys (DSN).

- [ ] **Step 2: Install @sentry/vue**

  ```bash
  cd web
  npm install @sentry/vue
  ```

- [ ] **Step 3: Add Sentry DSN to Vite env file**

  Vite reads `.env` files from its `root`, which is `web/assets/` (see `vite.config.ts`).

  For local testing, create `web/assets/.env.local` (gitignored):
  ```
  VITE_SENTRY_DSN=https://xxxxx@oXXXXX.ingest.de.sentry.io/YYYYYYY
  ```

  For production builds on the server, create `web/assets/.env` before running `npm run build`:
  ```
  VITE_SENTRY_DSN=https://xxxxx@oXXXXX.ingest.de.sentry.io/YYYYYYY
  ```

  > **Note:** Do NOT put `VITE_*` variables in the project-root `.env` — that file is read by Docker Compose for the backend, not by Vite.

- [ ] **Step 4: Initialize Sentry in app.ts**

  Open `web/assets/js/app.ts`.

  Add import at the top of the file, after the existing imports:
  ```typescript
  import * as Sentry from "@sentry/vue";
  ```

  Add this block **after** `const app = createApp(...)` and **before** `app.use(i18n)`:

  ```typescript
  if (import.meta.env.VITE_SENTRY_DSN) {
    Sentry.init({
      app,
      dsn: import.meta.env.VITE_SENTRY_DSN as string,
      environment: import.meta.env.MODE,
      // Only errors — no performance tracing (free tier)
      tracesSampleRate: 0,
      // PII scrubbing: strip personal data before sending
      beforeSend(event) {
        if (event.user) {
          delete event.user.email;
          delete event.user.username;
          delete event.user.ip_address;
        }
        if (event.request?.url) {
          event.request.url = event.request.url.replace(/\/users\/[^/]+/g, "/users/[id]");
        }
        return event;
      },
    });
  }
  ```

- [ ] **Step 5: Verify no Sentry requests in dev (no DSN set)**

  Run dev server: `npm run dev` (runs on port 7071).
  Open browser DevTools → Network tab.
  Filter by `sentry` — no requests should appear when no DSN is configured.

- [ ] **Step 6: Commit**

  ```bash
  git add web/assets/js/app.ts web/package.json web/package-lock.json
  git commit -m "feat: add Sentry EU frontend integration with PII scrubbing"
  ```

---

### Task 7: Sentry Backend (Go)

- [ ] **Step 1: Add sentry-go dependency**

  ```bash
  cd backend
  go get github.com/getsentry/sentry-go
  go get github.com/getsentry/sentry-go/gin
  ```

- [ ] **Step 2: Initialize Sentry in main.go**

  Open `backend/cmd/server/main.go`.

  Add to the import block (add any that are not already present):
  ```go
  "log"
  "os"
  "time"
  "github.com/getsentry/sentry-go"
  sentrygin "github.com/getsentry/sentry-go/gin"
  ```

  Add Sentry init at the start of `main()`, before the router is created:
  ```go
  if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
      if err := sentry.Init(sentry.ClientOptions{
          Dsn:         dsn,
          Environment: os.Getenv("APP_ENV"),
          // PII scrubbing
          BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
              event.User = sentry.User{}
              return event
          },
      }); err != nil {
          log.Printf("sentry init failed: %v", err)
      }
      defer sentry.Flush(2 * time.Second)
  }
  ```

  Register Sentry middleware on the Gin router (find where `gin.New()` or `gin.Default()` is called and add the middleware):
  ```go
  router.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
  ```

- [ ] **Step 3: Add SENTRY_DSN to docker-compose.yml backend environment**

  Open `docker-compose.yml`, under `backend.environment`, add:
  ```yaml
  SENTRY_DSN: "${SENTRY_DSN}"
  APP_ENV: "production"
  ```

- [ ] **Step 4: Add SENTRY_DSN to .env on server**

  The root `.env` (read by Docker Compose) should include:
  ```
  SENTRY_DSN=https://xxxxx@oXXXXX.ingest.de.sentry.io/YYYYYYY
  ```
  Use the **Go** DSN from Sentry (same project or a separate project → Project Settings → Client Keys → choose platform Go).

- [ ] **Step 5: Verify it compiles**

  ```bash
  cd backend
  go build ./...
  ```

  Expected: no errors.

- [ ] **Step 6: Commit**

  ```bash
  git add backend/cmd/server/main.go backend/go.mod backend/go.sum docker-compose.yml
  git commit -m "feat: add Sentry EU backend integration with PII scrubbing"
  ```

---

## Phase 5: Legal Pages

### Task 8: Add Legal Views and Routes

- [ ] **Step 1: Create Impressum.vue**

  Create `web/assets/js/views/Impressum.vue`:

  ```vue
  <template>
    <div class="container py-4">
      <h1>Impressum</h1>
      <!-- Paste generated HTML from e-recht24.de here (Task 9 Step 1) -->
      <p class="text-muted">Inhalt folgt.</p>
    </div>
  </template>

  <script lang="ts">
  import { defineComponent } from "vue";
  export default defineComponent({ name: "Impressum" });
  </script>
  ```

- [ ] **Step 2: Create Datenschutz.vue**

  Create `web/assets/js/views/Datenschutz.vue`:

  ```vue
  <template>
    <div class="container py-4">
      <h1>Datenschutzerklärung</h1>
      <!-- Paste generated HTML from e-recht24.de here (Task 9 Step 2) -->
      <p class="text-muted">Inhalt folgt.</p>
    </div>
  </template>

  <script lang="ts">
  import { defineComponent } from "vue";
  export default defineComponent({ name: "Datenschutz" });
  </script>
  ```

- [ ] **Step 3: Create Nutzungsbedingungen.vue**

  Create `web/assets/js/views/Nutzungsbedingungen.vue`:

  ```vue
  <template>
    <div class="container py-4">
      <h1>Nutzungsbedingungen</h1>
      <!-- Paste content from Task 9 Step 3 here -->
      <p class="text-muted">Inhalt folgt.</p>
    </div>
  </template>

  <script lang="ts">
  import { defineComponent } from "vue";
  export default defineComponent({ name: "Nutzungsbedingungen" });
  </script>
  ```

- [ ] **Step 4: Add routes to router.ts**

  Open `web/assets/js/router.ts`. Inside the `routes` array, add these three entries **before** the `{ path: "/login" ... }` entry:

  ```typescript
  { path: "/impressum", component: () => import("./views/Impressum.vue"), meta: { noAuth: true } },
  { path: "/datenschutz", component: () => import("./views/Datenschutz.vue"), meta: { noAuth: true } },
  { path: "/nutzungsbedingungen", component: () => import("./views/Nutzungsbedingungen.vue"), meta: { noAuth: true } },
  ```

- [ ] **Step 5: Verify routes work without login**

  Run dev server from `web/` dir: `npm run dev`
  The dev server runs on port **7071**.

  In the browser navigate to:
  - `http://localhost:7071/#/impressum`
  - `http://localhost:7071/#/datenschutz`
  - `http://localhost:7071/#/nutzungsbedingungen`

  All three should render without redirecting to `/login`.

- [ ] **Step 6: Commit**

  ```bash
  git add web/assets/js/views/Impressum.vue web/assets/js/views/Datenschutz.vue web/assets/js/views/Nutzungsbedingungen.vue web/assets/js/router.ts
  git commit -m "feat: add legal page views and public routes"
  ```

---

### Task 9: Add Legal Links to Footer

- [ ] **Step 1: Edit Footer.vue — add legal links row**

  Open `web/assets/js/components/Footer/Footer.vue`.

  The current template has:
  ```html
  <div class="d-flex justify-content-between gap-2">
    <Version v-bind="version" />
    <Savings v-bind="savings" />
  </div>
  ```

  Add a second div **after** that div, still inside `<div class="container py-2">`:

  ```html
  <div class="d-flex justify-content-center gap-3 mt-1">
    <router-link to="/impressum" class="text-muted small text-decoration-none">
      Impressum
    </router-link>
    <router-link to="/datenschutz" class="text-muted small text-decoration-none">
      Datenschutz
    </router-link>
    <router-link to="/nutzungsbedingungen" class="text-muted small text-decoration-none">
      Nutzungsbedingungen
    </router-link>
  </div>
  ```

  Do not remove or modify the existing `<script>` section — only add to the template.

- [ ] **Step 2: Verify in browser**

  Dev server at `http://localhost:7071`. Navigate to the main view.
  Footer should show three links at the bottom. Clicking each navigates without login required.

- [ ] **Step 3: Commit**

  ```bash
  git add web/assets/js/components/Footer/Footer.vue
  git commit -m "feat: add legal page links to app footer"
  ```

---

### Task 10: Fill in Legal Content

Manual content task — no code changes to deploy.

- [ ] **Step 1: Generate Impressum**

  Go to [e-recht24.de/impressum-generator.html](https://www.e-recht24.de/impressum-generator.html).
  Enter: full name, street address, ZIP, city, email.
  Copy the HTML output. Paste it into `Impressum.vue`, replacing `<p class="text-muted">Inhalt folgt.</p>`.

- [ ] **Step 2: Generate Datenschutzerklärung**

  Go to [e-recht24.de/datenschutz-generator.html](https://www.e-recht24.de/datenschutz-generator.html).
  Configure:
  - Hosting: **Hetzner Online GmbH, Gunzenhausen, Germany**
  - Add Sentry (EU, Frankfurt) manually as a third-party tool: purpose = "Fehlerprotokollierung", basis = Art. 6(1)(f)
  - Add UptimeRobot EU manually: purpose = "Verfügbarkeitsüberwachung", basis = Art. 6(1)(f)

  Cross-check the generated text against the data table in the spec before pasting:
  `docs/superpowers/specs/2026-03-23-production-launch-design.md` → "Dokument 2: Datenschutzerklärung"

  Paste into `Datenschutz.vue`.

- [ ] **Step 3: Write Nutzungsbedingungen**

  Paste the following into `Nutzungsbedingungen.vue` (replace placeholders):

  ```html
  <h2>Nutzungsbedingungen</h2>
  <p><strong>Stand:</strong> [Datum einsetzen]</p>

  <h3>1. Geltungsbereich</h3>
  <p>Diese Nutzungsbedingungen gelten für die Nutzung des Dienstes evcc Cloud, betrieben von [Dein Name], [Adresse].</p>

  <h3>2. Kostenloser Dienst ohne Gewähr</h3>
  <p>Der Dienst wird kostenlos und ohne jegliche Garantie bereitgestellt. Es besteht kein Anspruch auf Verfügbarkeit, Fehlerfreiheit oder Datensicherheit.</p>

  <h3>3. Kein Support-Versprechen</h3>
  <p>Es wird kein Support-Level vereinbart. Anfragen werden nach Möglichkeit bearbeitet.</p>

  <h3>4. Änderungen und Einstellung</h3>
  <p>Der Betreiber behält sich vor, den Dienst jederzeit ohne Ankündigung zu ändern, einzuschränken oder einzustellen.</p>

  <h3>5. Datenlöschung</h3>
  <p>Nutzer können die Löschung ihrer Daten jederzeit per E-Mail an [deine@email.de] beantragen. Die Löschung wird innerhalb von 30 Tagen bestätigt.</p>
  ```

- [ ] **Step 4: Check Sentry JS SDK for cookies**

  Open browser DevTools → Application → Cookies.
  Load the app with Sentry enabled (set `VITE_SENTRY_DSN` in `web/assets/.env.local` and run dev).
  Verify no cookies named `sentry*` or `__cfduid*` appear.

  If Sentry sets cookies → add a one-line notice in the footer near the legal links:
  ```html
  <span class="text-muted small">Diese App nutzt Sentry zur Fehleranalyse (keine Tracking-Cookies).</span>
  ```
  If no cookies are set → no action needed.

- [ ] **Step 5: Commit final legal content**

  ```bash
  git add web/assets/js/views/Impressum.vue web/assets/js/views/Datenschutz.vue web/assets/js/views/Nutzungsbedingungen.vue
  git commit -m "docs: add legal content (Impressum, Datenschutz, Nutzungsbedingungen)"
  ```

---

## Phase 6: First Deployment

### Task 11: Deploy to Production Server

- [ ] **Step 1: Build frontend for production**

  ```bash
  cd web
  npm run build
  ```

  Expected: `web/dist/` created with compiled assets.

- [ ] **Step 2: Push code to GitHub**

  ```bash
  git push origin main
  ```

- [ ] **Step 3: Clone repo on server**

  On the server as `deploy` user:
  ```bash
  cd /opt/evcc-cloud
  git clone https://github.com/YOUR_USERNAME/YOUR_REPO.git .
  ```

  If already cloned: `git pull origin main`

- [ ] **Step 4: Create Sentry env file for Vite on server**

  ```bash
  nano /opt/evcc-cloud/web/assets/.env
  ```
  Add:
  ```
  VITE_SENTRY_DSN=https://xxxxx@oXXXXX.ingest.de.sentry.io/YYYYYYY
  ```

- [ ] **Step 5: Build frontend on server**

  ```bash
  cd /opt/evcc-cloud/web
  npm ci
  npm run build
  ```

- [ ] **Step 6: Create root .env on server**

  ```bash
  nano /opt/evcc-cloud/.env
  ```
  Add (generate JWT_SECRET with `openssl rand -hex 32`):
  ```
  JWT_SECRET=<64-char-hex>
  SENTRY_DSN=https://xxxxx@oXXXXX.ingest.de.sentry.io/YYYYYYY
  APP_ENV=production
  ```
  Note: `DB_PATH` is set directly in `docker-compose.yml` — no need to add it here.

- [ ] **Step 7: Issue Let's Encrypt cert (Task 4)**

  Follow Task 4 steps now: create dirs, run certbot standalone, then start full stack.

- [ ] **Step 8: Smoke test**

  ```bash
  # From your laptop:
  curl -I https://yourdomain.de
  ```
  Expected: `HTTP/2 200`

  Open browser → `https://yourdomain.de`:
  - App loads
  - Padlock visible
  - HTTP redirects to HTTPS
  - Footer shows three legal links
  - Each legal link is accessible without login

---

## Phase 7: Monitoring Setup

### Task 12: Configure UptimeRobot EU

Manual setup — no code.

- [ ] **Step 1: Create account**

  Go to [eu.uptimerobot.com](https://eu.uptimerobot.com). Sign up.

- [ ] **Step 2: Create HTTPS monitor**

  - Type: **HTTPS**
  - URL: `https://yourdomain.de`
  - Check interval: **5 minutes**
  - Alert contact: your email

- [ ] **Step 3: Test the alert**

  Temporarily stop nginx: `docker compose stop nginx`
  Wait 5–10 minutes. UptimeRobot should send a down-alert email.
  Restart: `docker compose start nginx`

---

## Phase 8: Restore Test

### Task 13: Perform and Document Restore Test

- [ ] **Step 1: Run manual backup**

  ```bash
  /opt/evcc-cloud/deploy/scripts/backup.sh
  ```
  Expected output: `Backup completed: 2026-XX-XX`

- [ ] **Step 2: Run restore test**

  ```bash
  /opt/evcc-cloud/deploy/scripts/restore-test.sh
  ```
  Expected: final line from `sqlite3` command is `ok`

- [ ] **Step 3: Document result (server-local)**

  ```bash
  cat >> /opt/evcc-cloud/RESTORE-TEST.log << EOF
  $(date +%Y-%m-%d): PASS — integrity_check ok
  EOF
  ```

---

## Phase 9: Soft Launch Checklist

### Task 14: Pre-Launch Verification

- [ ] `https://yourdomain.de` loads without cert warning
- [ ] HTTP → HTTPS redirect works
- [ ] Footer shows Impressum / Datenschutz / Nutzungsbedingungen
- [ ] All three legal pages load without requiring login
- [ ] Impressum contains real name and address
- [ ] Datenschutzerklärung mentions Sentry EU + UptimeRobot EU
- [ ] UptimeRobot EU monitor is green
- [ ] Backup cron is installed: `crontab -l | grep backup`
- [ ] Restore test passed
- [ ] Sentry (EU) shows no unexpected errors after first load
- [ ] Docker Compose all 4 services healthy: `docker compose ps`

**You are live. 🚀**

No announcement for Soft Launch. Monitor for 4 weeks. Then post in the evcc community forum — see spec for content guidance: `docs/superpowers/specs/2026-03-23-production-launch-design.md`
