import { HttpResponse, http } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'
import { ResponseError, TokenPairToJSON, UserToJSON } from '~~/api'
import { errorResponse } from '../../test/mocks/handlers'
import { server } from '../../test/mocks/server'
import { buildTokenPair, buildUser } from '../../test/mocks/factories'
import { useApi } from './useApi'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  useAuthStore().$reset()
})

describe('useApi', () => {
  it('retries once with a refreshed token after a 401 on an authenticated request', async () => {
    const authStore = useAuthStore()
    authStore.setTokens(buildTokenPair({ accessToken: 'stale-token', refreshToken: 'refresh-1' }))
    const rotated = buildTokenPair({ accessToken: 'fresh-token' })

    let calls = 0
    server.use(
      http.get(`${BASE}/v1/shelves`, ({ request }) => {
        calls += 1
        if (calls === 1) {
          expect(request.headers.get('Authorization')).toBe('Bearer stale-token')
          return errorResponse(401, 'expired')
        }
        expect(request.headers.get('Authorization')).toBe('Bearer fresh-token')
        return HttpResponse.json([])
      }),
      http.post(`${BASE}/v1/auth/refresh`, () => HttpResponse.json(TokenPairToJSON(rotated)))
    )

    const api = useApi()
    const result = await api.shelf.listShelves()

    expect(result).toEqual([])
    expect(calls).toBe(2)
    expect(authStore.accessToken).toBe('fresh-token')
  })

  it('sends exactly one Authorization header on the retry', async () => {
    // Real browsers (and Node) merge two differently-cased "Authorization"
    // entries into "Bearer <old>, Bearer <new>", which the backend rejects.
    // The test environment's fetch silently keeps only the last one, so the
    // request is inspected as it is handed to fetch instead of as received.
    const authStore = useAuthStore()
    authStore.setTokens(buildTokenPair({ accessToken: 'stale-token', refreshToken: 'refresh-1' }))
    server.use(
      http.get(`${BASE}/v1/shelves`, ({ request }) => request.headers.get('Authorization')?.includes('fresh-token') ? HttpResponse.json([]) : errorResponse(401, 'expired')),
      http.post(`${BASE}/v1/auth/refresh`, () => HttpResponse.json(TokenPairToJSON(buildTokenPair({ accessToken: 'fresh-token' }))))
    )
    const realFetch = globalThis.fetch
    const sentHeaders: HeadersInit[] = []
    vi.spyOn(globalThis, 'fetch').mockImplementation((input, init) => {
      sentHeaders.push(init?.headers ?? {})
      return realFetch(input, init)
    })

    await useApi().shelf.listShelves()

    const retryHeaders = sentHeaders.at(-1)!
    const names = retryHeaders instanceof Headers
      ? [...retryHeaders.keys()]
      : Array.isArray(retryHeaders) ? retryHeaders.map(([name]) => name) : Object.keys(retryHeaders)
    expect(names.filter(name => name.toLowerCase() === 'authorization')).toHaveLength(1)
    vi.restoreAllMocks()
  })

  it('refreshes once when several requests hit an expired token at the same time', async () => {
    const authStore = useAuthStore()
    authStore.setTokens(buildTokenPair({ accessToken: 'stale-token', refreshToken: 'refresh-1' }))

    // The real backend rotates refresh tokens: each one works exactly once.
    const validRefreshTokens = new Set(['refresh-1'])
    let refreshCalls = 0
    const isFresh = (request: Request) => request.headers.get('Authorization') === 'Bearer fresh-token'
    server.use(
      http.get(`${BASE}/v1/shelves`, ({ request }) => isFresh(request) ? HttpResponse.json([]) : errorResponse(401, 'expired')),
      http.get(`${BASE}/v1/users/me`, ({ request }) => isFresh(request) ? HttpResponse.json(UserToJSON(buildUser())) : errorResponse(401, 'expired')),
      http.post(`${BASE}/v1/auth/refresh`, async ({ request }) => {
        refreshCalls += 1
        const { refresh_token: refreshToken } = await request.json() as { refresh_token: string }
        if (!validRefreshTokens.delete(refreshToken)) return errorResponse(401, 'refresh token invalid')
        validRefreshTokens.add('refresh-2')
        return HttpResponse.json(TokenPairToJSON(buildTokenPair({ accessToken: 'fresh-token', refreshToken: 'refresh-2' })))
      })
    )

    // A page load fires several requests at once, all with the same expired token.
    const api = useApi()
    const results = await Promise.allSettled([api.shelf.listShelves(), api.shelf.listShelves(), api.user.getCurrentUser()])

    expect(results.map(r => r.status)).toEqual(['fulfilled', 'fulfilled', 'fulfilled'])
    expect(refreshCalls).toBe(1)
    expect(authStore.refreshToken).toBe('refresh-2')
  })

  it('does not retry a 401 on a request with no Authorization header', async () => {
    const authStore = useAuthStore()
    authStore.setTokens(buildTokenPair())
    let calls = 0
    server.use(http.get(`${BASE}/v1/shelves/by-path/:path`, () => {
      calls += 1
      return errorResponse(401, 'unauthenticated')
    }))

    const api = useApi()

    await expect(api.shelf.getPublicShelfByPath({ path: 'foo' })).rejects.toBeInstanceOf(ResponseError)
    expect(calls).toBe(1)
  })

  it('does not retry when there is no refresh token to use', async () => {
    const authStore = useAuthStore()
    authStore.setTokens(buildTokenPair({ refreshToken: '' as unknown as string }))
    let calls = 0
    server.use(http.get(`${BASE}/v1/shelves`, () => {
      calls += 1
      return errorResponse(401, 'expired')
    }))

    const api = useApi()

    await expect(api.shelf.listShelves()).rejects.toBeInstanceOf(ResponseError)
    expect(calls).toBe(1)
  })

  it('propagates the original 401 when the refresh attempt itself fails', async () => {
    const authStore = useAuthStore()
    authStore.setTokens(buildTokenPair({ accessToken: 'stale-token', refreshToken: 'refresh-1' }))
    server.use(
      http.get(`${BASE}/v1/shelves`, () => errorResponse(401, 'expired')),
      http.post(`${BASE}/v1/auth/refresh`, () => errorResponse(401, 'refresh token invalid'))
    )

    const api = useApi()

    await expect(api.shelf.listShelves()).rejects.toBeInstanceOf(ResponseError)
    expect(authStore.accessToken).toBeNull()
  })

  it('leaves a non-401 error response untouched', async () => {
    const authStore = useAuthStore()
    authStore.setTokens(buildTokenPair())
    server.use(http.get(`${BASE}/v1/shelves`, () => errorResponse(500, 'boom')))

    const api = useApi()

    await expect(api.shelf.listShelves()).rejects.toBeInstanceOf(ResponseError)
  })
})
