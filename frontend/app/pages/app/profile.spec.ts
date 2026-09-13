import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { fireEvent, screen, waitFor } from '@testing-library/vue'
import { beforeEach, describe, expect, it } from 'vitest'
import { UserToJSON } from '~~/api'
import { errorResponse } from '../../../test/mocks/handlers'
import { server } from '../../../test/mocks/server'
import { buildUser } from '../../../test/mocks/factories'
import ProfilePage from './profile.vue'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  useAuthStore().$reset()
})

describe('app profile page', () => {
  it('loads the current user and pre-fills the form', async () => {
    server.use(http.get(`${BASE}/v1/users/me`, () => HttpResponse.json(UserToJSON(buildUser({ id: 'user-1', firstName: 'Jane', lastName: 'Doe', email: 'jane@example.com', role: 'user' })))))
    useAuthStore().setTokens({ accessToken: 'header.eyJzdWIiOiJ1c2VyLTEiLCJyb2xlIjoidXNlciIsImV4cCI6OTk5OTk5OTk5OX0.signature', refreshToken: 'refresh-1' })

    await renderSuspended(ProfilePage)

    await waitFor(() => {
      expect(screen.getByLabelText('First name')).toHaveValue('Jane')
    })
    expect(screen.getByLabelText('Last name')).toHaveValue('Doe')
    expect(screen.getByLabelText('Email')).toHaveValue('jane@example.com')
    expect(screen.getByText('user')).toBeInTheDocument()
  })

  it('saves profile changes and updates the auth store user', async () => {
    server.use(http.get(`${BASE}/v1/users/me`, () => HttpResponse.json(UserToJSON(buildUser({ id: 'user-1', firstName: 'Jane', lastName: 'Doe', email: 'jane@example.com' })))))
    useAuthStore().setTokens({ accessToken: 'header.eyJzdWIiOiJ1c2VyLTEiLCJyb2xlIjoidXNlciIsImV4cCI6OTk5OTk5OTk5OX0.signature', refreshToken: 'refresh-1' })

    let putBody: unknown
    server.use(http.put(`${BASE}/v1/users/user-1`, async ({ request }) => {
      putBody = await request.json()
      return HttpResponse.json(UserToJSON(buildUser({ id: 'user-1', firstName: 'Janet', lastName: 'Doe', email: 'jane@example.com' })))
    }))

    await renderSuspended(ProfilePage)

    await waitFor(() => {
      expect(screen.getByLabelText('First name')).toHaveValue('Jane')
    })

    await fireEvent.update(screen.getByLabelText('First name'), 'Janet')
    await fireEvent.click(screen.getByRole('button', { name: 'Save' }))

    await waitFor(() => {
      expect(putBody).toEqual({ email: 'jane@example.com', first_name: 'Janet', last_name: 'Doe', role: undefined })
    })
    expect(useAuthStore().user?.firstName).toBe('Janet')
  })

  it('shows a validation error for a blank required field without saving', async () => {
    server.use(http.get(`${BASE}/v1/users/me`, () => HttpResponse.json(UserToJSON(buildUser({ id: 'user-1' })))))
    useAuthStore().setTokens({ accessToken: 'header.eyJzdWIiOiJ1c2VyLTEiLCJyb2xlIjoidXNlciIsImV4cCI6OTk5OTk5OTk5OX0.signature', refreshToken: 'refresh-1' })
    let called = false
    server.use(http.put(`${BASE}/v1/users/user-1`, () => {
      called = true
      return HttpResponse.json(UserToJSON(buildUser({ id: 'user-1' })))
    }))

    await renderSuspended(ProfilePage)

    await waitFor(() => {
      expect(screen.getByLabelText('First name')).not.toHaveValue('')
    })

    await fireEvent.update(screen.getByLabelText('First name'), '')
    await fireEvent.click(screen.getByRole('button', { name: 'Save' }))

    await waitFor(() => {
      expect(screen.getByText('First name is required')).toBeInTheDocument()
    })
    expect(called).toBe(false)
  })

  it('re-enables the form after the API rejects the save', async () => {
    server.use(http.get(`${BASE}/v1/users/me`, () => HttpResponse.json(UserToJSON(buildUser({ id: 'user-1', email: 'jane@example.com' })))))
    useAuthStore().setTokens({ accessToken: 'header.eyJzdWIiOiJ1c2VyLTEiLCJyb2xlIjoidXNlciIsImV4cCI6OTk5OTk5OTk5OX0.signature', refreshToken: 'refresh-1' })
    let putCalled = false
    server.use(http.put(`${BASE}/v1/users/user-1`, () => {
      putCalled = true
      return errorResponse(422, 'validation failed', [{ location: 'body.email', message: 'already registered' }])
    }))

    await renderSuspended(ProfilePage)

    await waitFor(() => {
      expect(screen.getByLabelText('Email')).toHaveValue('jane@example.com')
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Save' }))

    await waitFor(() => {
      expect(putCalled).toBe(true)
    })
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Save' })).not.toHaveAttribute('aria-disabled')
    })
    expect(useAuthStore().user?.email).toBe('jane@example.com')
  })

  it('opens the change-password dialog', async () => {
    server.use(http.get(`${BASE}/v1/users/me`, () => HttpResponse.json(UserToJSON(buildUser({ id: 'user-1' })))))
    useAuthStore().setTokens({ accessToken: 'header.eyJzdWIiOiJ1c2VyLTEiLCJyb2xlIjoidXNlciIsImV4cCI6OTk5OTk5OTk5OX0.signature', refreshToken: 'refresh-1' })

    await renderSuspended(ProfilePage)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Change password' })).toBeInTheDocument()
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Change password' }))

    await waitFor(() => {
      expect(screen.getByLabelText('Current password')).toBeInTheDocument()
    })
  })
})
