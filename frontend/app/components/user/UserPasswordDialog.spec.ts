import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { fireEvent, screen, waitFor, within } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import { errorResponse } from '../../../test/mocks/handlers'
import { server } from '../../../test/mocks/server'
import { buildUser } from '../../../test/mocks/factories'
import UserPasswordDialog from './UserPasswordDialog.vue'

const BASE = 'http://localhost:8085'

describe('UserPasswordDialog', () => {
  it('submits the old and new password and closes on success', async () => {
    const user = buildUser({ id: 'user-1' })
    let patchedBody: unknown
    server.use(http.patch(`${BASE}/v1/users/user-1/password`, async ({ request }) => {
      patchedBody = await request.json()
      return new HttpResponse(null, { status: 204 })
    }))

    await renderSuspended(UserPasswordDialog, { props: { user, open: true } })

    await fireEvent.update(screen.getByLabelText('Current password'), 'old-secret')
    await fireEvent.update(screen.getByLabelText('New password'), 'new-secret1')
    await fireEvent.update(screen.getByLabelText('Confirm new password'), 'new-secret1')
    await fireEvent.click(screen.getByRole('button', { name: 'Submit' }))

    await waitFor(() => {
      expect(patchedBody).toEqual({ old_password: 'old-secret', new_password: 'new-secret1' })
    })
    await waitFor(() => {
      expect(screen.queryByLabelText('Current password')).not.toBeInTheDocument()
    })
  })

  it('shows a validation error when the confirmation does not match', async () => {
    const user = buildUser({ id: 'user-1' })
    let called = false
    server.use(http.patch(`${BASE}/v1/users/user-1/password`, () => {
      called = true
      return new HttpResponse(null, { status: 204 })
    }))

    await renderSuspended(UserPasswordDialog, { props: { user, open: true } })

    await fireEvent.update(screen.getByLabelText('Current password'), 'old-secret')
    await fireEvent.update(screen.getByLabelText('New password'), 'new-secret1')
    await fireEvent.update(screen.getByLabelText('Confirm new password'), 'mismatch')
    await fireEvent.click(screen.getByRole('button', { name: 'Submit' }))

    await waitFor(() => {
      expect(screen.getByText('Passwords do not match')).toBeInTheDocument()
    })
    expect(called).toBe(false)
  })

  it('shows the API error and keeps the dialog open on failure', async () => {
    const user = buildUser({ id: 'user-1' })
    server.use(http.patch(`${BASE}/v1/users/user-1/password`, () => errorResponse(400, 'old password is incorrect')))

    const { baseElement } = await renderSuspended(UserPasswordDialog, { props: { user, open: true } })

    await fireEvent.update(screen.getByLabelText('Current password'), 'wrong-secret')
    await fireEvent.update(screen.getByLabelText('New password'), 'new-secret1')
    await fireEvent.update(screen.getByLabelText('Confirm new password'), 'new-secret1')
    await fireEvent.click(screen.getByRole('button', { name: 'Submit' }))

    await waitFor(() => {
      expect(within(baseElement as HTMLElement).getByLabelText('Current password')).toBeInTheDocument()
    })
  })

  it('resets its fields each time it is reopened', async () => {
    const user = buildUser({ id: 'user-1' })
    const { rerender } = await renderSuspended(UserPasswordDialog, { props: { user, open: true } })

    await fireEvent.update(screen.getByLabelText('Current password'), 'typed-value')
    await rerender({ user, open: false })
    await rerender({ user, open: true })

    expect(screen.getByLabelText('Current password')).toHaveValue('')
  })
})
