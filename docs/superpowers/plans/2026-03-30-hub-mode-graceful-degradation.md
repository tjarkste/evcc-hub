# Hub Mode Graceful Degradation — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace broken REST-dependent sections on Config, Log, and Issue pages with a clean read-only / disclaimer experience when running in cloud hub mode.

**Architecture:** Export a `HUB_MODE = true` constant from `api.ts` as the single source of truth. A new `HubModeNotice.vue` component displays a translated disclaimer wherever REST content is replaced. Config.vue keeps MQTT-backed sections (loadpoints, vehicles) as read-only cards while hiding REST-only sections behind the notice. Log.vue and Issue.vue replace their log content with the notice and guard their polling/fetch calls.

**Tech Stack:** Vue 3 Options API, TypeScript, Bootstrap 5, Vue i18n, Vitest for unit tests. All tests run with `cd web && npm test`.

---

## File Map

| File | Action | Purpose |
|---|---|---|
| `web/assets/js/api.ts` | Modify | Add `export const HUB_MODE = true` |
| `web/assets/js/services/api.test.ts` | Modify | Add test asserting `HUB_MODE === true` |
| `web/assets/js/components/HubModeNotice.vue` | Create | Reusable notice component |
| `web/assets/js/views/Config.vue` | Modify | Hub mode branching for all sections; disable restart/backup |
| `web/assets/js/views/Log.vue` | Modify | Replace log UI + guard `mounted()` |
| `web/assets/js/views/Issue.vue` | Modify | Hide log section + guard `loadLogs()`/`updateAreas()` |
| `web/i18n/de.json` | Modify | Add 5 `hub.cloudNotAvailable.*` keys |
| `web/i18n/en.json` | Modify | Add 5 `hub.cloudNotAvailable.*` keys |

---

## Task 1: Add `HUB_MODE` flag to `api.ts`

**Files:**
- Modify: `web/assets/js/api.ts`
- Modify: `web/assets/js/services/api.test.ts`

- [ ] **Step 1: Write the failing test**

Open `web/assets/js/services/api.test.ts`. Line 2 already has:
```ts
import { restPathToMqttTopic } from '../api'
```
Change it to add `HUB_MODE` to the same import:
```ts
import { restPathToMqttTopic, HUB_MODE } from '../api'
```
Then add at the bottom of the file:
```ts
describe('HUB_MODE', () => {
  test('is true in hub/cloud mode', () => {
    expect(HUB_MODE).toBe(true)
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd web && npm test -- --reporter=verbose services/api.test.ts
```
Expected: FAIL with `HUB_MODE is not exported` or similar.

- [ ] **Step 3: Add `HUB_MODE` export to `api.ts`**

In `web/assets/js/api.ts`, insert this line as line 6 — immediately after the single import on line 4 and before the `interface MqttMapping` declaration on line 6:

```ts
export const HUB_MODE = true
```

- [ ] **Step 4: Run test to verify it passes**

