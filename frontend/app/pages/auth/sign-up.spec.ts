import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { fireEvent, screen, waitFor } from '@testing-library/vue'
import { beforeEach, describe, expect, it } from 'vitest'
import { UserToJSON } from '~~/api'
import type { SettingPageBody } from '~~/api'
import { errorResponse } from '../../../test/mocks/handlers'
import { server } from '../../../test/mocks/server'
import { buildSettingPageBody, buildUser } from '../../../test/mocks/factories'
import SignUpPage from './sign-up.vue'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  useAuthStore().$reset()
  const websiteSettings = useState<SettingPageBody | null>('settings')
  websiteSettings.value = buildSettingPageBody({ registrationEnabled: true })
})

async function fillForm() {
  await fireEvent.update(screen.getByLabelText('First name'), 'Ada')
  await fireEvent.update(screen.getByLabelText('Last name'), 'Lovelace')
  await fireEvent.update(screen.getByLabelText('Email'), 'ada@example.com')
  await fireEvent.update(screen.getByLabelText('Password'), 'secret123')
}

describe('sign-up page', () => {
  it('registers and redirects to /app when verification is not required', async () => {
    let postedBody: unknown
    server.use(
      http.post(`${BASE}/v1/users`, async ({ request }) => {
        postedBody = await request.json()
        return HttpResponse.json(UserToJSON(buildUser({ id: 'user-1' })))
      }),
      http.post(`${BASE}/v1/auth/login`, () => HttpResponse.json({
        access_token: 'header.eyJzdWIiOiJ1c2VyLTEiLCJyb2xlIjoidXNlciIsImV4cCI6OTk5OTk5OTk5OX0.signature',
        refresh_token: 'refresh-1'
      })),
      http.get(`${BASE}/v1/users/me`, () => HttpResponse.json(UserToJSON(buildUser({ id: 'user-1' }))))
    )

    await renderSuspended(SignUpPage)

    await fillForm()
    await fireEvent.click(screen.getByRole('button', { name: 'Create account' }))

    const route = useRoute()
    await waitFor(() => {
      expect(route.fullPath).toBe('/app')
    })
    expect(postedBody).toEqual({
      email: 'ada@example.com',
      first_name: 'Ada',
      last_name: 'Lovelace',
      password: 'secret123',
      role: undefined
    })
    expect(useAuthStore().isAuthenticated).toBe(true)
  })

  it('shows the check-your-email view when the account is pending verification', async () => {
    server.use(
      http.post(`${BASE}/v1/users`, () => HttpResponse.json(UserToJSON(buildUser()))),
      http.post(`${BASE}/v1/auth/login`, () => errorResponse(403, 'email not verified'))
    )

    await renderSuspended(SignUpPage)

    await fillForm()
    await fireEvent.click(screen.getByRole('button', { name: 'Create account' }))

    await waitFor(() => {
      expect(screen.getByText('Check your email')).toBeInTheDocument()
    })
    expect(useAuthStore().isAuthenticated).toBe(false)
    expect(screen.queryByLabelText('Email')).not.toBeInTheDocument()
  })

  it('shows a validation error for a short password without calling the API', async () => {
    let called = false
    server.use(http.post(`${BASE}/v1/users`, () => {
      called = true
      return HttpResponse.json(UserToJSON(buildUser()))
    }))

    await renderSuspended(SignUpPage)

    await fireEvent.update(screen.getByLabelText('First name'), 'Ada')
    await fireEvent.update(screen.getByLabelText('Last name'), 'Lovelace')
    await fireEvent.update(screen.getByLabelText('Email'), 'ada@example.com')
    await fireEvent.update(screen.getByLabelText('Password'), 'short')
    await fireEvent.click(screen.getByRole('button', { name: 'Create account' }))

    await waitFor(() => {
      expect(screen.getByText('Must be at least 8 characters')).toBeInTheDocument()
    })
    expect(called).toBe(false)
  })

  it('shows the API error message when registration fails', async () => {
    server.use(http.post(`${BASE}/v1/users`, () => errorResponse(422, 'validation failed', [
      { location: 'body.email', message: 'already registered' }
    ])))

    await renderSuspended(SignUpPage)

    await fillForm()
    await fireEvent.click(screen.getByRole('button', { name: 'Create account' }))

    await waitFor(() => {
      expect(screen.getByText('email: already registered')).toBeInTheDocument()
    })
  })

  it('shows a registration-disabled message instead of the form when registration is off', async () => {
    const websiteSettings = useState<SettingPageBody | null>('settings')
    websiteSettings.value = buildSettingPageBody({ registrationEnabled: false })

    await renderSuspended(SignUpPage)

    expect(screen.getByText('Registration is currently disabled. Please contact an administrator for an invite.')).toBeInTheDocument()
    expect(screen.queryByLabelText('Email')).not.toBeInTheDocument()
  })
})
