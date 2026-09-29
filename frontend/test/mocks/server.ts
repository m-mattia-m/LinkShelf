import { setupServer } from 'msw/node'
import { handlers } from './handlers'

// The "nuxt" environment evaluates setupFiles and test files in separate
// module graphs, so the server is cached on globalThis to keep one instance.
const KEY = Symbol.for('linkshelf.test.mswServer')

type GlobalWithMswServer = typeof globalThis & {
  [KEY]?: ReturnType<typeof setupServer>
}

const globalWithMswServer = globalThis as GlobalWithMswServer

export const server = globalWithMswServer[KEY] ?? (globalWithMswServer[KEY] = setupServer(...handlers))
