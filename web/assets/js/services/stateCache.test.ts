// web/assets/js/services/stateCache.test.ts
import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest'
import { saveStateCache, loadStateCache, clearStateCache, CACHE_KEY } from './stateCache'

beforeEach(() => {
  localStorage.clear()
  vi.useFakeTimers()
})

afterEach(() => {
  vi.useRealTimers()
})

describe('stateCache', () => {
  test('saves and loads state after debounce', () => {
    const state = { loadpoints: [{ mode: 'pv', chargePower: 3600 }], vehicles: {} }
    saveStateCache(state as any)
    vi.advanceTimersByTime(5000)

    const loaded = loadStateCache()
    expect(loaded).not.toBeNull()
    expect(loaded!.loadpoints[0].mode).toBe('pv')
    expect(loaded!.loadpoints[0].chargePower).toBe(3600)
  })

  test('does not write before debounce interval', () => {
    saveStateCache({ loadpoints: [], vehicles: {} } as any)
    vi.advanceTimersByTime(1000)
    expect(loadStateCache()).toBeNull()
  })

  test('returns null when no cache exists', () => {
    expect(loadStateCache()).toBeNull()
  })

  test('returns null for corrupted data', () => {
    localStorage.setItem(CACHE_KEY, 'not-json{{{')
    expect(loadStateCache()).toBeNull()
  })

  test('clearStateCache removes the cache', () => {
    saveStateCache({ loadpoints: [], vehicles: {} } as any)
    vi.advanceTimersByTime(5000)
    clearStateCache()
    expect(loadStateCache()).toBeNull()
  })

  test('does not save connectionState or offline fields', () => {
    const state = {
      offline: true,
      connectionState: 'reconnecting',
      lastDataAt: 12345,
      loadpoints: [{ mode: 'pv' }],
      vehicles: {},
    }
    saveStateCache(state as any)
    vi.advanceTimersByTime(5000)
    const loaded = loadStateCache()
    expect(loaded).not.toHaveProperty('offline')
    expect(loaded).not.toHaveProperty('connectionState')
    expect(loaded).not.toHaveProperty('lastDataAt')
    expect(loaded!.loadpoints[0].mode).toBe('pv')
  })
})
