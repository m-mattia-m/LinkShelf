import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { screen, waitFor } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import { errorResponse } from '../../../test/mocks/handlers'
import { server } from '../../../test/mocks/server'
import VerifyEmailPage from './verify-email.vue'

const BASE = 'http://localhost:8085'

describe('verify email page', () => {
  it('shows a failed state when no token is present in the query', async () => {
    await renderSuspended(VerifyEmailPage, { route: { path: '/auth/verify-email' } })

    await waitFor(() => {
      expect(screen.getByText('This link is invalid or has expired')).toBeInTheDocument()
    })
  })

  it('verifies the email and shows the success state when the token is valid', async () => {
    let calledWith: unknown
    server.use(http.post(`${BASE}/v1/auth/verify-email`, async ({ request }) => {
      calledWith = await request.json()
      return new HttpResponse(null, { status: 204 })
    }))

    await renderSuspended(VerifyEmailPage, { route: { path: '/auth/verify-email', query: { token: 'valid-token' } } })

    await waitFor(() => {
      expect(screen.getByText('Your email has been verified.')).toBeInTheDocument()
    })
    expect(calledWith).toEqual({ token: 'valid-token' })
  })

  it('shows the failed state when verification fails', async () => {
    server.use(http.post(`${BASE}/v1/auth/verify-email`, () => errorResponse(400, 'token expired')))

    await renderSuspended(VerifyEmailPage, { route: { path: '/auth/verify-email', query: { token: 'expired-token' } } })

    await waitFor(() => {
      expect(screen.getByText('This link is invalid or has expired')).toBeInTheDocument()
    })
  })
})
