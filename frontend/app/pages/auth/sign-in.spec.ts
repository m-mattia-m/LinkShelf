import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { fireEvent, screen, waitFor } from '@testing-library/vue'
import { beforeEach, describe, expect, it } from 'vitest'
import { TokenPairToJSON, UserToJSON } from '~~/api'
import type { SettingPageBody } from '~~/api'
import { errorResponse } from '../../../test/mocks/handlers'
import { server } from '../../../test/mocks/server'
import { buildSettingPageBody, buildTokenPair, buildUser } from '../../../test/mocks/factories'
import SignInPage from './sign-in.vue'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  useAuthStore().$reset()
  const websiteSettings = useState<SettingPageBody | null>('settings')
  websiteSettings.value = buildSettingPageBody()
})

describe('sign-in page', () => {
  it('logs in and redirects to /app on success', async () => {
    const tokens = buildTokenPair()
    server.use(
      http.post(`${BASE}/v1/auth/login`, () => HttpResponse.json(TokenPairToJSON(tokens))),
      http.get(`${BASE}/v1/users/me`, () => HttpResponse.json(UserToJSON(buildUser({ id: 'user-1' }))))
    )

    await renderSuspended(SignInPage)

    await fireEvent.update(screen.getByLabelText('Email'), 'jane@example.com')
    await fireEvent.update(screen.getByLabelText('Password'), 'correct-password')
    await fireEvent.click(screen.getByRole('button', { name: 'Sign in' }))

    const route = useRoute()
    await waitFor(() => {
      expect(route.fullPath).toBe('/app')
    })
    expect(useAuthStore().accessToken).toBe(tokens.accessToken)
  })

  it('redirects to the "redirect" query param when present', async () => {
    server.use(
      http.post(`${BASE}/v1/auth/login`, () => HttpResponse.json(TokenPairToJSON(buildTokenPair()))),
      http.get(`${BASE}/v1/users/me`, () => HttpResponse.json(UserToJSON(buildUser())))
    )

    await renderSuspended(SignInPage, { route: { path: '/auth/sign-in', query: { redirect: '/app/shelf/abc' } } })

    await fireEvent.update(screen.getByLabelText('Email'), 'jane@example.com')
    await fireEvent.update(screen.getByLabelText('Password'), 'correct-password')
    await fireEvent.click(screen.getByRole('button', { name: 'Sign in' }))

    const route = useRoute()
    await waitFor(() => {
      expect(route.fullPath).toBe('/app/shelf/abc')
    })
  })

  it('shows a validation error and does not call the API for an invalid email', async () => {
    let called = false
    server.use(http.post(`${BASE}/v1/auth/login`, () => {
      called = true
      return HttpResponse.json(TokenPairToJSON(buildTokenPair()))
    }))

    await renderSuspended(SignInPage)

    await fireEvent.update(screen.getByLabelText('Email'), 'not-an-email')
    await fireEvent.update(screen.getByLabelText('Password'), 'correct-password')
    await fireEvent.click(screen.getByRole('button', { name: 'Sign in' }))

    await waitFor(() => {
      expect(screen.getByText('Please enter a valid email')).toBeInTheDocument()
    })
    expect(called).toBe(false)
  })

  it('shows the pending-verification alert on a 403 and offers to resend', async () => {
    server.use(http.post(`${BASE}/v1/auth/login`, () => errorResponse(403, 'email not verified')))

    await renderSuspended(SignInPage)

    await fireEvent.update(screen.getByLabelText('Email'), 'jane@example.com')
    await fireEvent.update(screen.getByLabelText('Password'), 'correct-password')
    await fireEvent.click(screen.getByRole('button', { name: 'Sign in' }))

    await waitFor(() => {
      expect(screen.getByText('Email not verified')).toBeInTheDocument()
    })

    let resendCalledWith: unknown
    server.use(http.post(`${BASE}/v1/auth/resend-verification`, async ({ request }) => {
      resendCalledWith = await request.json()
      return new HttpResponse(null, { status: 204 })
    }))

    await fireEvent.click(screen.getByRole('button', { name: 'Resend email' }))

    await waitFor(() => {
      expect(resendCalledWith).toEqual({ email: 'jane@example.com' })
    })
    expect(screen.getByRole('button', { name: 'Resend email' })).not.toHaveAttribute('aria-disabled')
  })

  it('shows the generic error message for a non-403 failure', async () => {
    server.use(http.post(`${BASE}/v1/auth/login`, () => errorResponse(401, 'invalid credentials')))

    await renderSuspended(SignInPage)

    await fireEvent.update(screen.getByLabelText('Email'), 'jane@example.com')
    await fireEvent.update(screen.getByLabelText('Password'), 'wrong-password')
    await fireEvent.click(screen.getByRole('button', { name: 'Sign in' }))

    await waitFor(() => {
      expect(screen.getByText('invalid credentials')).toBeInTheDocument()
    })
    expect(screen.queryByText('Email not verified')).not.toBeInTheDocument()
  })

  it('starts OIDC login when SSO is enabled and clicked', async () => {
    const websiteSettings = useState<SettingPageBody | null>('settings')
    websiteSettings.value = buildSettingPageBody({ oidcEnabled: true })

    server.use(http.get(`${BASE}/v1/auth/oidc/login`, () => HttpResponse.json({
      authorization_url: 'https://idp.example.com/authorize',
      state: 'my-state'
    })))
    const originalLocation = window.location
    // @ts-expect-error - reassigning window.location for the test
    delete window.location
    // @ts-expect-error - a minimal stub is enough for this redirect assertion
    window.location = { href: '' }

    await renderSuspended(SignInPage)

    await fireEvent.click(screen.getByRole('button', { name: 'Continue with SSO' }))

    await waitFor(() => {
      expect(window.location.href).toBe('https://idp.example.com/authorize')
    })
    expect(sessionStorage.getItem('linkshelf.oidcState')).toBe('my-state')
    // @ts-expect-error - restoring the real window.location after the test
    window.location = originalLocation
  })

  it('does not show the SSO button when OIDC is disabled', async () => {
    await renderSuspended(SignInPage)

    expect(screen.queryByRole('button', { name: 'Continue with SSO' })).not.toBeInTheDocument()
  })
})
