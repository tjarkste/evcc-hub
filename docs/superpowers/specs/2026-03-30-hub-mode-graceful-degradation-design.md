# Hub Mode Graceful Degradation — Design Spec

**Date:** 2026-03-30
**Status:** Approved
**Scope:** Frontend-only changes — no backend required

---

## Problem

The evcc hub dashboard includes pages (Config, Log, Issue/Report-a-problem) that were originally designed for direct REST API access to a local evcc instance. In the cloud hub context, the evcc instance is behind NAT and the only communication channel is an outbound MQTT connection the instance opens to the hub's broker.

Stock evcc publishes live telemetry via MQTT (loadpoints, vehicles, battery, site state) but does not expose configuration trees, log streams, or a restart command over MQTT. Three features are consequently broken or misleading:

1. **Config page** — REST config calls return `{}`, leaving device sections empty; crashes occur in sub-components
2. **Log page** — REST log calls return `{}`, the page is empty or crashes
3. **Restart button** — silently does nothing

The fix must work with stock evcc (no changes to the home installation).

---

## Constraints

- Stock evcc only — no companion agent, no reverse tunnel
- Frontend-only solution — no new backend endpoints required
- MQTT-published data (loadpoints, vehicles, battery, site) is available and must still be shown
- REST-only data (meter configs, tariff configs, logs) cannot be retrieved — must be replaced with a disclaimer
- All disclaimers use existing i18n infrastructure (de + en)

---

## Architecture

### Hub-mode flag

A single exported constant **does not yet exist** and must be created as the first implementation step. Add it to `web/assets/js/api.ts`:

```ts
export const HUB_MODE = true
```

This is the authoritative source of truth for cloud/hub mode. It lives in `api.ts` because that file is already the boundary between "real REST backend" and "MQTT stub". Any component that needs to branch on hub mode imports `{ HUB_MODE }` from `@/api`.

---

## Component Changes

### New component: `web/assets/js/components/HubModeNotice.vue`

A small reusable notice component used wherever REST content is replaced.

**Props:**
- `message: string` — the **translated string** to display. The parent resolves the i18n key via `$t()` and passes the result. The component does not call `$t()` internally.

**Appearance:**
- Styled as a muted info block (Bootstrap `alert alert-secondary` or `text-muted` with padding)
- One paragraph of text, no action buttons

**Correct usage:**
```html
<HubModeNotice :message="$t('hub.cloudNotAvailable.config')" />
```

**Incorrect usage (do not do this):**
```html
<HubModeNotice message="hub.cloudNotAvailable.config" />  <!-- raw key, not translated -->
```

---

### `web/assets/js/views/Config.vue`

Config.vue sections fall into three categories in hub mode:

**1. MQTT-backed — show read-only (data is available via store):**
- Loadpoints
- Vehicles

Treatment: `DeviceCard` already accepts `:editable` prop. Pass `:editable="false"` in hub mode. `NewDeviceButton` components (Add loadpoint, Add vehicle) are hidden with `v-if="!hubMode"`. No disclaimer needed — the data is real and up-to-date.

**2. REST-only — replace with `HubModeNotice`:**
- General settings (`<GeneralConfig>` component)
- Grid meter section
- PV/Battery meters section
- Additional meters section (Aux/Ext)
- Tariffs section
- Integrations section — includes `<AuthProvidersCard>` (OAuth provider flows) plus DeviceCards for MQTT, Influx, Messaging, Circuits, ModbusProxy, HEMS. The **entire section content block** (the `<div class="p-0 config-list">` that contains `AuthProvidersCard` and all the DeviceCards) is replaced with `HubModeNotice`. `AuthProvidersCard` is treated as REST/cloud-only — OAuth flows require direct instance access and have no hub-mode equivalent.
- Services section (OCPP, SHM, EEBUS)

Treatment: Each section's content block is wrapped `v-if="!hubMode"` with a `v-else` that renders `<HubModeNotice>`. No changes are needed inside child components like `GeneralConfig.vue`, `OcppModal.vue`, `AuthProvidersCard.vue`, etc. — they are simply not rendered in hub mode.

**3. System section (partial):**
- Logs link → kept active (navigates to Log page which handles hub mode itself)
- Report a problem link → kept active (Issue page handles hub mode itself)
- Restart button → `disabled` in hub mode; Bootstrap `title` attribute set to `$t('hub.cloudNotAvailable.restartTooltip')`
- Backup/Restore button → `disabled` in hub mode; `title` set to `$t('hub.cloudNotAvailable.backupTooltip')`

**`mounted()` / `created()` guards in Config.vue:**
Config.vue's `loadConfig()` and related REST calls already return `{}` silently via the stub. No explicit guard is needed — the REST calls produce no crash and no visible effect in hub mode because the REST-only sections are not rendered.

---

### `web/assets/js/views/Log.vue`

The entire log interface (filter bar, log lines, download button, area selector) is replaced in hub mode with `HubModeNotice`. The `TopHeader` and page chrome remain.

**Template:** The `<div class="logs ...">` block is wrapped with `v-if="!hubMode"`. A `v-else` renders `<HubModeNotice :message="$t('hub.cloudNotAvailable.logs')" />`.

