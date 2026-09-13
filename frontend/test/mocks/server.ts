import { setupServer } from 'msw/node'
import { handlers } from './handlers'

// @nuxt/test-utils' "nuxt" vitest environment evaluates setupFiles and test
// files through two separate module graphs, so a plain module-level
// `setupServer(...)` would produce two independent instances: one listening
// (created via setupFiles) and one that test files call `.use()` on but that
// never actually intercepts anything. Caching the instance on `globalThis`
// (shared across both graphs, since they run in the same process) makes
// every import - regardless of which graph evaluated the module - resolve
// to the one server that's actually listening.
const KEY = Symbol.for('linkshelf.test.mswServer')

type GlobalWithMswServer = typeof globalThis & {
  [KEY]?: ReturnType<typeof setupServer>
}

const globalWithMswServer = globalThis as GlobalWithMswServer

export const server = globalWithMswServer[KEY] ?? (globalWithMswServer[KEY] = setupServer(...handlers))
