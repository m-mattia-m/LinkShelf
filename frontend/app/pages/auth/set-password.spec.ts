import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { fireEvent, screen, waitFor } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import { errorResponse } from '../../../test/mocks/handlers'
import { server } from '../../../test/mocks/server'
import SetPasswordPage from './set-password.vue'

const BASE = 'http://localhost:8085'

describe('set password page', () => {
  it('shows a missing-token message when there is no token in the query', async () => {
    await renderSuspended(SetPasswordPage, { route: { path: '/auth/set-password' } })

    expect(screen.getByText('This link is missing its token.')).toBeInTheDocument()
    expect(screen.queryByLabelText('New password')).not.toBeInTheDocument()
  })

  it('sets the password and shows the success view', async () => {
    let postedBody: unknown
    server.use(http.post(`${BASE}/v1/auth/set-password`, async ({ request }) => {
      postedBody = await request.json()
      return new HttpResponse(null, { status: 204 })
    }))

    await renderSuspended(SetPasswordPage, { route: { path: '/auth/set-password', query: { token: 'reset-token' } } })

    await fireEvent.update(screen.getByLabelText('New password'), 'newsecret1')
    await fireEvent.update(screen.getByLabelText('Confirm password'), 'newsecret1')
    await fireEvent.click(screen.getByRole('button', { name: 'Set password' }))

    await waitFor(() => {
      expect(screen.getByText('Your password has been set.')).toBeInTheDocument()
    })
    expect(postedBody).toEqual({ token: 'reset-token', new_password: 'newsecret1' })
  })

  it('shows a validation error when the passwords do not match, without calling the API', async () => {
    let called = false
    server.use(http.post(`${BASE}/v1/auth/set-password`, () => {
      called = true
      return new HttpResponse(null, { status: 204 })
    }))

    await renderSuspended(SetPasswordPage, { route: { path: '/auth/set-password', query: { token: 'reset-token' } } })

    await fireEvent.update(screen.getByLabelText('New password'), 'newsecret1')
    await fireEvent.update(screen.getByLabelText('Confirm password'), 'different')
    await fireEvent.click(screen.getByRole('button', { name: 'Set password' }))

    await waitFor(() => {
      expect(screen.getByText('Passwords do not match')).toBeInTheDocument()
    })
    expect(called).toBe(false)
  })

  it('shows the API error message when the token is invalid or expired', async () => {
    server.use(http.post(`${BASE}/v1/auth/set-password`, () => errorResponse(400, 'token is invalid or expired')))

    await renderSuspended(SetPasswordPage, { route: { path: '/auth/set-password', query: { token: 'bad-token' } } })

    await fireEvent.update(screen.getByLabelText('New password'), 'newsecret1')
    await fireEvent.update(screen.getByLabelText('Confirm password'), 'newsecret1')
    await fireEvent.click(screen.getByRole('button', { name: 'Set password' }))

    await waitFor(() => {
      expect(screen.getByText('token is invalid or expired')).toBeInTheDocument()
    })
    expect(screen.queryByText('Your password has been set.')).not.toBeInTheDocument()
  })
})