```bash
cd web && npm test -- --reporter=verbose services/api.test.ts
```
Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add web/assets/js/api.ts web/assets/js/services/api.test.ts
git commit -m "feat: add HUB_MODE flag to api.ts"
```

---

## Task 2: Add i18n keys for hub mode notices

**Files:**
- Modify: `web/i18n/de.json`
- Modify: `web/i18n/en.json`

These five keys are entirely new — they do not exist yet. Do not touch the pre-existing `hub.issue.*` keys.

- [ ] **Step 1: Add keys to `de.json`**

Open `web/i18n/de.json`. Find the `"hub"` top-level key. Inside it, add a new `"cloudNotAvailable"` object. Place it after the existing `"issue"` object (search for `"issue":` inside the `hub` namespace to find the right location):

```json
"cloudNotAvailable": {
  "config": "Diese Einstellungen werden von deiner lokalen evcc-Installation verwaltet. Die Konfiguration über die Cloud wird von evcc noch nicht unterstützt.",
  "logs": "Log-Zugriff ist in der Cloud-Ansicht nicht verfügbar. Öffne dein lokales evcc-Dashboard, um Logs einzusehen.",
  "issue": "Logs können in der Cloud-Ansicht nicht automatisch eingebunden werden. Bitte beschreibe das Problem manuell.",
  "restartTooltip": "Neustart ist in der Cloud-Ansicht nicht verfügbar.",
  "backupTooltip": "Backup/Wiederherstellung ist in der Cloud-Ansicht nicht verfügbar."
}
```

- [ ] **Step 2: Add keys to `en.json`**

Open `web/i18n/en.json`. Find the `"hub"` top-level key. Add the same `"cloudNotAvailable"` object after the `"issue"` object:

```json
"cloudNotAvailable": {
  "config": "These settings are managed by your local evcc installation. Remote configuration via the cloud is not yet supported by evcc.",
  "logs": "Log access is not available in the cloud view. Open your local evcc dashboard to view logs.",
  "issue": "Logs cannot be included automatically in the cloud view. Please describe the problem manually.",
  "restartTooltip": "Restart is not available in the cloud view.",
  "backupTooltip": "Backup/restore is not available in the cloud view."
}
```

- [ ] **Step 3: Validate i18n**

```bash
cd web && npm run lint:i18n
```
Expected: no missing key warnings. Fix any JSON syntax errors if reported.

- [ ] **Step 4: Commit**

```bash
git add web/i18n/de.json web/i18n/en.json
git commit -m "feat: add hub.cloudNotAvailable i18n keys (de + en)"
```

---

## Task 3: Create `HubModeNotice.vue` component

**Files:**
- Create: `web/assets/js/components/HubModeNotice.vue`

This component accepts a single `message` prop — already a **translated string** (the parent calls `$t()` and passes the result). The component does not do its own i18n lookup.

- [ ] **Step 1: Create the component**

Create `web/assets/js/components/HubModeNotice.vue`:

```vue
<template>
	<div class="hub-mode-notice alert alert-secondary d-flex align-items-start gap-2 my-2">
		<span class="text-muted small">{{ message }}</span>
	</div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'

export default defineComponent({
	name: 'HubModeNotice',
	props: {
		message: {
			type: String,
			required: true,
		},
	},
})
</script>

<style scoped>
.hub-mode-notice {
	border-radius: 0.5rem;
}
</style>
```

- [ ] **Step 2: Verify no build errors**

```bash
cd web && npm run lint:tsc
```
Expected: no TypeScript errors.

- [ ] **Step 3: Commit**

```bash
git add web/assets/js/components/HubModeNotice.vue
git commit -m "feat: add HubModeNotice reusable component"
```

---

## Task 4: Update `Log.vue` — replace UI and guard polling

**Files:**
- Modify: `web/assets/js/views/Log.vue`

Two changes: (1) replace the log content div with a `v-if/v-else` pairing, (2) guard `mounted()` so the polling interval never starts in hub mode.

- [ ] **Step 1: Import `HUB_MODE` and `HubModeNotice`**

In `Log.vue`'s `<script>` section, add to the imports:

```ts
import { HUB_MODE } from '../api'
import HubModeNotice from '../components/HubModeNotice.vue'
```

Add `HubModeNotice` to the `components` object:

```ts
components: {
    TopHeader: Header,
    Play,
    Record,
    MultiSelect,
    HubModeNotice,
},
```

Add `hubMode` as a data property:

```ts
data() {
    return {
        // ... existing properties ...
        hubMode: HUB_MODE,
    }
}
```

- [ ] **Step 2: Gate the logs div in the template**

Find the opening tag of the logs container (line 5 of the template):
```html
<div class="logs d-flex flex-column overflow-hidden flex-grow-1 px-4 mx-2 mx-sm-4">
```

Add `v-if="!hubMode"` as an attribute to that opening tag only — do not touch any inner content:
```html
<div v-if="!hubMode" class="logs d-flex flex-column overflow-hidden flex-grow-1 px-4 mx-2 mx-sm-4">
```

Then immediately after the closing `</div>` of that element (which closes the entire logs block), add the hub mode fallback sibling:
```html
<div v-else class="px-4 mx-2 mx-sm-4 py-5">
    <HubModeNotice :message="$t('hub.cloudNotAvailable.logs')" />
