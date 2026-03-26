# UI Polish — Design Spec

**Date:** 2026-03-26
**Source:** Manual test findings after testreport-fixes deployment
**Scope:** 6 focused UI improvements, no backend changes

---

## Architecture Note

Navigation (`TopNavigationArea`) and Footer live inside `Site.vue`, which is only rendered after MQTT data arrives (via `router-view → Main.vue → Site.vue`). During initial loading (`waitingForData === true`), `App.vue` renders `<WaitingForData>` instead of `<router-view>`, so no nav or footer exists in the DOM at all.

---

## Issues & Fixes

### 1. Branding: "EVCC Cloud Connect" → "evcc Hub"

**Affected file:** `web/assets/js/views/LoginView.vue` (lines 7, 42)

Replace both occurrences of `"☀ evcc Cloud Connect"` with `"⚡ evcc Hub"`.

---

### 2. Login Page — Subtitle, GitHub Link & Footer

**Affected file:** `web/assets/js/views/LoginView.vue`

**Under the "evcc Hub" title** (in both `login` and `register` modes), add:

```html
<p class="text-center text-muted small mb-4">
  Dein evcc-Dashboard, von überall erreichbar.<br>
  <a href="https://github.com/tjarkste/evcc-hub" target="_blank" rel="noopener" class="text-primary">
    Open Source auf GitHub →
  </a>
</p>
```

**New footer section** at the bottom of `LoginView.vue`. Add it once as a sibling `<div>` directly after the last `v-else-if` mode block (the `onboarding` block ending at line 145), still inside the outer card wrapper `<div class="card p-4">`. Do not add it inside each mode block. `LoginView.vue` currently has no footer — add one:

```html
<div class="text-center mt-4 pt-3 border-top">
  <small class="text-muted">
    <router-link to="/impressum">Impressum</router-link> ·
    <router-link to="/datenschutz">Datenschutz</router-link> ·
    <router-link to="/nutzungsbedingungen">Nutzungsbedingungen</router-link> ·
    <a href="https://github.com/tjarkste/evcc-hub" target="_blank" rel="noopener">
      <svg width="12" height="12" viewBox="0 0 16 16" fill="currentColor" style="vertical-align:middle; margin-right:2px;"><path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/></svg>
      GitHub
    </a>
  </small>
</div>
```

---

### 3. Loading Screen — Persistent Nav & Footer

**Affected files:** `web/assets/js/views/App.vue`, `web/assets/js/components/WaitingForData.vue`

**Problem:** During loading (`waitingForData === true`), `App.vue` renders only `<WaitingForData>`. `TopNavigationArea` and `Footer` live in `Site.vue` which is not in the DOM yet. The user has no access to Settings, MQTT credentials, or legal links.

**Fix — App.vue:** When `waitingForData` is true, render `TopNavigationArea` and a minimal footer directly in `App.vue`, flanking `WaitingForData`:

```html
<!-- In App.vue template, replace the WaitingForData branch: -->
<template v-if="waitingForData">
  <TopNavigationArea :notifications="notifications ?? []" />
  <WaitingForData />
  <div class="text-center py-3">
    <small class="text-muted">
      <router-link to="/impressum">Impressum</router-link> ·
      <router-link to="/datenschutz">Datenschutz</router-link> ·
      <router-link to="/nutzungsbedingungen">Nutzungsbedingungen</router-link>
    </small>
  </div>
</template>
```

**Note on TopNavigationArea:** Its only accepted prop is `notifications: Notification[]` (defaults to `[]`). All other data it reads from `store.state` internally. Do not pass `:site`, `:sites`, or `@select-site` — these are not part of its interface.

**Fix — App.vue `.app` style:** Add `display: flex; flex-direction: column` to the `.app` scoped style so that `flex: 1` on `WaitingForData` works:

```css
.app {
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
}
```

**Fix — WaitingForData.vue:** Remove `min-height: 100vh` / `min-height: 100dvh` from the `.waiting-overlay` style. Replace with `flex: 1` so it fills the space between nav and footer:

```css
.waiting-overlay {
  flex: 1;
  /* remove: min-height: 100vh; min-height: 100dvh; */
}
```

**Result:** The Navigation dropdown (Settings, Profile, Logout, MQTT credentials) and legal links are accessible even while MQTT is connecting.

---

### 4. Loading State — Consistent Initial Message

**Affected file:** `web/assets/js/components/WaitingForData.vue`

**Problem:** Chrome shows "Verbindung wird wiederhergestellt..." (RECONNECTING state) on initial page load; Safari shows "Verbinde mit MQTT-Broker..." (OFFLINE). These describe the same scenario inconsistently.

**Fix:** In the `statusTitle` computed property, map `RECONNECTING` to the same Stage 1 message as `OFFLINE`:

```typescript
// In statusTitle computed:
if (store.state.connectionState === ConnectionState.RECONNECTING) {
  return "Verbinde mit MQTT-Broker...";  // was: "Verbindung wird wiederhergestellt..."
}
```

`WaitingForData` is only shown during initial load (before first data arrives). RECONNECTING in this context means "no data yet", semantically identical to OFFLINE. The "Verbindung wird wiederhergestellt..." message is still used by `ConnectionStatus.vue` during active reconnects after data loss.

**Updated state mapping:**

| State | Stage | Title |
|-------|-------|-------|
| OFFLINE | 1 | "Verbinde mit MQTT-Broker..." |
| RECONNECTING | 1 | "Verbinde mit MQTT-Broker..." *(changed)* |
| CONNECTED, no data, < 30s | 2 | "Verbunden — warte auf Daten von deiner evcc-Instanz..." |
| CONNECTED, no data, ≥ 30s | 3 | Warning icon + "Keine Daten empfangen..." |

---

### 5. Onboarding Screen — MQTT Config Layout Fix

**Affected file:** `web/assets/js/views/LoginView.vue` (lines 109–113)

**Problem:** The `<pre>` tag containing `{{ mqttConfig }}` has no overflow handling. The content overflows and clips on small screens or when credentials are long.

**Fix:** Add `overflow-x: auto` to the existing inline style:

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
- Changing the evcc Hub logo/icon
- Restructuring the full app layout beyond what is described above

---

## Affected Files Summary

| File | Change |
|------|--------|
| `web/assets/js/views/LoginView.vue` | Rename title, add subtitle + GitHub links, add footer section, fix pre overflow |
| `web/assets/js/components/WaitingForData.vue` | Remove fullscreen height, fix RECONNECTING message |
| `web/assets/js/views/App.vue` | Render TopNavigationArea + minimal footer during loading |
