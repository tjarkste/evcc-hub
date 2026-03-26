# UI Polish — Design Spec

**Date:** 2026-03-26
**Source:** Manual test findings after testreport-fixes deployment
**Scope:** 6 focused UI improvements, no backend changes

---

## Issues & Decisions

### 1. Branding: "EVCC Cloud Connect" → "evcc Hub"

**Affected files:** `LoginView.vue` (lines 7, 42)

Replace all occurrences of "☀ evcc Cloud Connect" (the login card title) with "⚡ evcc Hub".

---

### 2. Login Page — Subtitle & GitHub Link

**Affected file:** `LoginView.vue`

**Under the "evcc Hub" title**, add a subtitle and GitHub link in both login and register modes:

```
Dein evcc-Dashboard, von überall erreichbar.
Open Source auf GitHub →
```

- Subtitle text: `"Dein evcc-Dashboard, von überall erreichbar."`
- GitHub link text: `"Open Source auf GitHub →"`
- GitHub URL: `https://github.com/tjarkste/evcc-hub`
- Link opens in new tab (`target="_blank" rel="noopener"`)
- Style: subtitle in `text-muted small`, GitHub link in primary color (`text-primary`)

**In the footer** of the login page (next to Impressum / Datenschutz / Nutzungsbedingungen), add a GitHub link:

- GitHub icon (SVG) + text "GitHub"
- Same URL as above, opens in new tab
- Style consistent with existing footer links

---

### 3. Loading Screen — Navigation always visible

**Affected files:** `WaitingForData.vue`, `App.vue`

**Problem:** `WaitingForData` is a fullscreen overlay (`min-height: 100vh`) that covers the top navigation and footer, making Settings, MQTT credentials, Impressum/Datenschutz etc. inaccessible while MQTT is connecting.

**Fix:** `WaitingForData` fills only the main content area, not the full viewport. The component must NOT use `min-height: 100vh`. Instead it uses `flex: 1` or a fixed height that fills the space between nav and footer.

**Result:** Navigation dropdown (Settings, Profile, Logout), footer links (Impressum, Datenschutz, Nutzungsbedingungen) remain accessible at all times — including during initial MQTT connection.

No changes to which routes are accessible — protected routes still require auth.

---

### 4. Loading State — Consistent Initial Message

**Affected file:** `WaitingForData.vue`

**Problem:** Chrome shows "Verbindung wird wiederhergestellt..." (RECONNECTING state) on initial page load because it has a cached connection attempt. Safari shows "Verbinde mit MQTT-Broker..." (OFFLINE state). These are inconsistent for the same scenario.

**Fix:** In `WaitingForData.vue`, map both `OFFLINE` and `RECONNECTING` states to the same Stage 1 message:

> "Verbinde mit MQTT-Broker..."

The "Verbindung wird wiederhergestellt..." message (RECONNECTING) should only appear when a connection was already established and then dropped during active use — which is handled by `ConnectionStatus.vue`, not `WaitingForData.vue`.

`WaitingForData` is shown exclusively during the initial load (before first data arrives), so RECONNECTING is semantically equivalent to OFFLINE in this context.

**Updated state mapping in `WaitingForData.vue`:**

| State | Stage | Title |
|-------|-------|-------|
| OFFLINE | 1 | "Verbinde mit MQTT-Broker..." |
| RECONNECTING | 1 | "Verbinde mit MQTT-Broker..." *(same as OFFLINE)* |
| CONNECTED, no data, < 30s | 2 | "Verbunden — warte auf Daten von deiner evcc-Instanz..." |
| CONNECTED, no data, ≥ 30s | 3 | Warning icon + "Keine Daten empfangen..." |

---

### 5. Onboarding Screen — MQTT Config Layout Fix

**Affected file:** `LoginView.vue` (line 109–113, the `<pre>` block)

**Problem:** The `<pre>` tag containing the MQTT config (`mqttConfig`) has no overflow handling. The content (broker URL, username, password) is wider than the card on small screens, causing the text to be clipped.

**Fix:** Add `overflow-x: auto` to the `<pre>` style:

```html
<pre
  class="bg-dark text-light p-2 rounded mb-2"
  style="font-size: 0.8em; overflow-x: auto;"
  data-test="mqtt-config"
>{{ mqttConfig }}</pre>
```

---

## Out of Scope

- Backend changes (none required)
- Navigation restructuring beyond making it accessible during loading
- Changing the evcc Hub logo/icon
- Changing footer link order

---

## Affected Files Summary

| File | Change |
|------|--------|
| `web/assets/js/views/LoginView.vue` | Rename title, add subtitle + GitHub links, fix pre overflow |
| `web/assets/js/components/WaitingForData.vue` | Remove fullscreen, unify RECONNECTING message |
| `web/assets/js/views/App.vue` | Adjust layout so nav stays visible behind WaitingForData |
