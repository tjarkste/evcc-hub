# User Acquisition Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Execute the user acquisition strategy from the approved spec — pre-launch checks, signup notification, demo artifact, three channel posts, and ongoing cadence setup — to reach 50–100 active users in 3 months.

**Architecture:** This plan is a mix of one small backend code change (signup log + daily check query) and several content/process tasks. Content drafts are version-controlled in `docs/acquisition/`. The backend change adds a single structured log line to `auth_handler.go` — no new dependencies.

**Tech Stack:** Go (Gin backend), PostgreSQL, psql CLI, QuickTime or OBS (screen recording), ffmpeg (GIF export), Markdown for post drafts.

---

## File Map

| Action | Path | Purpose |
|--------|------|---------|
| Modify | `backend/internal/api/auth_handler.go` | Add structured log on new signup |
| Create | `scripts/check-signups.sql` | Manual daily signup check query |
| Create | `docs/acquisition/forum-post-de.md` | German forum post draft (Day 1) |
| Create | `docs/acquisition/github-discussions-post.md` | GitHub Discussions post draft (Day 3) |
| Create | `docs/acquisition/reddit-post-en.md` | Reddit post draft (Day 5) |
| Create | `docs/acquisition/signup-email-template.md` | Personal follow-up email template |
| Create | `docs/acquisition/forum-update-template.md` | Bi-weekly forum update template |

---

## Task 1: Pre-launch verification

Confirm all blockers from the spec are resolved before recording the demo or writing any posts.

**Files:** none (manual checks only)

- [ ] **Step 1: Verify evcc-hub.de is live and reachable**

  Open `https://evcc-hub.de` in a browser. You should see the login/register page.

  Expected: The page loads without SSL errors or DNS failures.
  If it fails: The domain or server is not yet configured — do not proceed with the rest of this plan until resolved.

- [ ] **Step 2: Verify MQTT setup documentation is accessible on the website**

  After logging in, confirm a user can find:
  - Their MQTT credentials (broker URL, topic, username, password)
  - A clear explanation of how to paste them into `evcc.yaml`

  Expected: A new user could complete MQTT setup without emailing you for help.
  If missing: Write the MQTT setup guide in the product before recording the demo.

- [ ] **Step 3: Verify user and site data is queryable in the database**

  Run this query against the production database:

  ```sql
  SELECT u.email, u.created_at, COUNT(s.id) AS site_count
  FROM users u
  LEFT JOIN sites s ON s.user_id = u.id
  GROUP BY u.id, u.email, u.created_at
  ORDER BY u.created_at DESC
  LIMIT 20;
  ```

  Run via: `psql $DATABASE_URL -c "<query above>"`

  Expected: Query executes without error (zero rows is fine — the schema is correct).
  This confirms activity tracking is possible from existing data.

---

## Task 2: Add signup notification to backend

Add a structured log line when a new user registers. This enables daily log scanning to identify new signups without adding SMTP dependencies.

**Files:**
- Modify: `backend/internal/api/auth_handler.go` (line ~38, after `CreateUser` succeeds)
- Create: `scripts/check-signups.sql`

- [ ] **Step 1: Write the failing test**

  In `backend/internal/api/auth_handler_test.go`, add this test to the existing register test group:

  ```go
  func TestRegister_LogsNewSignup(t *testing.T) {
      // This is a log-output test — we capture stderr/stdout
      // to verify the [NEWSIGNUP] marker is emitted.
      var buf bytes.Buffer
      log.SetOutput(&buf)
      defer log.SetOutput(os.Stderr)

      router, db := setupTestRouter(t)
      defer db.TruncateAll()

      body := `{"email":"signup-log@example.com","password":"password123"}`
      w := httptest.NewRecorder()
      req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
      req.Header.Set("Content-Type", "application/json")
      router.ServeHTTP(w, req)

      assert.Equal(t, http.StatusCreated, w.Code)
      assert.Contains(t, buf.String(), "[NEWSIGNUP]")
  }
  ```

  Make sure `bytes` and `os` are imported at the top of the test file.

- [ ] **Step 2: Run the test to confirm it fails**

  ```bash
  cd backend && go test ./internal/api/... -run TestRegister_LogsNewSignup -v
  ```

  Expected output: `FAIL` — `[NEWSIGNUP]` is not yet in the log output.

