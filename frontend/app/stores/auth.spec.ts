import { HttpResponse, http } from 'msw'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { ResponseError, TokenPairToJSON, UserToJSON } from '~~/api'
import { errorResponse } from '../../test/mocks/handlers'
import { server } from '../../test/mocks/server'
import { buildAdminTokenPair, buildTokenPair, buildUser } from '../../test/mocks/factories'
import { useAuthStore } from './auth'

const BASE = 'http://localhost:8085'

// The "nuxt" vitest environment boots one Nuxt app (and thus one Pinia
// instance) per test file, so creating a fresh Pinia per test would leave
// it disconnected from the store instance auto-imports elsewhere in the app
// resolve to. Resetting the existing store's state is what actually starts
// each test clean - $reset() re-runs the store's own state() factory,
// including `initialized`, which the "init" tests below depend on.
beforeEach(() => {
  useAuthStore().$reset()
})

describe('useAuthStore', () => {
  describe('login', () => {
    it('stores the token pair and loads the current user', async () => {
      const tokens = buildTokenPair()
      server.use(
        http.post(`${BASE}/v1/auth/login`, () => HttpResponse.json(TokenPairToJSON(tokens))),
        http.get(`${BASE}/v1/users/me`, () => HttpResponse.json(UserToJSON(buildUser({ id: 'user-1' }))))
      )
      const store = useAuthStore()

      await store.login('jane@example.com', 'password')

      expect(store.accessToken).toBe(tokens.accessToken)
      expect(store.refreshToken).toBe(tokens.refreshToken)
      expect(store.user?.id).toBe('user-1')
      expect(localStorage.getItem('linkshelf.accessToken')).toBe(tokens.accessToken)
    })

    it('propagates a login failure and does not persist tokens', async () => {
      server.use(http.post(`${BASE}/v1/auth/login`, () => errorResponse(401, 'invalid credentials')))
      const store = useAuthStore()

      await expect(store.login('jane@example.com', 'wrong')).rejects.toBeInstanceOf(ResponseError)
      expect(store.accessToken).toBeNull()
      expect(localStorage.getItem('linkshelf.accessToken')).toBeNull()
    })
  })

  describe('register', () => {
    it('creates the account and logs in when email verification is not required', async () => {
      server.use(http.post(`${BASE}/v1/users`, () => HttpResponse.json(UserToJSON(buildUser()))))
      const store = useAuthStore()

      const result = await store.register({ email: 'new@example.com', username: 'new-user', firstName: 'New', lastName: 'User', password: 'secret123' })

      expect(result).toEqual({ pendingVerification: false })
      expect(store.isAuthenticated).toBe(true)
    })

    it('reports pendingVerification when the follow-up login 403s, without throwing', async () => {
      server.use(
        http.post(`${BASE}/v1/users`, () => HttpResponse.json(UserToJSON(buildUser()))),
        http.post(`${BASE}/v1/auth/login`, () => errorResponse(403, 'email not verified'))
      )
      const store = useAuthStore()

      const result = await store.register({ email: 'new@example.com', username: 'new-user', firstName: 'New', lastName: 'User', password: 'secret123' })

      expect(result).toEqual({ pendingVerification: true })
      expect(store.isAuthenticated).toBe(false)
    })

    it('rethrows a non-403 failure from the follow-up login', async () => {
      server.use(
        http.post(`${BASE}/v1/users`, () => HttpResponse.json(UserToJSON(buildUser()))),
        http.post(`${BASE}/v1/auth/login`, () => errorResponse(500, 'boom'))
      )
      const store = useAuthStore()

      await expect(store.register({ email: 'new@example.com', username: 'new-user', firstName: 'New', lastName: 'User', password: 'secret123' })).rejects.toBeInstanceOf(ResponseError)
    })
  })

  describe('fetchUser', () => {
    it('is a no-op without an access token', async () => {
      const store = useAuthStore()
      await store.fetchUser()
      expect(store.user).toBeNull()
    })
  })

  describe('refresh', () => {
    it('returns false and clears the session when there is no refresh token', async () => {
      const store = useAuthStore()
      const result = await store.refresh()
      expect(result).toBe(false)
    })

    it('rotates the token pair on success', async () => {
      const store = useAuthStore()
      store.setTokens(buildTokenPair({ refreshToken: 'old-refresh' }))
      const rotated = buildTokenPair()
      server.use(http.post(`${BASE}/v1/auth/refresh`, () => HttpResponse.json(TokenPairToJSON(rotated))))

      const result = await store.refresh()

      expect(result).toBe(true)
      expect(store.accessToken).toBe(rotated.accessToken)
      expect(store.refreshToken).toBe(rotated.refreshToken)
    })

    it('shares one request between concurrent refreshes', async () => {
      const store = useAuthStore()
      store.setTokens(buildTokenPair({ refreshToken: 'old-refresh' }))
      const rotated = buildTokenPair()
      let calls = 0
      server.use(http.post(`${BASE}/v1/auth/refresh`, async () => {
        calls += 1
        await new Promise(resolve => setTimeout(resolve, 20))
        return HttpResponse.json(TokenPairToJSON(rotated))
      }))

      const results = await Promise.all([store.refresh(), store.refresh(), store.refresh()])

      expect(results).toEqual([true, true, true])
      expect(calls).toBe(1)
      expect(store.refreshToken).toBe(rotated.refreshToken)
    })

    describe('when the refresh token is rejected', () => {
      afterEach(async () => {
        await navigateTo('/')
      })

      it('sends the user from an app page to sign-in and remembers where they were', async () => {
        const store = useAuthStore()
        store.setTokens(buildTokenPair({ refreshToken: 'old-refresh' }))
        await navigateTo('/app/themes?sort=name')
        server.use(http.post(`${BASE}/v1/auth/refresh`, () => errorResponse(401, 'expired')))

        await store.refresh()

        await vi.waitFor(() => {
          const route = useRouter().currentRoute.value
          expect(route.path).toBe('/auth/sign-in')
          expect(route.query.redirect).toBe('/app/themes?sort=name')
        })
      })

      it('leaves the user alone on a public page', async () => {
        const store = useAuthStore()
        store.setTokens(buildTokenPair({ refreshToken: 'old-refresh' }))
        await navigateTo('/docs')
        server.use(http.post(`${BASE}/v1/auth/refresh`, () => errorResponse(401, 'expired')))

        await store.refresh()
        await new Promise(resolve => setTimeout(resolve, 50))

        expect(useRouter().currentRoute.value.path).toBe('/docs')
      })
    })

    it('clears the session when the refresh token is rejected', async () => {
      const store = useAuthStore()
      store.setTokens(buildTokenPair())
      server.use(http.post(`${BASE}/v1/auth/refresh`, () => errorResponse(401, 'expired')))

      const result = await store.refresh()

      expect(result).toBe(false)
      expect(store.accessToken).toBeNull()
      expect(store.refreshToken).toBeNull()
      expect(localStorage.getItem('linkshelf.accessToken')).toBeNull()
    })
  })

  describe('logout', () => {
    it('invalidates the refresh token on the server and clears local state', async () => {
      const store = useAuthStore()
      store.setTokens(buildTokenPair())
      let called = false
      server.use(http.post(`${BASE}/v1/auth/logout`, () => {
        called = true
        return new HttpResponse(null, { status: 204 })
      }))

      await store.logout()

      expect(called).toBe(true)
      expect(store.accessToken).toBeNull()
      expect(store.isAuthenticated).toBe(false)
    })

    it('still clears local state when the server call fails', async () => {
      const store = useAuthStore()
      store.setTokens(buildTokenPair())
      server.use(http.post(`${BASE}/v1/auth/logout`, () => errorResponse(500, 'boom')))

      await store.logout()

      expect(store.accessToken).toBeNull()
    })

    it('skips the server call when there is no refresh token', async () => {
      const store = useAuthStore()
      await expect(store.logout()).resolves.toBeUndefined()
    })
  })

  describe('OIDC', () => {
    it('startOidcLogin stashes the state and redirects the browser', async () => {
      const store = useAuthStore()
      server.use(http.get(`${BASE}/v1/auth/oidc/login`, () => HttpResponse.json({
        authorization_url: 'https://idp.example.com/authorize',
        state: 'my-state'
      })))
      const originalLocation = window.location
      // @ts-expect-error - deleting window.location to allow reassignment below
      delete window.location
      // @ts-expect-error - a minimal stub is enough for this redirect assertion
      window.location = { href: '' }

      await store.startOidcLogin()

      expect(sessionStorage.getItem('linkshelf.oidcState')).toBe('my-state')
      expect(window.location.href).toBe('https://idp.example.com/authorize')
      // @ts-expect-error - restoring the real window.location after the test
      window.location = originalLocation
    })

    it('completeOidcLogin stores the returned tokens and loads the user', async () => {
      const store = useAuthStore()
      const tokens = buildTokenPair()
      server.use(
        http.post(`${BASE}/v1/auth/oidc/callback`, () => HttpResponse.json(TokenPairToJSON(tokens))),
        http.get(`${BASE}/v1/users/me`, () => HttpResponse.json(UserToJSON(buildUser({ id: 'user-1' }))))
      )

      await store.completeOidcLogin('the-code', 'the-state')

      expect(store.accessToken).toBe(tokens.accessToken)
      expect(store.user?.id).toBe('user-1')
    })

    it('consumeStoredOidcState reads and clears the stashed state', () => {
      sessionStorage.setItem('linkshelf.oidcState', 'stashed')
      const store = useAuthStore()

      expect(store.consumeStoredOidcState()).toBe('stashed')
      expect(store.consumeStoredOidcState()).toBeNull()
    })
  })

  describe('claims-derived getters', () => {
    it('reports authenticated state, role and admin status from the access token', () => {
      const store = useAuthStore()
      store.setTokens(buildAdminTokenPair())

      expect(store.isAuthenticated).toBe(true)
      expect(store.role).toBe('admin')
      expect(store.isAdmin).toBe(true)
      expect(store.userId).toBe('user-1')
    })

    it('reports unauthenticated defaults with no token', () => {
      const store = useAuthStore()

      expect(store.isAuthenticated).toBe(false)
      expect(store.claims).toBeNull()
      expect(store.userId).toBeNull()
      expect(store.role).toBeNull()
      expect(store.isAdmin).toBe(false)
    })
  })

  describe('init', () => {
    it('restores tokens previously saved to localStorage', () => {
      localStorage.setItem('linkshelf.accessToken', 'stored-access')
      localStorage.setItem('linkshelf.refreshToken', 'stored-refresh')
      const store = useAuthStore()

      store.init()

      expect(store.accessToken).toBe('stored-access')
      expect(store.refreshToken).toBe('stored-refresh')
    })

    it('only restores once', () => {
      localStorage.setItem('linkshelf.accessToken', 'stored-access')
      const store = useAuthStore()

      store.init()
      localStorage.setItem('linkshelf.accessToken', 'changed-after-init')
      store.init()

      expect(store.accessToken).toBe('stored-access')
    })
  })

  describe('clear', () => {
    it('removes tokens, user and localStorage entries', () => {
      const store = useAuthStore()
      store.setTokens(buildTokenPair())
      store.user = buildUser()

      store.clear()

      expect(store.accessToken).toBeNull()
      expect(store.refreshToken).toBeNull()
      expect(store.user).toBeNull()
      expect(localStorage.getItem('linkshelf.accessToken')).toBeNull()
    })
  })
})
