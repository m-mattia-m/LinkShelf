import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { screen, waitFor } from '@testing-library/vue'
import { beforeEach, describe, expect, it } from 'vitest'
import { TokenPairToJSON, UserToJSON } from '~~/api'
import { errorResponse } from '../../../test/mocks/handlers'
import { server } from '../../../test/mocks/server'
import { buildTokenPair, buildUser } from '../../../test/mocks/factories'
import CallbackPage from './callback.vue'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  useAuthStore().$reset()
})

describe('auth callback page', () => {
  it('completes the login and redirects to /app when code and state match the stashed value', async () => {
    sessionStorage.setItem('linkshelf.oidcState', 'my-state')
    const tokens = buildTokenPair()
    server.use(
      http.post(`${BASE}/v1/auth/oidc/callback`, () => HttpResponse.json(TokenPairToJSON(tokens))),
      http.get(`${BASE}/v1/users/me`, () => HttpResponse.json(UserToJSON(buildUser({ id: 'user-1' }))))
    )

    await renderSuspended(CallbackPage, {
      route: { path: '/auth/callback', query: { code: 'the-code', state: 'my-state' } }
    })

    const route = useRoute()
    await waitFor(() => {
      expect(route.fullPath).toBe('/app')
    })

    const authStore = useAuthStore()
    expect(authStore.accessToken).toBe(tokens.accessToken)
    expect(authStore.user?.id).toBe('user-1')
    expect(sessionStorage.getItem('linkshelf.oidcState')).toBeNull()
  })

  it('shows a failed state without calling the API when the state does not match', async () => {
    sessionStorage.setItem('linkshelf.oidcState', 'expected-state')
    let called = false
    server.use(http.post(`${BASE}/v1/auth/oidc/callback`, () => {
      called = true
      return HttpResponse.json(TokenPairToJSON(buildTokenPair()))
    }))

    await renderSuspended(CallbackPage, {
      route: { path: '/auth/callback', query: { code: 'the-code', state: 'wrong-state' } }
    })

    await waitFor(() => {
      expect(screen.getByText('Sign-in could not be completed.')).toBeInTheDocument()
    })
    expect(called).toBe(false)
    expect(useAuthStore().isAuthenticated).toBe(false)
  })

  it('shows a failed state when the code or state query params are missing', async () => {
    sessionStorage.setItem('linkshelf.oidcState', 'expected-state')

    await renderSuspended(CallbackPage, { route: { path: '/auth/callback' } })

    await waitFor(() => {
      expect(screen.getByText('Sign-in could not be completed.')).toBeInTheDocument()
    })
  })

  it('shows a failed state when there is no stashed state to compare against', async () => {
    await renderSuspended(CallbackPage, {
      route: { path: '/auth/callback', query: { code: 'the-code', state: 'my-state' } }
    })

    await waitFor(() => {
      expect(screen.getByText('Sign-in could not be completed.')).toBeInTheDocument()
    })
  })

  it('shows a failed state when the API call fails', async () => {
    sessionStorage.setItem('linkshelf.oidcState', 'my-state')
    server.use(http.post(`${BASE}/v1/auth/oidc/callback`, () => errorResponse(400, 'invalid code')))

    await renderSuspended(CallbackPage, {
      route: { path: '/auth/callback', query: { code: 'the-code', state: 'my-state' } }
    })

    await waitFor(() => {
      expect(screen.getByText('Sign-in could not be completed.')).toBeInTheDocument()
    })
    expect(useAuthStore().isAuthenticated).toBe(false)
  })
})
