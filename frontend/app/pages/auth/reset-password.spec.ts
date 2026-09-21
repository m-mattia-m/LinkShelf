import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { fireEvent, screen, waitFor } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import { errorResponse } from '../../../test/mocks/handlers'
import { server } from '../../../test/mocks/server'
import ResetPasswordPage from './reset-password.vue'

const BASE = 'http://localhost:8085'

async function fill(password: string, confirmation = password) {
  await fireEvent.update(screen.getByLabelText('New password'), password)
  await fireEvent.update(screen.getByLabelText('Confirm new password'), confirmation)
  await fireEvent.click(screen.getByRole('button', { name: 'Change password' }))
}

describe('reset password page', () => {
  it('points to a new link when the token is missing', async () => {
    await renderSuspended(ResetPasswordPage, { route: { path: '/auth/reset-password' } })

    expect(screen.getByText('This link is incomplete. Request a new one to continue.')).toBeInTheDocument()
    expect(screen.queryByLabelText('New password')).not.toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Request a new link' })).toHaveAttribute('href', '/auth/forgot-password')
  })

  it('sends the token and the new password, then offers to sign in', async () => {
    let postedBody: unknown
    server.use(http.post(`${BASE}/v1/auth/reset-password`, async ({ request }) => {
      postedBody = await request.json()
      return new HttpResponse(null, { status: 204 })
    }))

    await renderSuspended(ResetPasswordPage, { route: { path: '/auth/reset-password', query: { token: 'emailed-token' } } })

    await fill('a-brand-new-password')

    await waitFor(() => {
      expect(screen.getByText('Your password was changed')).toBeInTheDocument()
    })
    expect(postedBody).toEqual({ token: 'emailed-token', new_password: 'a-brand-new-password' })
    expect(screen.getByRole('link', { name: 'Sign in' })).toHaveAttribute('href', '/auth/sign-in')
  })

  it('rejects mismatching passwords without calling the API', async () => {
    let called = false
    server.use(http.post(`${BASE}/v1/auth/reset-password`, () => {
      called = true
      return new HttpResponse(null, { status: 204 })
    }))

    await renderSuspended(ResetPasswordPage, { route: { path: '/auth/reset-password', query: { token: 'emailed-token' } } })

    await fill('a-brand-new-password', 'something-else')

    await waitFor(() => {
      expect(screen.getByText('Passwords do not match')).toBeInTheDocument()
    })
    expect(called).toBe(false)
  })

  it('rejects a password shorter than eight characters without calling the API', async () => {
    let called = false
    server.use(http.post(`${BASE}/v1/auth/reset-password`, () => {
      called = true
      return new HttpResponse(null, { status: 204 })
    }))

    await renderSuspended(ResetPasswordPage, { route: { path: '/auth/reset-password', query: { token: 'emailed-token' } } })

    await fill('short')

    await waitFor(() => {
      expect(screen.getByText('Must be at least 8 characters')).toBeInTheDocument()
    })
    expect(called).toBe(false)
  })

  it('shows the error and a way to request a fresh link when the link is invalid or expired', async () => {
    server.use(http.post(`${BASE}/v1/auth/reset-password`, () => errorResponse(400, 'invalid or expired reset link')))

    await renderSuspended(ResetPasswordPage, { route: { path: '/auth/reset-password', query: { token: 'stale-token' } } })

    await fill('a-brand-new-password')

    await waitFor(() => {
      expect(screen.getByText('invalid or expired reset link')).toBeInTheDocument()
    })
    expect(screen.queryByText('Your password was changed')).not.toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Request a new link' })).toHaveAttribute('href', '/auth/forgot-password')
  })
})
