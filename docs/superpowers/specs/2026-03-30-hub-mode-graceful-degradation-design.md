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

A single exported constant is added to `web/assets/js/api.ts`:

```ts
export const HUB_MODE = true
```

This is the authoritative source of truth for cloud/hub mode. It lives in `api.ts` because that file is already the boundary between "real REST backend" and "MQTT stub". Any component that needs to branch on hub mode imports it from here.

---

## Component Changes

### `web/assets/js/views/Config.vue`

Config.vue has two classes of sections:

**MQTT-backed (data available):** Loadpoints, Vehicles
**REST-only (data not available):** General settings, Grid meter, PV/Battery meters, Additional meters, Tariffs, System actions

**Loadpoints and Vehicles sections:**
- DeviceCard components already accept an `:editable` prop that controls edit button visibility
- In hub mode: pass `:editable="false"` to all DeviceCards — cards render with live MQTT data but without edit affordances
- `NewDeviceButton` components (Add loadpoint, Add vehicle) are hidden with `v-if="!hubMode"`
- No disclaimer needed — the data is real and visible

**REST-only sections (General, Meters, Tariffs):**
- Each section's content block is replaced with `v-if="!hubMode"` / `v-else`
- The `v-else` renders a `HubModeNotice` component (see below) with the config-specific message

**System section:**
- Restart button: `disabled` attribute added in hub mode; a Bootstrap `title` tooltip reads the `hub.cloudNotAvailable.restartTooltip` i18n key
- Backup/Restore button: same — disabled in hub mode
- Logs link and Report-a-problem link: kept active (they navigate to the Log/Issue pages which handle hub mode themselves)

---

### `web/assets/js/views/Log.vue`

The entire log interface (filter bar, log lines, download button, area selector) is replaced in hub mode with the `HubModeNotice` component. The `TopHeader` and page chrome remain so navigation works normally.

Concretely: the outermost `<div class="logs ...">` is wrapped in a `v-if="!hubMode"` / `v-else` that renders the notice instead.

---

### `web/assets/js/views/Issue.vue`

The log-collection portion of the issue report form is hidden in hub mode:
- `updateAreas()` and `updateLogs()` calls are skipped
- The log-area selector and log-content section of the form are hidden with `v-if="!hubMode"`
- A notice appears at the top of the form explaining that logs cannot be auto-included

The rest of the form (description, contact fields, submit) continues to work.

---

### New component: `web/assets/js/components/HubModeNotice.vue`

A small reusable notice component used wherever REST content is replaced:

**Props:**
- `message: string` — the i18n key to display (passed as a translated string by the parent)

**Appearance:**
- Styled as a muted info block (Bootstrap `text-muted` + small icon)
- Single line of text, no action buttons

Example usage:
```html
<HubModeNotice :message="$t('hub.cloudNotAvailable.config')" />
```

---

## i18n Keys

Added to both `web/i18n/de.json` and `web/i18n/en.json` under `hub.cloudNotAvailable`:

| Key | German | English |
|---|---|---|
| `hub.cloudNotAvailable.config` | „Diese Einstellungen werden von deiner lokalen evcc-Installation verwaltet. Die Konfiguration über die Cloud wird von evcc noch nicht unterstützt." | "These settings are managed by your local evcc installation. Remote configuration via the cloud is not yet supported by evcc." |
| `hub.cloudNotAvailable.logs` | „Log-Zugriff ist in der Cloud-Ansicht nicht verfügbar. Öffne dein lokales evcc-Dashboard, um Logs einzusehen." | "Log access is not available in the cloud view. Open your local evcc dashboard to view logs." |
| `hub.cloudNotAvailable.issue` | „Logs können in der Cloud-Ansicht nicht automatisch eingebunden werden. Beschreibe das Problem manuell." | "Logs cannot be included automatically in the cloud view. Please describe the problem manually." |
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
    ├─▶ Config.vue — Loadpoints / Vehicles sections  ← renders with real data
    │
    ├─▶ Config.vue — Meters / Tariffs / General      ← replaced by HubModeNotice
    │
    ├─▶ Log.vue                                       ← replaced by HubModeNotice
    │
    └─▶ Issue.vue — log section hidden                ← HubModeNotice at top
```

---

## Files Changed

| File | Change |
|---|---|
| `web/assets/js/api.ts` | Add `export const HUB_MODE = true` |
| `web/assets/js/components/HubModeNotice.vue` | New reusable notice component |
| `web/assets/js/views/Config.vue` | Hub mode branching for all sections; restart/backup disabled |
| `web/assets/js/views/Log.vue` | Replace log UI with HubModeNotice in hub mode |
| `web/assets/js/views/Issue.vue` | Hide log section, show notice at top |
| `web/i18n/de.json` | Add `hub.cloudNotAvailable.*` keys |
| `web/i18n/en.json` | Add `hub.cloudNotAvailable.*` keys |

---

## Out of Scope

- Backend proxy for REST config data (requires evcc to support MQTT-based config API — not yet available)
- Log streaming via MQTT (not part of evcc's MQTT protocol)
- Restart via MQTT (no MQTT command exists in stock evcc)
- Any changes to the evcc home installation
