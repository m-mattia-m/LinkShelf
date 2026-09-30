import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { screen, waitFor } from '@testing-library/vue'
import { beforeEach, describe, expect, it } from 'vitest'
import { errorResponse } from '../../../test/mocks/handlers'
import { server } from '../../../test/mocks/server'
import ConfirmEmailPage from './confirm-email.vue'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  useAuthStore().$reset()
})

describe('confirm email page', () => {
  it('shows a failed state when no token is present in the query', async () => {
    await renderSuspended(ConfirmEmailPage, { route: { path: '/auth/confirm-email' } })

    await waitFor(() => {
      expect(screen.getByText('This link is invalid or has expired')).toBeInTheDocument()
    })
  })

  it('confirms the change with the token and shows the success state', async () => {
    let calledWith: unknown
    server.use(http.post(`${BASE}/v1/auth/confirm-email-change`, async ({ request }) => {
      calledWith = await request.json()
      return new HttpResponse(null, { status: 204 })
    }))

    await renderSuspended(ConfirmEmailPage, { route: { path: '/auth/confirm-email', query: { token: 'valid-token' } } })

    await waitFor(() => {
      expect(screen.getByText('Your new email is confirmed.')).toBeInTheDocument()
    })
    expect(calledWith).toEqual({ token: 'valid-token' })
    expect(screen.getByRole('link', { name: 'Continue' })).toHaveAttribute('href', '/auth/sign-in')
  })

  it('explains when another account took the address meanwhile', async () => {
    server.use(http.post(`${BASE}/v1/auth/confirm-email-change`, () => errorResponse(409, 'taken')))

    await renderSuspended(ConfirmEmailPage, { route: { path: '/auth/confirm-email', query: { token: 'valid-token' } } })

    await waitFor(() => {
      expect(screen.getByText('This email is already in use')).toBeInTheDocument()
    })
  })

  it('shows the failed state for an invalid or expired link', async () => {
    server.use(http.post(`${BASE}/v1/auth/confirm-email-change`, () => errorResponse(400, 'expired')))

    await renderSuspended(ConfirmEmailPage, { route: { path: '/auth/confirm-email', query: { token: 'expired-token' } } })

    await waitFor(() => {
      expect(screen.getByText('This link is invalid or has expired')).toBeInTheDocument()
    })
  })
})
