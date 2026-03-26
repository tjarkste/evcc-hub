# UI Issues & i18n Migration Design

**Date**: 2026-03-26
**Status**: Draft

## Problem Statement

Four UI issues were identified during testing of evcc-hub.de:

1. **Duplicate loading indicator**: Both a central spinner ("Verbinde mit MQTT-Broker...") and a bottom-left spinner ("Reconnecting...") show simultaneously during initial load
2. **Settings page unreachable**: Navigating to "Profil & Einstellungen" does nothing because `waitingForData` blocks the `<router-view>` — the settings page doesn't need MQTT data
3. **Irrelevant help/links**: "Need Help?" modal and "evcc.io" link reference the original evcc project, not the evcc-hub
4. **Hardcoded strings**: Many user-facing strings are hardcoded in German or English instead of using i18n translation keys

## Design

### 1. Remove Spinner from ConnectionStatus

**File**: `web/assets/js/components/ConnectionStatus.vue`

Remove the `<span v-if="isReconnecting" class="spinner-border spinner-border-sm">` element. The text status ("Reconnecting...", "Offline", stale data info) remains visible as a non-intrusive status indicator. The central spinner in `WaitingForData.vue` is sufficient as visual loading feedback.

### 2. Route-Level `noDataRequired` Meta Property

**Files**: `web/assets/js/router.ts`, `web/assets/js/views/App.vue`

Instead of adding per-path `if` statements in `App.vue`, introduce a `noDataRequired` route meta property:

**router.ts** — Mark routes that don't need MQTT data:
```ts
{ path: '/settings', component: SettingsView, meta: { noDataRequired: true } }
{ path: '/overview', component: SiteOverview, meta: { noDataRequired: true } }
```

**App.vue** — Simplify `waitingForData` computed:
```js
waitingForData(): boolean {
    if (this.hasCachedState) return false;
    if (this.$route.meta["noAuth"]) return false;
    if (this.$route.meta["noDataRequired"]) return false;
    return store.state.lastDataAt === null;
},
```

Remove the existing `if (this.$route.path === '/overview') return false;` line — it's now covered by the meta property.

### 3. Adapt Help Modal and Remove evcc.io Link

**File**: `web/assets/js/components/Top/Navigation.vue`

- Remove the "evcc.io" external link entirely (lines 123-131)

**File**: `web/assets/js/components/HelpModal.vue`

Reduce the modal content to three items:
- **Keep**: Docs link (docs.evcc.io) — relevant for all evcc users
- **Keep**: GitHub Discussions link — community support
- **Add**: Link to evcc-hub GitHub repo (https://github.com/evcc-hub or actual repo URL)
- **Remove**: "Logs anzeigen" button — doesn't work via the hub
- **Remove**: "Issue erstellen" button — refers to local instance
- **Remove**: "Neustart" button — hub users can't restart the local instance

### 4. Full i18n Migration

**Translation files**: `web/i18n/en.json`, `web/i18n/de.json`

All hardcoded user-facing strings will be replaced with `$t()` calls using a `hub.*` namespace to clearly separate hub-specific keys from original evcc keys (which are loaded from the API).

**Namespace structure**:
- `hub.connection.*` — ConnectionStatus, WaitingForData
- `hub.auth.*` — LoginView (login, register, etc.)
- `hub.settings.*` — SettingsView (profile, password, etc.)
- `hub.sites.*` — SiteManager, SiteOverview
- `hub.nav.*` — Navigation-specific strings
- `hub.error.*` — ErrorBoundary
- `hub.debug.*` — Optimize, Energy (dev pages)
- `hub.issue.*` — Issue page
- `hub.legal.*` — Loading text for Impressum/Datenschutz/Nutzungsbedingungen

**Languages**: DE and EN only. All other 24 supported languages fall back to EN automatically (Vue i18n fallback behavior).

**Files to modify** (~18 Vue components):

| File | Hardcoded Strings |
|---|---|
| `components/ConnectionStatus.vue` | "Reconnecting...", "Offline", "last update Xs/Xm ago" |
| `components/WaitingForData.vue` | "Verbinde mit MQTT-Broker...", "Verbunden — warte auf Daten...", "Keine Daten empfangen.", "Laden..." |
| `components/Top/Navigation.vue` | "Profil & Einstellungen", "Optimize 🧪" |
| `components/ErrorBoundary.vue` | "failed to render.", "Retry", "Unknown error" |
| `components/HelpModal.vue` | Content restructuring (see section 3) |
| `components/SiteCredentialsModal.vue` | "Lädt…" |
| `components/Energyflow/Visualization.vue` | "In", "Out" |
| `components/Optimize/BatteryConfigurationTable.vue` | "Charge / Discharge", "Charge", "Discharge", "None" |
| `components/Config/AuthSuccessBanner.vue` | "Unknown" |
| `views/App.vue` | "Laden..." |
| `views/LoginView.vue` | "Anmelden", "Registrieren", ~10 more strings |
| `views/SettingsView.vue` | "Profil & Einstellungen", "Konto", "Passwort ändern", ~5 more |
| `views/SiteManager.vue` | "Meine Standorte", "Auswählen", "Aktiv", ~6 more |
| `views/SiteOverview.vue` | "Meine Standorte" |
| `views/Optimize.vue` | ~12 debug page strings |
| `views/Energy.vue` | ~3 WIP page strings |
| `views/Issue.vue` | ~8 strings |
| `views/Datenschutz.vue`, `Impressum.vue`, `Nutzungsbedingungen.vue` | "Lädt..." each |

## Out of Scope

- Translating legal pages content (Impressum, Datenschutz, Nutzungsbedingungen) — these are loaded dynamically
- Adding translations beyond DE and EN
- Translating console.log messages or technical strings