</div>
```

All content inside the logs div is untouched.

- [ ] **Step 3: Guard `mounted()`**

Find the `mounted()` hook:

```ts
mounted() {
    this.startInterval();
    this.updateAreas();
},
```

Replace with:

```ts
mounted() {
    if (!HUB_MODE) {
        this.startInterval();
        this.updateAreas();
    }
},
```

- [ ] **Step 4: Run lint and type check**

```bash
cd web && npm run lint:tsc
```
Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add web/assets/js/views/Log.vue
git commit -m "feat: replace log UI with HubModeNotice in hub mode; skip polling"
```

---

## Task 5: Update `Issue.vue` — hide log section and guard log calls

**Files:**
- Modify: `web/assets/js/views/Issue.vue`

Three changes: (1) import HUB_MODE + HubModeNotice, (2) show notice at top of form in hub mode, (3) wrap the logs `IssueAdditionalItem` with `v-if="!hubMode"`, (4) guard `loadLogs()` and `updateAreas()` in `mounted()`.

- [ ] **Step 1: Import `HUB_MODE` and `HubModeNotice`**

In `Issue.vue`'s `<script>` section, add:

```ts
import { HUB_MODE } from '../api'
import HubModeNotice from '../components/HubModeNotice.vue'
```

Add `HubModeNotice` to the `components` object (find the existing `components: {` declaration and add it).

Add `hubMode` as a data property:

```ts
data() {
    return {
        // ... existing properties ...
        hubMode: HUB_MODE,
    }
}
```

- [ ] **Step 2: Add notice at top of form in template**

Find the opening `<form v-if="helpType"` tag. Immediately inside it, before the "Essential Form Section" comment, insert:

```html
<div v-if="hubMode" class="mb-4">
    <HubModeNotice :message="$t('hub.cloudNotAvailable.issue')" />
</div>
```

- [ ] **Step 3: Wrap the logs `IssueAdditionalItem` with `v-if`**

Find the `<!-- Logs Section with Special Controls -->` comment and the `<IssueAdditionalItem id="issueLogs"` block. Wrap the entire `IssueAdditionalItem` for logs with `v-if="!hubMode"`:

```html
<IssueAdditionalItem
    v-if="!hubMode"
    id="issueLogs"
    ...
>
```

(Add `v-if="!hubMode"` as a prop on the existing component tag — do not restructure the surrounding code.)

- [ ] **Step 4: Guard `mounted()`**

Find the `async mounted()` hook:

```ts
async mounted() {
    this.loadYamlConfig();
    this.loadUiConfig();
    this.loadState();
    this.loadLogs();
    this.updateAreas();
},
```

Replace with:

```ts
async mounted() {
    this.loadYamlConfig();
    this.loadUiConfig();
    this.loadState();
    if (!HUB_MODE) {
        this.loadLogs();
        this.updateAreas();
    }
},
```

The `loadYamlConfig`, `loadUiConfig`, and `loadState` calls are intentionally left running — they return empty strings silently via the stub and their UI sections are already hidden by other `v-if` conditions.

- [ ] **Step 5: Run lint and type check**

```bash
cd web && npm run lint:tsc
```
Expected: no errors.

- [ ] **Step 6: Commit**

```bash
git add web/assets/js/views/Issue.vue
git commit -m "feat: hide log section in Issue.vue in hub mode; guard log calls"
```

---

## Task 6: Update `Config.vue` — read-only MQTT sections, disable REST sections

**Files:**
- Modify: `web/assets/js/views/Config.vue`

This is the largest change. Work section by section. Read the current template carefully before editing — many sections share similar `<div class="p-0 config-list">` wrappers.

- [ ] **Step 1: Import `HUB_MODE` and `HubModeNotice`**

In `Config.vue`'s `<script>` section, add:

```ts
import { HUB_MODE } from '../api'
import HubModeNotice from '../components/HubModeNotice.vue'
```