- [ ] **Step 3: Add the log line to auth_handler.go**

  In `backend/internal/api/auth_handler.go`, after the `CreateUser` call succeeds (around line 30), add one log line:

  ```go
  user, err := h.db.CreateUser(req.Email, req.Password)
  if err != nil {
      // ... existing error handling ...
  }

  log.Printf("[NEWSIGNUP] new user registered (id=%s)", user.ID) // email omitted for PII
  ```

  Do NOT log the email address — user IDs are sufficient for counting and cross-referencing.

- [ ] **Step 4: Run the test to confirm it passes**

  ```bash
  cd backend && go test ./internal/api/... -run TestRegister_LogsNewSignup -v
  ```

  Expected output: `PASS`

- [ ] **Step 5: Run the full backend test suite to confirm no regressions**

  ```bash
  cd backend && go test ./... -short
  ```

  Expected: all tests pass.

- [ ] **Step 6: Create the daily manual check query**

  Create `scripts/check-signups.sql`:

  ```sql
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
  ```

- [ ] **Step 7: Commit**

  ```bash
  git add backend/internal/api/auth_handler.go backend/internal/api/auth_handler_test.go scripts/check-signups.sql
  git commit -m "feat: log [NEWSIGNUP] on registration, add daily check query"
  ```

---

## Task 3: Produce the demo artifact

Record a ~60-second screen capture showing end-to-end remote access. Export as MP4 and GIF.

**Files:** `docs/acquisition/` (store the GIF for forum embedding — MP4 for Reddit)

- [ ] **Step 1: Prepare the recording environment**

  Before hitting record:
  - evcc is running locally and the dashboard is visible in a browser tab
  - A second tab has `evcc.yaml` open (use a text editor, not terminal — easier to read)
  - Browser is at normal desktop width (you will resize later)
  - Notifications and Spotlight are disabled (macOS: System Settings → Notifications → Do Not Disturb on)
  - Browser bookmarks bar is hidden

- [ ] **Step 2: Record the four-scene sequence**

  Use QuickTime (File → New Screen Recording) or OBS. Record at full resolution.

  Scene 1 (~5s): evcc running locally — dashboard is visible with real data (charging power, battery, etc.)
  Scene 2 (~15s): Open `evcc.yaml`. Show the existing `mqtt:` block (or add it). The four lines:
  ```yaml
  mqtt:
    broker: tls://evcc-hub.de:8883
    topic: <your-topic>
    user: <your-username>
    password: "<your-password>"
  ```
  Scene 3 (~5s): Run `sudo systemctl restart evcc` in a terminal — wait for the green "started" output.
  Scene 4 (~20s): Switch to a browser window resized to ~375px wide (simulating mobile). Navigate to `evcc-hub.de`, log in if needed, watch the dashboard load with live data.

  Total target: 45–60 seconds.

- [ ] **Step 3: Export as MP4**

  Save the raw recording as `demo-raw.mov`. Then export a clean MP4:

  ```bash
  ffmpeg -i demo-raw.mov -vcodec h264 -acodec aac -crf 23 docs/acquisition/demo.mp4
  ```

  Check the file size — target under 50 MB for Reddit uploads.

- [ ] **Step 4: Export as GIF**

  Generate a palette first for better quality, then render the GIF:

  ```bash
  ffmpeg -i demo-raw.mov -vf "fps=10,scale=800:-1:flags=lanczos,palettegen" /tmp/palette.png
  ffmpeg -i demo-raw.mov -i /tmp/palette.png -vf "fps=10,scale=800:-1:flags=lanczos,paletteuse" docs/acquisition/demo.gif
  ```

  Target: under 15 MB (forum/GitHub inline embed limit). If too large, reduce fps to 8 or scale to 640:

  ```bash
  ffmpeg -i demo-raw.mov -vf "fps=8,scale=640:-1:flags=lanczos,palettegen" /tmp/palette.png
  ffmpeg -i demo-raw.mov -i /tmp/palette.png -vf "fps=8,scale=640:-1:flags=lanczos,paletteuse" docs/acquisition/demo.gif
  ```

- [ ] **Step 5: Review the output**

  Watch both files. Check:
  - The dashboard data is real (not a loading spinner)
  - The MQTT credentials are blurred or replaced with `<your-username>` etc. — do not publish real credentials
  - The mobile view shows the dashboard fully loaded

- [ ] **Step 6: Commit the artifacts**

  ```bash
  git add docs/acquisition/demo.mp4 docs/acquisition/demo.gif
  git commit -m "feat: add launch demo artifact (MP4 + GIF)"
  ```

