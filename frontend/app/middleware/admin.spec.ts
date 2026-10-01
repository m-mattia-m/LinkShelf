import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it } from 'vitest'
import { buildAdminTokenPair, buildTokenPair } from '../../test/mocks/factories'
import adminMiddleware from './admin'

// Passthrough navigateTo so the redirect is deterministic.
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