**`mounted()` guard:** The polling interval and area fetch must also be gated. In `mounted()`, wrap the interval start and area fetch:
```js
if (!HUB_MODE) {
  this.startInterval()
  this.updateAreas()
}
```
Without this guard, Log.vue sets up a `setInterval` polling loop on every page visit even though no log content is ever displayed, wasting resources for the lifetime of the component.

---

### `web/assets/js/views/Issue.vue`

Issue.vue's `mounted()` calls five methods. Their hub mode treatment:

| Method | REST call | Hub mode action |
|---|---|---|
| `loadYamlConfig()` | Yes | Leave running — returns `{}`, sets config string to `''`, no visible effect since the YAML section is hidden by hub mode `v-if` |
| `loadUiConfig()` | Yes | Leave running — same rationale |
| `loadState()` | Yes | Leave running — same rationale |
| `loadLogs()` | Yes | **Guard with `if (!HUB_MODE)`** — result feeds the log section which must be hidden |
| `updateAreas()` | Yes | **Guard with `if (!HUB_MODE)`** — result feeds `logAvailableAreas` which drives the area selector |

Only the two log-related calls are guarded. The config/state calls are intentionally left running because they return empty strings silently and their output sections are already hidden in hub mode by `v-if`.

**Template changes:**
- Add `<HubModeNotice :message="$t('hub.cloudNotAvailable.issue')" />` at the top of the form body, shown only `v-if="hubMode"`
- The log-area selector and log-content preview section are wrapped `v-if="!hubMode"`

---

## i18n Keys

**Pre-existing keys (do not add again):** `Issue.vue` references `hub.issue.loadConfigError`, `hub.issue.loadLogsError`, and `hub.issue.loadStateError`. These keys already exist in both locale files and are unrelated to this work. Do not modify or duplicate them.

**New keys — all five below do not yet exist** in either locale file and must be added as part of this work under the existing `hub` namespace:

Added to both `web/i18n/de.json` and `web/i18n/en.json` under `hub.cloudNotAvailable`:

| Key | German | English |
|---|---|---|
| `hub.cloudNotAvailable.config` | „Diese Einstellungen werden von deiner lokalen evcc-Installation verwaltet. Die Konfiguration über die Cloud wird von evcc noch nicht unterstützt." | "These settings are managed by your local evcc installation. Remote configuration via the cloud is not yet supported by evcc." |
| `hub.cloudNotAvailable.logs` | „Log-Zugriff ist in der Cloud-Ansicht nicht verfügbar. Öffne dein lokales evcc-Dashboard, um Logs einzusehen." | "Log access is not available in the cloud view. Open your local evcc dashboard to view logs." |
| `hub.cloudNotAvailable.issue` | „Logs können in der Cloud-Ansicht nicht automatisch eingebunden werden. Bitte beschreibe das Problem manuell." | "Logs cannot be included automatically in the cloud view. Please describe the problem manually." |
| `hub.cloudNotAvailable.restartTooltip` | „Neustart ist in der Cloud-Ansicht nicht verfügbar." | "Restart is not available in the cloud view." |
| `hub.cloudNotAvailable.backupTooltip` | „Backup/Wiederherstellung ist in der Cloud-Ansicht nicht verfügbar." | "Backup/restore is not available in the cloud view." |

---

## Data Flow Summary

```
MQTT broker
    │
    ▼
store.state (populated by mqtt.ts)
    │
    ├─▶ Config.vue — Loadpoints / Vehicles     ← renders read-only with real MQTT data
    │
    ├─▶ Config.vue — Meters / Tariffs /        ← replaced by HubModeNotice
    │   General / Integrations / Services
    │
    ├─▶ Log.vue                                ← replaced by HubModeNotice; polling skipped
    │
    └─▶ Issue.vue — log section hidden         ← HubModeNotice at top; log calls skipped
```

---

## Files Changed

| File | Change | Net-new? |
|---|---|---|
| `web/assets/js/api.ts` | Add `export const HUB_MODE = true` | Constant is new |
| `web/assets/js/components/HubModeNotice.vue` | Create new reusable notice component | New file |
| `web/assets/js/views/Config.vue` | Import `HUB_MODE`; hub mode branching for all sections; restart/backup disabled | Modifying existing |
| `web/assets/js/views/Log.vue` | Import `HUB_MODE`; replace log UI + guard `mounted()` calls | Modifying existing |
| `web/assets/js/views/Issue.vue` | Import `HUB_MODE`; hide log section + guard `loadLogs()`/`updateAreas()` | Modifying existing |
| `web/i18n/de.json` | Add 5 `hub.cloudNotAvailable.*` keys (all new) | Keys are new |
| `web/i18n/en.json` | Add 5 `hub.cloudNotAvailable.*` keys (all new) | Keys are new |

**Not changed:** `GeneralConfig.vue`, `OcppModal.vue`, and all other child components. They are conditionally excluded from the render tree by the parent `v-if` in Config.vue and require no internal modification.

---

## Out of Scope

- Backend proxy for REST config data (requires evcc to support MQTT-based config API — not yet available)
- Log streaming via MQTT (not part of evcc's MQTT protocol)
- Restart via MQTT (no MQTT command exists in stock evcc)
- Any changes to the evcc home installation
- Changes to child config components (GeneralConfig, OcppModal, etc.)