---

## Task 4: Write and publish the forum post (Day 1)

Draft, review, commit, then publish to community.evcc.io.

**Files:** Create `docs/acquisition/forum-post-de.md`

- [ ] **Step 1: Write the German forum post draft**

  Create `docs/acquisition/forum-post-de.md` with the following content. Edit the bracketed placeholders before publishing:

  ```markdown
  # evcc von unterwegs? Ich hab's so gelöst.

  Ich wollte mein evcc-Dashboard auch von unterwegs im Blick haben — ohne VPN,
  ohne Portfreigabe, ohne Frickelei. Also hab ich **evcc hub** gebaut.

  [DEMO-GIF HIER EINBETTEN]

  Das ist kein offizielles evcc-Projekt, sondern ein Community-Ding von mir.
  Was es kann:

  - **Remote-Zugriff** — Dashboard von überall, direkt im Browser
  - **Multi-Site** — mehrere Standorte (z.B. Zuhause + Ferienhaus) in einem Account
  - **Kein VPN, kein Portforwarding** — dein evcc baut die Verbindung selbst auf
  - **Kostenlos** — solange ich's betreibe (Beta)

  Einrichtung dauert ~5 Minuten: Account anlegen, 4 Zeilen in die `evcc.yaml`, evcc neustarten.
  Anleitung gibt's direkt nach der Registrierung.

  → **[evcc-hub.de](https://evcc-hub.de)**

  Ich bin Solo-Dev und such Beta-Tester. Feedback — ob positiv oder "das ist kaputt" — ist willkommen.
  ```

- [ ] **Step 2: Review the draft against spec criteria**

  Check each point:
  - [ ] Leads with the pain/outcome, not "I built this"
  - [ ] GIF is embedded inline (not just linked)
  - [ ] No marketing language or superlatives
  - [ ] Explicitly says "Beta" and "kostenlos solange ich's betreibe" — no forever-free promise
  - [ ] Clear single call to action
  - [ ] Under 200 words

- [ ] **Step 3: Commit the draft**

  ```bash
  git add docs/acquisition/forum-post-de.md
  git commit -m "docs: add German forum post draft for Day 1 launch"
  ```

- [ ] **Step 4: Publish to community.evcc.io**

  - Log in to community.evcc.io
  - Create a new topic in the most appropriate category (likely "Allgemein" or "Projekte")
  - Copy the post content from the draft
  - Upload or embed `docs/acquisition/demo.gif` inline
  - Preview the post — confirm the GIF renders
  - Publish

- [ ] **Step 5: Save the published post URL**

  After publishing, note the URL. You will need it for the GitHub Discussions post.

  Add the URL to the top of `docs/acquisition/forum-post-de.md`:
  ```markdown
  **Published:** https://community.evcc.io/t/[your-post-url]
  ```

  Commit:
  ```bash
  git add docs/acquisition/forum-post-de.md
  git commit -m "docs: add published forum post URL"
  ```

---

## Task 5: Write and publish the GitHub Discussions post (Day 3)

Two days after the forum post. Check the active language in evcc's GitHub Discussions first.

**Files:** Create `docs/acquisition/github-discussions-post.md`

- [ ] **Step 1: Confirm the GitHub repo is public and the license is MIT**

  The post claims "Open source (MIT)" — verify this before publishing:
  - Go to `https://github.com/tjarksteenblock/evcc_hub` and confirm the repo is public
  - Confirm a `LICENSE` file exists at the root with MIT text

  If the repo is private, either make it public now or remove the "open source / self-hostable" claims from both the GitHub and Reddit posts.

- [ ] **Step 3: Check active language in evcc GitHub Discussions**

  Visit `https://github.com/evcc-io/evcc/discussions`. Scan the 10 most recent posts:
  - If predominantly German: write the post in German
  - If predominantly English or mixed: write in English

  Note your finding — it affects the draft below.

