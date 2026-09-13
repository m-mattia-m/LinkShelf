import { HttpResponse, http } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'
import { UserToJSON } from '~~/api'
import { server } from '../../test/mocks/server'
import { buildTokenPair, buildUser } from '../../test/mocks/factories'
import { useCurrentUser } from './useCurrentUser'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  useAuthStore().$reset()
})

describe('useCurrentUser', () => {
  describe('userId / user', () => {
    it('reflects the auth store\'s claims-derived userId and loaded user', () => {
      const authStore = useAuthStore()
      authStore.setTokens(buildTokenPair())
      authStore.user = buildUser({ id: 'user-1' })
      const { userId, user } = useCurrentUser()

      expect(userId.value).toBe('user-1')
      expect(user.value?.id).toBe('user-1')
    })

    it('is null when unauthenticated', () => {
      const { userId, user } = useCurrentUser()

      expect(userId.value).toBeNull()
      expect(user.value).toBeNull()
    })
  })

  describe('ensureUserId', () => {
    it('returns the userId once tokens are present', async () => {
      const authStore = useAuthStore()
      authStore.setTokens(buildTokenPair())
      const { ensureUserId } = useCurrentUser()

      await expect(ensureUserId()).resolves.toBe('user-1')
    })

    it('throws when there is no authenticated user', async () => {
      const { ensureUserId } = useCurrentUser()

      await expect(ensureUserId()).rejects.toThrow('Not authenticated')
    })
  })

  describe('ensureUser', () => {
    it('returns the already-loaded user without fetching', async () => {
      const authStore = useAuthStore()
      authStore.setTokens(buildTokenPair())
      authStore.user = buildUser({ id: 'user-1', firstName: 'Cached' })
      let calls = 0
      server.use(http.get(`${BASE}/v1/users/me`, () => {
        calls += 1
        return HttpResponse.json(UserToJSON(buildUser({ id: 'user-1' })))
      }))
      const { ensureUser } = useCurrentUser()

      const result = await ensureUser()

      expect(result.firstName).toBe('Cached')
      expect(calls).toBe(0)
    })

    it('fetches the user when authenticated but not yet loaded', async () => {
      const authStore = useAuthStore()
      authStore.setTokens(buildTokenPair())
      server.use(http.get(`${BASE}/v1/users/me`, () => HttpResponse.json(UserToJSON(buildUser({ id: 'user-1', firstName: 'Fetched' })))))
      const { ensureUser } = useCurrentUser()

      const result = await ensureUser()

      expect(result.firstName).toBe('Fetched')
    })

    it('throws when there is no access token to fetch with', async () => {
      const { ensureUser } = useCurrentUser()

      await expect(ensureUser()).rejects.toThrow('Not authenticated')
    })
  })
})
