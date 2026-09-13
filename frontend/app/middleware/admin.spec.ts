import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it } from 'vitest'
import { buildAdminTokenPair, buildTokenPair } from '../../test/mocks/factories'
import adminMiddleware from './admin'

// navigateTo() normally performs a real router push, whose return value
// (Promise vs. route location) depends on whether Nuxt's router considers
// itself "inside" middleware processing at the moment it's called - state
// this test file has no control over when the middleware is invoked
// directly instead of through the router pipeline. Mocking it to a plain
// passthrough makes the middleware's return value (what it actually hands
// back to the router) deterministic and directly assertable, matching how
// route middleware communicates a redirect.
mockNuxtImport('navigateTo', () => (to: unknown) => to)

beforeEach(() => {
  useAuthStore().$reset()
})

describe('admin middleware', () => {
  it('redirects a non-admin (or unauthenticated) user to /app', () => {
    const result = adminMiddleware({ path: '/app/admin', fullPath: '/app/admin' } as never, {} as never)

    expect(result).toBe('/app')
  })

  it('redirects a regular authenticated user to /app', () => {
    const authStore = useAuthStore()
    authStore.setTokens(buildTokenPair())

    const result = adminMiddleware({ path: '/app/admin', fullPath: '/app/admin' } as never, {} as never)

    expect(result).toBe('/app')
  })

  it('lets an admin through', () => {
    const authStore = useAuthStore()
    authStore.setTokens(buildAdminTokenPair())

    const result = adminMiddleware({ path: '/app/admin', fullPath: '/app/admin' } as never, {} as never)

    expect(result).toBeUndefined()
  })

  it('initializes the auth store from localStorage before checking', () => {
    localStorage.setItem('linkshelf.accessToken', buildAdminTokenPair().accessToken)

    const result = adminMiddleware({ path: '/app/admin', fullPath: '/app/admin' } as never, {} as never)

    expect(result).toBeUndefined()
  })
})
