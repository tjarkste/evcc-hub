// web/assets/js/services/stateCache.ts
import type { State } from '../types/evcc'

export const CACHE_KEY = 'evcc-cloud-state-cache'

const EXCLUDED_KEYS = ['offline', 'connectionState', 'lastDataAt']

let saveTimer: ReturnType<typeof setTimeout> | null = null

/**
 * Debounced save — writes at most once per 5 seconds to avoid
 * performance issues from frequent MQTT messages.
 * Uses a trailing debounce: resets the timer on each call so only the
 * last update within a 5-second quiet window is persisted.
 */
export function saveStateCache(state: State): void {
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = setTimeout(() => {
    saveTimer = null
    try {
      const filtered: Record<string, unknown> = {}
      for (const [key, value] of Object.entries(state)) {
        if (!EXCLUDED_KEYS.includes(key) && value !== undefined) {
          filtered[key] = value
        }
      }
      localStorage.setItem(CACHE_KEY, JSON.stringify(filtered))
    } catch {
      // localStorage full or unavailable — silently ignore
    }
  }, 5000)
}

export function loadStateCache(): Partial<State> | null {
  try {
    const raw = localStorage.getItem(CACHE_KEY)
    if (!raw) return null
    return JSON.parse(raw)
  } catch {
    return null
  }
}

export function clearStateCache(): void {
  localStorage.removeItem(CACHE_KEY)
}