- [ ] **Step 4: Write the GitHub Discussions post draft**

  Create `docs/acquisition/github-discussions-post.md`:

  ```markdown
  **Published:** [add URL after posting]

  ---

  ## Remote access to evcc without VPN — I built evcc hub

  Hey everyone — sharing a side project that might be useful here.

  I wanted to check my evcc dashboard remotely without setting up VPN or port forwarding,
  so I built **evcc hub**: a free cloud dashboard that syncs with your local evcc instance via MQTT.

  **[Link to demo video: docs/acquisition/demo.mp4 or YouTube upload]**

  What it does:
  - Remote access to your evcc dashboard from anywhere
  - Multi-site support (e.g. home + vacation house)
  - No VPN, no port forwarding — evcc connects outbound
  - Open source, self-hostable
  - Free in beta

  Setup takes ~5 minutes. Full details and discussion in the evcc community forum:
  **[FORUM POST URL]**

  Happy to answer questions here or there. Looking for beta testers — feedback welcome.

  → evcc-hub.de
  ```

  Replace `[FORUM POST URL]` with the URL saved in Task 4 Step 5.
  Translate to German if that was the finding in Step 1.

- [ ] **Step 3: Review the draft**

  Check:
  - [ ] Links to the forum thread (drives discussion to one place)
  - [ ] Mentions open source and self-hostable (GitHub audience cares about this)
  - [ ] Demo is linked or embedded
  - [ ] "I built this" / disclosure is present
  - [ ] No duplicate CTAs — one link to the site, one to the forum

- [ ] **Step 5: Commit the draft**

  ```bash
  git add docs/acquisition/github-discussions-post.md
  git commit -m "docs: add GitHub Discussions post draft for Day 3"
  ```

- [ ] **Step 6: Publish to evcc GitHub Discussions**

  - Go to `https://github.com/evcc-io/evcc/discussions`
  - Click "New discussion"
  - Choose appropriate category
  - Paste the post content
  - Attach or embed the demo (GitHub supports MP4 uploads directly)
  - Publish

- [ ] **Step 7: Save the published URL**

  Update `docs/acquisition/github-discussions-post.md` with the published URL, then commit:

  ```bash
  git add docs/acquisition/github-discussions-post.md
  git commit -m "docs: add published GitHub Discussions URL"
  ```

---

## Task 6: Write and publish the Reddit posts (Day 5)

Two days after GitHub Discussions. Check subreddit rules before writing.

**Files:** Create `docs/acquisition/reddit-post-en.md`

- [ ] **Step 1: Check r/selfhosted rules**

  Visit `https://www.reddit.com/r/selfhosted/about/rules`. Look for:
  - Rules on self-promotion (is it allowed? limited to once per X days?)
  - Required post flair (e.g. "Show and Tell", "Project")
  - Any specific disclosure requirements

  Note your findings.

- [ ] **Step 2: Check r/homeautomation rules**

  Visit `https://www.reddit.com/r/homeautomation/about/rules`. Same checks.

  If either subreddit prohibits self-promotion entirely, skip it and add `r/evcharging` or `r/electricvehicles` as an alternative.

- [ ] **Step 3: Write the Reddit post draft**

  Create `docs/acquisition/reddit-post-en.md`:

  ```markdown
  **Published:**
  - r/selfhosted: [add URL after posting]
  - r/homeautomation: [add URL after posting]

  ---

  **Title:** I built a free cloud dashboard for evcc (EV charger controller) — no VPN, no port forwarding needed

  **Body:**

  Hey r/selfhosted — I built this and wanted to share.

  [DEMO VIDEO/GIF]

  **The problem:** evcc is a great open-source EV charging controller, but accessing it
  remotely meant setting up VPN or opening ports — annoying for a home server setup.

  **What I built:** evcc hub — a free cloud dashboard that syncs with your local evcc
  instance via MQTT over TLS. Your evcc connects outbound, so no inbound ports needed.

  **Features:**
  - Remote access to your evcc dashboard from anywhere
  - Multi-site support (multiple locations in one account)
  - Real-time data via MQTT with TLS encryption
  - Open source and self-hostable (MIT)
  - Free in beta

  **Setup:** Register → get 4 lines of MQTT config → paste into evcc.yaml → restart evcc. Done.

  → **evcc-hub.de** | GitHub: **https://github.com/tjarksteenblock/evcc_hub**

  I'm the developer — happy to answer questions. Looking for beta testers and feedback.
  ```

  Adjust flair and any required formatting based on the rules found in Steps 1–2.

- [ ] **Step 4: Review the draft**

  Check:
  - [ ] Title is descriptive, not clickbait
  - [ ] "I built this" disclosure is present in the body
  - [ ] Self-promotion rules from Steps 1–2 are satisfied
  - [ ] Demo is embedded or linked
  - [ ] GitHub link is present (selfhosted audience always wants the source)
  - [ ] Post stays factual — no "revolutionary", "game-changing" language

