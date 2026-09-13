import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/vue'
import { afterAll, afterEach, beforeAll } from 'vitest'
import { server } from './mocks/server'
import { resetFactoryCounter } from './mocks/factories'

// Node's built-in `localStorage`/`sessionStorage` (enabled by @nuxt/test-utils'
// "nuxt" vitest environment) is only functional when Node is started with
// `--localstorage-file=<path>` - without it, the objects exist but every
// method throws. Replace both with a plain in-memory implementation so
// app code (e.g. the auth store) can use them like a real browser would.
class MemoryStorage implements Storage {
  private store = new Map<string, string>()

  get length(): number {
    return this.store.size
  }

  clear(): void {
    this.store.clear()
  }

  getItem(key: string): string | null {
    return this.store.has(key) ? this.store.get(key)! : null
  }

  key(index: number): string | null {
    return Array.from(this.store.keys())[index] ?? null
  }

  removeItem(key: string): void {
    this.store.delete(key)
  }

  setItem(key: string, value: string): void {
    this.store.set(key, String(value))
  }
}

globalThis.localStorage = new MemoryStorage()
globalThis.sessionStorage = new MemoryStorage()

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))

afterEach(() => {
  server.resetHandlers()
  resetFactoryCounter()
  localStorage.clear()
  sessionStorage.clear()
  cleanup()
})

afterAll(() => server.close())