Add `HubModeNotice` to the `components` object (it's a large object — add `HubModeNotice,` anywhere inside it).

Add `hubMode` as a data property in `data()`:

```ts
hubMode: HUB_MODE,
```

- [ ] **Step 2: General section — wrap `GeneralConfig` with `v-if`**

Find:
```html
<h2 class="my-4 mt-5">{{ $t("config.section.general") }}</h2>
<GeneralConfig
    :experimental="experimental"
    :sponsor-error="hasClassError('sponsorship')"
    @site-changed="siteChanged"
/>
```

Replace with:
```html
<h2 class="my-4 mt-5">{{ $t("config.section.general") }}</h2>
<template v-if="!hubMode">
    <GeneralConfig
        :experimental="experimental"
        :sponsor-error="hasClassError('sponsorship')"
        @site-changed="siteChanged"
    />
</template>
<HubModeNotice v-else :message="$t('hub.cloudNotAvailable.config')" />
```

- [ ] **Step 3: Loadpoints section — make DeviceCards read-only, hide Add button**

Find `:editable="!!loadpoint.id"` and change to:
```html
:editable="!hubMode && !!loadpoint.id"
```

Find `<NewDeviceButton data-testid="add-loadpoint"` and add `v-if="!hubMode"`:
```html
<NewDeviceButton
    v-if="!hubMode"
    data-testid="add-loadpoint"
    :title="$t('config.main.addLoadpoint')"
    @click="openModal('loadpoint')"
/>
```

- [ ] **Step 4: Vehicles section — make DeviceCards read-only, hide Add button**

Find `:editable="vehicle.id >= 0"` and change to:
```html
:editable="!hubMode && vehicle.id >= 0"
```

Find `<NewDeviceButton data-testid="add-vehicle"` and add `v-if="!hubMode"`:
```html
<NewDeviceButton
    v-if="!hubMode"
    data-testid="add-vehicle"
    :title="$t('config.main.addVehicle')"
    @click="openModal('vehicle')"
/>
```

- [ ] **Step 5: Grid meter section — gate the config-list div**

Find the `<div class="p-0 config-list">` that immediately follows the `config.section.grid` h2 heading. Add `v-if="!hubMode"` to that opening tag only — do not touch any inner content:
```html
<div v-if="!hubMode" class="p-0 config-list">
```
Immediately after the closing `</div>` of that element, add:
```html
<HubModeNotice v-else :message="$t('hub.cloudNotAvailable.config')" />
```

- [ ] **Step 6: PV/Battery meters section — gate the config-list div**

Find the `<div class="p-0 config-list">` that immediately follows the `config.section.meter` h2 heading. Add `v-if="!hubMode"` to that opening tag:
```html
<div v-if="!hubMode" class="p-0 config-list">
```
Immediately after its closing `</div>`, add:
```html
<HubModeNotice v-else :message="$t('hub.cloudNotAvailable.config')" />
```

- [ ] **Step 7: Additional meters section — gate the config-list div**

Find the `<div class="p-0 config-list">` that immediately follows the `config.section.additionalMeter` h2 heading. Add `v-if="!hubMode"` to that opening tag:
```html
<div v-if="!hubMode" class="p-0 config-list">
```
Immediately after its closing `</div>`, add:
```html
<HubModeNotice v-else :message="$t('hub.cloudNotAvailable.config')" />
```

- [ ] **Step 8: Tariffs section — wrap both inner divs with a `<template>`**

The tariffs section currently has:
```html
<h2 class="my-4 mt-5">{{ $t("config.tariff.title") }}</h2>
<div v-if="!!tariffsYamlSource" class="p-0 config-list">...</div>
<div v-else class="p-0 config-list">...</div>
```

Wrap both existing `<div>` elements (the `v-if` and `v-else` pair) inside a new `<template v-if="!hubMode">` ... `</template>`, and add the `HubModeNotice` sibling after the closing `</template>`. Do not touch any inner content of either div:

```html
<h2 class="my-4 mt-5">{{ $t("config.tariff.title") }}</h2>
<template v-if="!hubMode">
    <div v-if="!!tariffsYamlSource" class="p-0 config-list">
        [unchanged inner content]
    </div>
    <div v-else class="p-0 config-list">
        [unchanged inner content]
    </div>
</template>
<HubModeNotice v-else :message="$t('hub.cloudNotAvailable.config')" />
```

In practice: insert `<template v-if="!hubMode">` on the line before `<div v-if="!!tariffsYamlSource"`, and add `</template>` + `<HubModeNotice .../>` after the closing `</div>` of the `v-else` div (the one ending around line 229). All inner content is unchanged.

- [ ] **Step 9: Integrations section — gate the config-list div**

Find the `<div class="p-0 config-list">` that immediately follows the `config.section.integrations` h2 heading (this div contains `AuthProvidersCard` and all integration DeviceCards). Add `v-if="!hubMode"` to that opening tag only — do not touch any inner content:
```html
<div v-if="!hubMode" class="p-0 config-list">
```
Immediately after the closing `</div>` of that element, add:
```html
<HubModeNotice v-else :message="$t('hub.cloudNotAvailable.config')" />
```

- [ ] **Step 10: Services section — gate the config-list div**

Find the `<div class="p-0 config-list">` that immediately follows the `config.section.services` h2 heading (this div contains OCPP, SHM, EEBUS DeviceCards). Add `v-if="!hubMode"` to that opening tag only:
```html
<div v-if="!hubMode" class="p-0 config-list">
```
Immediately after its closing `</div>`, add:
```html
<HubModeNotice v-else :message="$t('hub.cloudNotAvailable.config')" />
```

- [ ] **Step 11: System section — disable restart and backup buttons**

Find the restart button:
```html
<button class="btn btn-outline-danger" @click="restart">
    {{ $t("config.system.restart") }}
</button>
```

Replace with:
```html
<button
    class="btn btn-outline-danger"
    :disabled="hubMode"
    :title="hubMode ? $t('hub.cloudNotAvailable.restartTooltip') : undefined"
    @click="restart"
>
    {{ $t("config.system.restart") }}
</button>
```

Find the backup/restore button:
```html
<button
    data-testid="backup-restore"
    class="btn btn-outline-secondary text-truncate"
    @click="openModal('backuprestore')"
>
    {{ $t("config.system.backupRestore.title") }}
</button>
```

Replace with:
```html
<button
    data-testid="backup-restore"
    class="btn btn-outline-secondary text-truncate"
    :disabled="hubMode"
    :title="hubMode ? $t('hub.cloudNotAvailable.backupTooltip') : undefined"
    @click="openModal('backuprestore')"
>
    {{ $t("config.system.backupRestore.title") }}
</button>
```

- [ ] **Step 12: Run lint and type check**

```bash
cd web && npm run lint:tsc
```
Expected: no errors.

- [ ] **Step 13: Run full test suite**

```bash
cd web && npm test
```
Expected: all tests pass.

- [ ] **Step 14: Commit**

```bash
git add web/assets/js/views/Config.vue
git commit -m "feat: hub mode read-only for Config.vue — MQTT sections read-only, REST sections hidden"
```

---

## Task 7: Final validation and push

- [ ] **Step 1: Run full lint suite**

```bash
cd web && npm run lint
```
Expected: no errors, no warnings.

- [ ] **Step 2: Run all tests**

```bash
cd web && npm test
```
Expected: all tests pass.

- [ ] **Step 3: Push to GitHub**

```bash
git push origin main
```

---

## Reference: Section-by-section Config.vue change summary

| Section | h2 i18n key | Change |
|---|---|---|
| General | `config.section.general` | `GeneralConfig` replaced with `HubModeNotice` |
| Loadpoints | `config.section.loadpoints` | DeviceCards `:editable="!hubMode && !!loadpoint.id"`, `NewDeviceButton` hidden |
| Vehicles | `config.section.vehicles` | DeviceCards `:editable="!hubMode && vehicle.id >= 0"`, `NewDeviceButton` hidden |
| Grid | `config.section.grid` | `config-list` div replaced with `HubModeNotice` |
| PV/Battery | `config.section.meter` | `config-list` div replaced with `HubModeNotice` |
| Additional | `config.section.additionalMeter` | `config-list` div replaced with `HubModeNotice` |
| Tariffs | `config.tariff.title` | Both tariff divs wrapped in `v-if`, replaced with `HubModeNotice` |
| Integrations | `config.section.integrations` | Entire `config-list` (incl. `AuthProvidersCard`) replaced with `HubModeNotice` |
| Services | `config.section.services` | `config-list` div replaced with `HubModeNotice` |
| System (restart) | — | Button `disabled` + `title` tooltip |
| System (backup) | — | Button `disabled` + `title` tooltip |
