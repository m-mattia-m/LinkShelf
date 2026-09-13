import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { fireEvent, screen, waitFor, within } from '@testing-library/vue'
import { beforeEach, describe, expect, it } from 'vitest'
import { UserToJSON } from '~~/api'
import type { SettingPageBody } from '~~/api'
import { errorResponse } from '../../../../test/mocks/handlers'
import { server } from '../../../../test/mocks/server'
import { buildSettingPageBody, buildTokenPair, buildUser } from '../../../../test/mocks/factories'
import { resetOnceCache } from '../../../../test/reset-once-cache'
import UsersSettingsPage from './users.vue'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  resetOnceCache()
  useAuthStore().$reset()
  const websiteSettings = useState<SettingPageBody | null>('settings')
  websiteSettings.value = buildSettingPageBody({ emailVerificationEnabled: false })
})

function signInAsAdmin() {
  useAuthStore().setTokens(buildTokenPair({
    accessToken: 'header.eyJzdWIiOiJ1c2VyLTEiLCJyb2xlIjoiYWRtaW4iLCJleHAiOjk5OTk5OTk5OTl9.signature'
  }))
}

describe('app settings users page', () => {
  it('shows a loading state, then the list of users', async () => {
    signInAsAdmin()
    server.use(http.get(`${BASE}/v1/users`, () => HttpResponse.json([
      buildUser({ id: 'user-1', firstName: 'Jane', lastName: 'Doe', email: 'jane@example.com' })
    ].map(UserToJSON))))

    await renderSuspended(UsersSettingsPage)

    await waitFor(() => {
      expect(screen.getByText('jane@example.com')).toBeInTheDocument()
    })
  })

  it('shows an empty-state prompt when there are no users', async () => {
    signInAsAdmin()
    server.use(http.get(`${BASE}/v1/users`, () => HttpResponse.json([])))

    await renderSuspended(UsersSettingsPage)

    await waitFor(() => {
      expect(screen.getByText('There are no users yet.')).toBeInTheDocument()
    })
  })

  it('edits a user through the dropdown menu, pre-filling the dialog', async () => {
    signInAsAdmin()
    const user = buildUser({ id: 'user-2', firstName: 'Old', lastName: 'Name', email: 'old@example.com', role: 'user' })
    let putBody: unknown
    server.use(
      http.get(`${BASE}/v1/users`, () => HttpResponse.json([user].map(UserToJSON))),
      http.put(`${BASE}/v1/users/user-2`, async ({ request }) => {
        putBody = await request.json()
        return HttpResponse.json(UserToJSON({ ...user, firstName: 'New' }))
      })
    )

    const { baseElement } = await renderSuspended(UsersSettingsPage)

    await waitFor(() => {
      expect(screen.getByText('old@example.com')).toBeInTheDocument()
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Actions' }))
    const dialog = within(baseElement as HTMLElement)
    await fireEvent.click(await dialog.findByText('Edit'))

    expect(dialog.getByLabelText('First name')).toHaveValue('Old')

    await fireEvent.update(dialog.getByLabelText('First name'), 'New')
    await fireEvent.click(dialog.getByRole('button', { name: 'Submit' }))

    await waitFor(() => {
      expect(putBody).toMatchObject({ first_name: 'New' })
    })
  })

  it('deletes another user after confirming, but offers change-password instead of delete for the current user', async () => {
    signInAsAdmin()
    const otherUser = buildUser({ id: 'user-2', email: 'other@example.com' })
    let deleteCalled = false
    server.use(
      http.get(`${BASE}/v1/users`, () => HttpResponse.json([otherUser].map(UserToJSON))),
      http.delete(`${BASE}/v1/users/user-2`, () => {
        deleteCalled = true
        return new HttpResponse(null, { status: 204 })
      })
    )

    const { baseElement } = await renderSuspended(UsersSettingsPage)

    await waitFor(() => {
      expect(screen.getByText('other@example.com')).toBeInTheDocument()
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Actions' }))
    const dialog = within(baseElement as HTMLElement)
    expect(dialog.queryByText('Change password')).not.toBeInTheDocument()
    await fireEvent.click(await dialog.findByText('Delete'))

    const confirmButton = await dialog.findByRole('button', { name: /^Delete$/, hidden: true })
    await fireEvent.click(confirmButton)

    await waitFor(() => {
      expect(deleteCalled).toBe(true)
    })
  })

  it('offers change-password instead of delete for the signed-in user\'s own row', async () => {
    signInAsAdmin()
    const self = buildUser({ id: 'user-1', email: 'me@example.com' })
    server.use(http.get(`${BASE}/v1/users`, () => HttpResponse.json([self].map(UserToJSON))))

    const { baseElement } = await renderSuspended(UsersSettingsPage)

    await waitFor(() => {
      expect(screen.getByText('me@example.com')).toBeInTheDocument()
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Actions' }))
    const dialog = within(baseElement as HTMLElement)
    expect(await dialog.findByText('Change password')).toBeInTheDocument()
    expect(dialog.queryByText('Delete')).not.toBeInTheDocument()
  })

  it('resends the verification email and marks a user verified when email verification is enabled', async () => {
    signInAsAdmin()
    const websiteSettings = useState<SettingPageBody | null>('settings')
    websiteSettings.value = buildSettingPageBody({ emailVerificationEnabled: true })

    const pending = buildUser({ id: 'user-2', email: 'pending@example.com', emailVerified: false, hasPassword: true })
    let resendCalledWith: unknown
    let markVerifiedCalled = false
    server.use(
      http.get(`${BASE}/v1/users`, () => HttpResponse.json([pending].map(UserToJSON))),
      http.post(`${BASE}/v1/auth/resend-verification`, async ({ request }) => {
        resendCalledWith = await request.json()
        return new HttpResponse(null, { status: 204 })
      }),
      http.patch(`${BASE}/v1/users/user-2/verify`, () => {
        markVerifiedCalled = true
        return HttpResponse.json(UserToJSON({ ...pending, emailVerified: true }))
      })
    )

    const { baseElement } = await renderSuspended(UsersSettingsPage)

    await waitFor(() => {
      expect(screen.getByText('Pending verification')).toBeInTheDocument()
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Actions' }))
    const dialog = within(baseElement as HTMLElement)
    await fireEvent.click(await dialog.findByText('Resend verification email'))

    await waitFor(() => {
      expect(resendCalledWith).toEqual({ email: 'pending@example.com' })
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Actions' }))
    await fireEvent.click(await dialog.findByText('Mark as verified'))

    await waitFor(() => {
      expect(markVerifiedCalled).toBe(true)
    })
  })

  it('shows the API error and stops the action-loading state when deletion fails', async () => {
    signInAsAdmin()
    const otherUser = buildUser({ id: 'user-2', email: 'other@example.com' })
    server.use(
      http.get(`${BASE}/v1/users`, () => HttpResponse.json([otherUser].map(UserToJSON))),
      http.delete(`${BASE}/v1/users/user-2`, () => errorResponse(500, 'boom'))
    )

    const { baseElement } = await renderSuspended(UsersSettingsPage)

    await waitFor(() => {
      expect(screen.getByText('other@example.com')).toBeInTheDocument()
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Actions' }))
    const dialog = within(baseElement as HTMLElement)
    await fireEvent.click(await dialog.findByText('Delete'))
    const confirmButton = await dialog.findByRole('button', { name: /^Delete$/, hidden: true })
    await fireEvent.click(confirmButton)

    await waitFor(() => {
      expect(screen.getByText('other@example.com')).toBeInTheDocument()
    })
  })
})
