import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it } from 'vitest'
import { buildTokenPair } from '../../test/mocks/factories'
import authGlobalMiddleware from './auth.global'

// See admin.spec.ts for why navigateTo is mocked to a plain passthrough.
mockNuxtImport('navigateTo', () => (to: unknown) => to)

beforeEach(() => {
  useAuthStore().$reset()
})

function route(path: string) {
  return { path, fullPath: path } as never
}

describe('auth.global middleware', () => {
  it('redirects an unauthenticated visitor away from an /app route, preserving the target as a redirect query', () => {
    const result = authGlobalMiddleware(route('/app/shelves'), {} as never)

    expect(result).toEqual({
      path: '/auth/sign-in',
      query: { redirect: '/app/shelves' }
    })
  })

  it('lets an authenticated visitor through to an /app route', () => {
    const authStore = useAuthStore()
    authStore.setTokens(buildTokenPair())

    const result = authGlobalMiddleware(route('/app/shelves'), {} as never)

    expect(result).toBeUndefined()
  })

  it('redirects an already-authenticated visitor away from an /auth route to /app', () => {
    const authStore = useAuthStore()
    authStore.setTokens(buildTokenPair())

    const result = authGlobalMiddleware(route('/auth/sign-in'), {} as never)

    expect(result).toEqual('/app')
  })

  it('lets an unauthenticated visitor reach an /auth route', () => {
    const result = authGlobalMiddleware(route('/auth/sign-in'), {} as never)

    expect(result).toBeUndefined()
  })

  it.each([
    '/auth/callback',
    '/auth/verify-email',
    '/auth/set-password',
    '/auth/reset-password'
  ])('always lets a token-action route (%s) through even when already authenticated', (path) => {
    const authStore = useAuthStore()
    authStore.setTokens(buildTokenPair())

    const result = authGlobalMiddleware(route(path), {} as never)

    expect(result).toBeUndefined()
  })

  it('leaves routes outside /app and /auth untouched', () => {
    const result = authGlobalMiddleware(route('/'), {} as never)

    expect(result).toBeUndefined()
  })

  it('initializes the auth store from localStorage before checking', () => {
    localStorage.setItem('linkshelf.accessToken', buildTokenPair().accessToken)
    localStorage.setItem('linkshelf.refreshToken', 'stored-refresh')

    const result = authGlobalMiddleware(route('/app/shelves'), {} as never)

    expect(result).toBeUndefined()
  })
})
