// test-setup.ts — global test setup for vitest
// Node.js v22+ exposes a built-in localStorage stub that requires --localstorage-file.
// We replace it with a simple in-memory implementation for all tests.

class InMemoryStorage implements Storage {
  private _data: Record<string, string> = {}

  get length(): number {
    return Object.keys(this._data).length
  }

  clear(): void {
    this._data = {}
  }

  getItem(key: string): string | null {
    return Object.prototype.hasOwnProperty.call(this._data, key)
      ? this._data[key]
      : null
  }

  key(index: number): string | null {
    const keys = Object.keys(this._data)
    return keys[index] ?? null
  }

  removeItem(key: string): void {
    delete this._data[key]
  }

  setItem(key: string, value: string): void {
    this._data[key] = String(value)
  }
}

Object.defineProperty(globalThis, 'localStorage', {
  value: new InMemoryStorage(),
  writable: true,
  configurable: true,
})

Object.defineProperty(globalThis, 'sessionStorage', {
  value: new InMemoryStorage(),
  writable: true,
  configurable: true,
})
