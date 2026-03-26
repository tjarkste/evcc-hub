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

Instead of adding per-path `if` statements in `App.vue`, introduce a `noDataRequired` route meta property.

**Important**: `noDataRequired` does NOT imply `noAuth` — authentication is still enforced by the existing `beforeEach` guard in `router.ts`. These are independent concerns.

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

**Note on SiteOverview rendering**: `SiteOverview` is rendered inline in `App.vue` via `v-else-if="$route.path === '/overview'"`, NOT through `<router-view>`. This is unaffected by the change — when `waitingForData` returns false, the template falls through to the `SiteOverview` branch (which checks the path), then to `<router-view>`. The behavior is identical to before.

### 3. Adapt Help Modal and Remove evcc.io Link

**File**: `web/assets/js/components/Top/Navigation.vue`

- Remove the "evcc.io" external link entirely (lines 123-131)

**File**: `web/assets/js/components/HelpModal.vue`

Reduce the modal content to three items:
- **Keep**: Docs link (docs.evcc.io) — relevant for all evcc users
- **Keep**: GitHub Discussions link — community support
- **Add**: Link to evcc-hub GitHub repo (https://github.com/tjarkste/evcc-hub)
- **Remove**: "Logs anzeigen" button — doesn't work via the hub
- **Remove**: "Issue erstellen" button — refers to local instance
- **Remove**: "Neustart" button and confirm modal — hub users can't restart the local instance

**Cleanup**: When removing the Restart button, also remove:
- The `confirmRestartModal` Teleport section (lines 90-132)
- The `performRestart` import
- The `openConfirmRestartModal` and `restartConfirmed` methods

### 4. Full i18n Migration

**Translation files**: `web/i18n/en.json`, `web/i18n/de.json`

All hardcoded user-facing strings will be replaced with `$t()` calls using a `hub.*` namespace to clearly separate hub-specific keys from original evcc keys (which are loaded from the API).

**Namespace structure**:
- `hub.connection.*` — ConnectionStatus, WaitingForData
- `hub.auth.*` — LoginView (login, register, onboarding flow)
- `hub.settings.*` — SettingsView (profile, password, etc.) AND "Profil & Einstellungen" in Navigation
- `hub.sites.*` — SiteManager, SiteOverview, SiteCredentialsModal
- `hub.nav.*` — Navigation-only strings (e.g. "Optimize 🧪")
- `hub.error.*` — ErrorBoundary
- `hub.debug.*` — Optimize, Energy (dev pages), BatteryConfigurationTable
- `hub.issue.*` — Issue page
- `hub.legal.*` — Loading text for Impressum/Datenschutz/Nutzungsbedingungen
- `hub.general.*` — Shared strings (e.g. "Kopiert!", "Kopieren", "Laden...")

**Languages**: DE and EN only. All other 24 supported languages fall back to EN automatically (Vue i18n fallback behavior).

**Additional i18n fix**: `SettingsView.vue` line 117 uses hardcoded `toLocaleDateString('de-DE', ...)` — this should use the current i18n locale instead.

**Legal page loading text**: The legal pages (`Datenschutz.vue`, `Impressum.vue`, `Nutzungsbedingungen.vue`) use `ref('<p class="text-muted">Lädt...</p>')` with `v-html`. These should be refactored to use conditional template rendering (`v-if="loading"` with `$t()`) instead of injecting translated text into a ref.

**Complete file inventory**:

| File | Hardcoded Strings |
|---|---|
| `components/ConnectionStatus.vue` | "Reconnecting...", "Offline", "(last update {X}s ago)", "(last update {X}m ago)" |
| `components/WaitingForData.vue` | "Verbinde mit MQTT-Broker...", "Verbunden — warte auf Daten von deiner evcc-Instanz...", "Keine Daten empfangen.", "Ist deine evcc-Instanz online und korrekt konfiguriert?", "Laden..." |
| `components/Top/Navigation.vue` | "Profil & Einstellungen", "Optimize 🧪" |
| `components/ErrorBoundary.vue` | "failed to render.", "Retry", "Unknown error" |
| `components/HelpModal.vue` | Content restructuring (see section 3), plus all remaining labels |
| `components/SiteCredentialsModal.vue` | "MQTT-Zugangsdaten — {siteName}", "Diese Daten brauchst du...", "Broker URL", "Broker Port", "MQTT-Benutzername", "MQTT-Passwort", "Topic Prefix", "Kopiert!", "Kopieren", "Verbergen", "Anzeigen", "Schließen", "Zugangsdaten konnten nicht geladen werden.", "Lädt…" |
| `components/Energyflow/Visualization.vue` | "In", "Out" |
| `components/Optimize/BatteryConfigurationTable.vue` | "Charge / Discharge", "Charge", "Discharge", "None", "Battery", "State of Charge", "SoC Range", "Energy Value", "Power Range", "Max Discharge", "Grid Interaction", "Demand Profile", "SoC Goals", "steps", "goals" |
| `components/Config/AuthSuccessBanner.vue` | "Unknown" (low priority — edge-case fallback) |
| `views/App.vue` | "Laden..." |
| `views/LoginView.vue` | "Dein evcc-Dashboard, von überall erreichbar.", "Open Source auf GitHub →", "Anmelden", "Wird angemeldet...", "Noch kein Konto? Registrieren", "Kostenlosen Account erstellen", "Registrieren", "Wird registriert...", "Bereits ein Konto? Anmelden", "Konto erstellt!", "Verbinde jetzt deine evcc-Instanz.", "evcc installiert?", "Falls noch nicht:", "evcc installieren", "MQTT-Konfiguration hinzufügen", "Füge diese Zeilen in deine evcc.yaml ein:", "evcc neu starten", "Weiter zum Dashboard", "Kopiert!", "Kopieren" |
| `views/SettingsView.vue` | "Profil & Einstellungen", "Konto", "E-Mail-Adresse", "Registriert seit", "Lädt...", "Aktuelles Passwort", "Neues Passwort", "Neues Passwort bestätigen", "Passwort speichern", "Speichern…", "Passwort ändern", "Abmelden", "Passwort erfolgreich geändert...", "Die neuen Passwörter stimmen nicht überein.", "Profil konnte nicht geladen werden.", "Passwort konnte nicht geändert werden.", hardcoded `de-DE` locale |
| `views/SiteManager.vue` | "Meine Standorte", "Auswählen", "Aktiv", "Löschen", "Neuen Standort hinzufügen", "Name (z.B. Ferienhaus)", "Hinzufügen", "wurde erstellt.", "Füge diese Zeilen in deine evcc.yaml ein:", "Kopiert!", "Kopieren", "Zurück zum Dashboard", "Standorte konnten nicht geladen werden.", "Standort konnte nicht erstellt werden.", "Standort konnte nicht gelöscht werden." |
| `views/SiteOverview.vue` | "Meine Standorte", "Anzeigen", "MQTT", "Aktiv", "Standorte verwalten" |
| `views/Optimize.vue` | "Optimize Debug", "This page is for development purposes only...", "Result: Charging Plan", "saved", "Result: SoC Projection", "Input: Grid Prices", "Input: Battery", "charge efficiency", "discharge efficiency", "Time Series", "Raw Data", "Request:", "Response:", "nothing to see here" |
| `views/Energy.vue` | "Energy Overview", "This page is work in progress.", "nothing to see here" |
| `views/Issue.vue` | "Please write your issue in English...", "Brief description of the problem", "Discussion"/"Issue", "Failed to load configuration", "Failed to load API configuration", "Failed to load logs", "Failed to load system state", plus form placeholders |
| `views/Datenschutz.vue`, `Impressum.vue`, `Nutzungsbedingungen.vue` | "Lädt..." each (via `v-html` ref — needs refactor) |

**Note on legal page link text**: The footer links "Impressum", "Datenschutz", "Nutzungsbedingungen" in `App.vue` are German legal terms and will remain as-is (they are proper nouns, not translatable).

## Out of Scope

- Translating legal pages content (Impressum, Datenschutz, Nutzungsbedingungen) — these are loaded dynamically
- Adding translations beyond DE and EN
- Translating console.log messages or technical strings
- Refactoring `SiteOverview` inline rendering in `App.vue` to use `<router-view>`
