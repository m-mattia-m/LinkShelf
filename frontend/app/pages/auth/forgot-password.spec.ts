import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { fireEvent, screen, waitFor } from '@testing-library/vue'
import { beforeEach, describe, expect, it } from 'vitest'
import type { SettingPageBody } from '~~/api'
import { errorResponse } from '../../../test/mocks/handlers'
import { server } from '../../../test/mocks/server'
import { buildSettingPageBody } from '../../../test/mocks/factories'
import ForgotPasswordPage from './forgot-password.vue'

const BASE = 'http://localhost:8085'

function setPasswordResetEnabled(enabled: boolean) {
  useState<SettingPageBody | null>('settings').value = buildSettingPageBody({ passwordResetEnabled: enabled })
}

beforeEach(() => {
  setPasswordResetEnabled(true)
})

describe('forgot password page', () => {
  it('asks for the email address and sends the request', async () => {
    let postedBody: unknown
    server.use(http.post(`${BASE}/v1/auth/forgot-password`, async ({ request }) => {
      postedBody = await request.json()
      return new HttpResponse(null, { status: 204 })
    }))

    await renderSuspended(ForgotPasswordPage)

    await fireEvent.update(screen.getByLabelText('Email'), 'ada@example.com')
    await fireEvent.click(screen.getByRole('button', { name: 'Send reset link' }))

    await waitFor(() => {
      expect(screen.getByText('Check your email')).toBeInTheDocument()
    })
    expect(postedBody).toEqual({ email: 'ada@example.com' })
    expect(screen.getByText(/If an account exists for ada@example\.com/)).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Back to sign in' })).toHaveAttribute('href', '/auth/sign-in')
  })

  it('never says whether the address has an account', async () => {
    // The backend answers 204 for every address; the page must read the same.
    server.use(http.post(`${BASE}/v1/auth/forgot-password`, () => new HttpResponse(null, { status: 204 })))

    await renderSuspended(ForgotPasswordPage)

    await fireEvent.update(screen.getByLabelText('Email'), 'nobody@example.com')
    await fireEvent.click(screen.getByRole('button', { name: 'Send reset link' }))

    await waitFor(() => {
      expect(screen.getByText('Check your email')).toBeInTheDocument()
    })
    expect(screen.queryByText(/no account|not found|doesn't exist/i)).not.toBeInTheDocument()
  })

  it('validates the address without calling the API', async () => {
    let called = false
    server.use(http.post(`${BASE}/v1/auth/forgot-password`, () => {
      called = true
      return new HttpResponse(null, { status: 204 })
    }))

    await renderSuspended(ForgotPasswordPage)

    await fireEvent.update(screen.getByLabelText('Email'), 'not-an-email')
    await fireEvent.click(screen.getByRole('button', { name: 'Send reset link' }))

    await waitFor(() => {
      expect(screen.getByText('Please enter a valid email')).toBeInTheDocument()
    })
    expect(called).toBe(false)
  })

  it('shows the API error and keeps the form when the request fails', async () => {
    server.use(http.post(`${BASE}/v1/auth/forgot-password`, () => errorResponse(500, 'mail server unreachable')))

    await renderSuspended(ForgotPasswordPage)

    await fireEvent.update(screen.getByLabelText('Email'), 'ada@example.com')
    await fireEvent.click(screen.getByRole('button', { name: 'Send reset link' }))

    await waitFor(() => {
      expect(screen.getByText('mail server unreachable')).toBeInTheDocument()
    })
    expect(screen.queryByText('Check your email')).not.toBeInTheDocument()
    expect(screen.getByLabelText('Email')).toBeInTheDocument()
  })

  it('explains that reset is unavailable when the instance has it switched off', async () => {
    setPasswordResetEnabled(false)

    await renderSuspended(ForgotPasswordPage)

    expect(screen.getByText(/Password reset is not available here/)).toBeInTheDocument()
    expect(screen.queryByLabelText('Email')).not.toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Back to sign in' })).toHaveAttribute('href', '/auth/sign-in')
  })
})
