import { HttpResponse, http } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'
import { ResponseError, TokenPairToJSON } from '~~/api'
import { errorResponse } from '../../test/mocks/handlers'
import { server } from '../../test/mocks/server'
import { buildTokenPair } from '../../test/mocks/factories'
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