- [ ] **Step 5: Commit the draft**

  ```bash
  git add docs/acquisition/reddit-post-en.md
  git commit -m "docs: add Reddit post draft for Day 5"
  ```

- [ ] **Step 6: Post to r/selfhosted**

  - Log in to Reddit
  - Go to r/selfhosted and create a new post
  - Apply the correct flair
  - Paste the content, upload or link the demo
  - Submit

- [ ] **Step 7: Post to r/homeautomation**

  Same process. Do not post to both subreddits simultaneously — wait at least 30 minutes between posts.

- [ ] **Step 8: Save the published URLs**

  Update `docs/acquisition/reddit-post-en.md` with both URLs, then commit:

  ```bash
  git add docs/acquisition/reddit-post-en.md
  git commit -m "docs: add published Reddit post URLs"
  ```

---

## Task 7: Set up the ongoing cadence

Prepare templates and reminders for the personal email follow-up and bi-weekly forum updates.

**Files:**
- Create: `docs/acquisition/signup-email-template.md`
- Create: `docs/acquisition/forum-update-template.md`

- [ ] **Step 1: Write the signup email template**

  Create `docs/acquisition/signup-email-template.md`:

  ```markdown
  # Personal Signup Follow-up Email

  **When to send:** Within 24 hours of a new user signup (check logs daily with `scripts/check-signups.sql`).
  **From:** your personal email address
  **To:** the new user's email (look up from DB by user ID from the [NEWSIGNUP] log)
  **Subject:** evcc hub — alles geklappt?

  ---

  Hey,

  danke fürs Ausprobieren von evcc hub! Hat alles mit der Einrichtung geklappt,
  oder ist irgendwo ein Haken?

  Viele Grüße,
  [Dein Name]

  ---

  **Notes:**
  - Keep it personal and short — one or two sentences max
  - Do not use a template tool or bulk-send — write each email manually
  - No sales language
  - If they don't reply, that's fine — no follow-up
  ```

- [ ] **Step 2: Write the forum update template**

  Create `docs/acquisition/forum-update-template.md`:

  ```markdown
  # Bi-weekly Forum Update Template

  **When to post:** Every 2–3 weeks as a reply to the original forum thread (Task 4).
  **Frequency:** At minimum every 3 weeks while user count is below 50.

  ---

  **Template:**

  Kurzes Update nach [X] Wochen:

  - **Nutzer:** [X] Accounts, [X] aktive Standorte
  - **Neu seit letztem Update:** [list 2–3 concrete improvements or bug fixes]
  - **Bekannte Probleme:** [be honest — list anything still broken]

  Feedback und Bugreports weiterhin willkommen.

  ---

  **Tips:**
  - Fetch the user/site count before writing: `psql $DATABASE_URL -f scripts/check-signups.sql`
  - Be honest about problems — this community respects transparency over spin
  - Keep it short: 5–8 lines max
  - Do not post if there is nothing new to report — skip that cycle
  ```

- [ ] **Step 3: Set up calendar reminders**

  Create two recurring calendar events:

  1. **Daily signup check** — Recurring every day at 09:00
     - Title: "evcc hub — neue Signups prüfen"
     - Note: `psql $DATABASE_URL -f scripts/check-signups.sql` — send personal email if any new users

  2. **Forum update** — Recurring every 2 weeks starting from the forum post date
     - Title: "evcc hub — Forum-Update schreiben"
     - Note: Post a reply to the original forum thread using `docs/acquisition/forum-update-template.md`

- [ ] **Step 4: Commit the templates**

  ```bash
  git add docs/acquisition/signup-email-template.md docs/acquisition/forum-update-template.md
  git commit -m "docs: add signup email and forum update templates for ongoing cadence"
  ```

---

## Success Milestones

Track these after launch using `scripts/check-signups.sql` and the user/site count query from Task 1 Step 3.

| Checkpoint | Target | Action if not met |
|------------|--------|-------------------|
| End of Week 1 | 10 signups | Reply to every forum comment; check if GIF embedded correctly |
| End of Month 1 | 25 active users | Consider a second forum post with a concrete update |
| End of Month 3 | 50–100 active users | Add "Buy me a coffee" link to the site footer |

**Definition of active:** At least one site connected + dashboard viewed in the last 30 days (check `sites` table — `updated_at` is a proxy for connection activity).
