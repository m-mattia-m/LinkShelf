import { HttpResponse, http } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'
import { UserToJSON } from '~~/api'
import { server } from '../../test/mocks/server'
import { buildTokenPair, buildUser } from '../../test/mocks/factories'
import { useUserStore } from './user'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  useUserStore().$reset()
  useAuthStore().$reset()
})

describe('useUserStore', () => {
  describe('fetch', () => {
    it('loads the list of users', async () => {
      const users = [buildUser(), buildUser()]
      server.use(http.get(`${BASE}/v1/users`, () => HttpResponse.json(users.map(UserToJSON))))
      const store = useUserStore()

      await store.fetch()

      expect(store.users).toHaveLength(2)
      expect(store.loaded).toBe(true)
    })
  })

  describe('create', () => {
    it('sends the caller\'s bearer token and refetches the list', async () => {
      const authStore = useAuthStore()
      authStore.setTokens(buildTokenPair({ accessToken: 'admin-token' }))
      const store = useUserStore()
      const created = buildUser({ id: 'user-new' })
      server.use(
        http.post(`${BASE}/v1/users`, ({ request }) => {
          expect(request.headers.get('Authorization')).toBe('Bearer admin-token')
          return HttpResponse.json(UserToJSON(created))
        }),
        http.get(`${BASE}/v1/users`, () => HttpResponse.json([UserToJSON(created)]))
      )

      const result = await store.create({ email: 'new@example.com', username: 'new-user', firstName: 'New', lastName: 'User', password: 'secret123' })

      expect(result.id).toBe('user-new')
      expect(store.users).toHaveLength(1)
    })

    it('omits the Authorization header when there is no access token', async () => {
      const store = useUserStore()
      server.use(
        http.post(`${BASE}/v1/users`, ({ request }) => {
          expect(request.headers.has('Authorization')).toBe(false)
          return HttpResponse.json(UserToJSON(buildUser()))
        }),
        http.get(`${BASE}/v1/users`, () => HttpResponse.json([]))
      )

      await store.create({ email: 'new@example.com', username: 'new-user', firstName: 'New', lastName: 'User', password: 'secret123' })
    })
  })

  describe('update', () => {
    it('updates a user and refetches the list', async () => {
      const store = useUserStore()
      const updated = buildUser({ id: 'user-1', firstName: 'Updated' })
      server.use(
        http.put(`${BASE}/v1/users/:userId`, ({ params }) => {
          expect(params.userId).toBe('user-1')
          return HttpResponse.json(UserToJSON(updated))
        }),
        http.get(`${BASE}/v1/users`, () => HttpResponse.json([UserToJSON(updated)]))
      )

      const result = await store.update('user-1', { email: updated.email, username: updated.username, firstName: 'Updated', lastName: 'Doe' })

      expect(result.firstName).toBe('Updated')
      expect(store.users).toHaveLength(1)
    })
  })

  describe('patchPassword', () => {
    it('sends the password change without refetching the list', async () => {
      const store = useUserStore()
      let calls = 0
      server.use(
        http.patch(`${BASE}/v1/users/:userId/password`, ({ params }) => {
          calls += 1
          expect(params.userId).toBe('user-1')
          return new HttpResponse(null, { status: 204 })
        }),
        http.get(`${BASE}/v1/users`, () => {
          throw new Error('should not refetch')
        })
      )

      await store.patchPassword('user-1', { oldPassword: 'old', newPassword: 'new-secret' })

      expect(calls).toBe(1)
      expect(store.users).toEqual([])
    })
  })

  describe('remove', () => {
    it('deletes a user and refetches the list', async () => {
      const store = useUserStore()
      let deleteCalled = false
      server.use(
        http.delete(`${BASE}/v1/users/:userId`, ({ params }) => {
          deleteCalled = true
          expect(params.userId).toBe('user-1')
          return new HttpResponse(null, { status: 204 })
        }),
        http.get(`${BASE}/v1/users`, () => HttpResponse.json([]))
      )

      await store.remove('user-1')

      expect(deleteCalled).toBe(true)
      expect(store.loaded).toBe(true)
    })
  })

  describe('resendVerification', () => {
    it('forwards the email to the resend-verification endpoint without refetching', async () => {
      const store = useUserStore()
      let receivedEmail: string | undefined
      server.use(
        http.post(`${BASE}/v1/auth/resend-verification`, async ({ request }) => {
          const body = await request.json() as { email: string }
          receivedEmail = body.email
          return new HttpResponse(null, { status: 204 })
        }),
        http.get(`${BASE}/v1/users`, () => {
          throw new Error('should not refetch')
        })
      )

      await store.resendVerification('someone@example.com')

      expect(receivedEmail).toBe('someone@example.com')
    })
  })

  describe('markVerified', () => {
    it('marks a user verified and refetches the list', async () => {
      const store = useUserStore()
      const verified = buildUser({ id: 'user-1', emailVerified: true })
      server.use(
        http.patch(`${BASE}/v1/users/:userId/verify`, ({ params }) => {
          expect(params.userId).toBe('user-1')
          return HttpResponse.json(UserToJSON(verified))
        }),
        http.get(`${BASE}/v1/users`, () => HttpResponse.json([UserToJSON(verified)]))
      )

      await store.markVerified('user-1')

      expect(store.users).toHaveLength(1)
      expect(store.users[0]?.emailVerified).toBe(true)
    })
  })
})
