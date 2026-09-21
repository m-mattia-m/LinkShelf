import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { fireEvent, screen, waitFor, within } from '@testing-library/vue'
import { beforeEach, describe, expect, it } from 'vitest'
import { UserToJSON } from '~~/api'
import type { SettingPageBody } from '~~/api'
import { errorResponse } from '../../../test/mocks/handlers'
import { server } from '../../../test/mocks/server'
import { buildSettingPageBody, buildShelf, buildUser } from '../../../test/mocks/factories'
import { useShelfStore } from '~/stores/shelf'
import UserFormDialog from './UserFormDialog.vue'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  const websiteSettings = useState<SettingPageBody | null>('settings')
  websiteSettings.value = buildSettingPageBody({ emailVerificationEnabled: false })
})

describe('UserFormDialog', () => {
  it('creates a user with the entered fields, including password when required', async () => {
    let postedBody: unknown
    server.use(
      http.get(`${BASE}/v1/users`, () => HttpResponse.json([buildUser({ id: 'user-1' })].map(UserToJSON))),
      http.post(`${BASE}/v1/users`, async ({ request }) => {
        postedBody = await request.json()
        return HttpResponse.json(UserToJSON(buildUser({ id: 'user-new' })))
      })
    )

    const { emitted, baseElement } = await renderSuspended(UserFormDialog)

    await fireEvent.click(screen.getByRole('button', { name: 'New' }))
    const dialog = within(baseElement as HTMLElement)
    await fireEvent.update(dialog.getByLabelText('First name'), 'Ada')
    await fireEvent.update(dialog.getByLabelText('Last name'), 'Lovelace')
    await fireEvent.update(dialog.getByLabelText('Username'), 'ada-lovelace')
    await fireEvent.update(dialog.getByLabelText('Email'), 'ada@example.com')
    await fireEvent.update(dialog.getByLabelText('Password'), 'secret123')
    await fireEvent.click(dialog.getByRole('button', { name: 'Submit' }))

    await waitFor(() => {
      expect(emitted().saved).toBeTruthy()
    })
    expect(postedBody).toEqual({
      email: 'ada@example.com',
      username: 'ada-lovelace',
      first_name: 'Ada',
      last_name: 'Lovelace',
      password: 'secret123',
      role: 'user'
    })
  })

  it('requires a password on create when email verification is disabled', async () => {
    let called = false
    server.use(http.post(`${BASE}/v1/users`, () => {
      called = true
      return HttpResponse.json(UserToJSON(buildUser()))
    }))

    const { baseElement } = await renderSuspended(UserFormDialog)

    await fireEvent.click(screen.getByRole('button', { name: 'New' }))
    const dialog = within(baseElement as HTMLElement)
    await fireEvent.update(dialog.getByLabelText('First name'), 'Ada')
    await fireEvent.update(dialog.getByLabelText('Last name'), 'Lovelace')
    await fireEvent.update(dialog.getByLabelText('Username'), 'ada-lovelace')
    await fireEvent.update(dialog.getByLabelText('Email'), 'ada@example.com')
    await fireEvent.click(dialog.getByRole('button', { name: 'Submit' }))

    await waitFor(() => {
      expect(dialog.getByText('Required')).toBeInTheDocument()
    })
    expect(called).toBe(false)
  })

  it('allows leaving the password blank when email verification is enabled (invite flow)', async () => {
    const websiteSettings = useState<SettingPageBody | null>('settings')
    websiteSettings.value = buildSettingPageBody({ emailVerificationEnabled: true })

    let postedBody: unknown
    server.use(http.post(`${BASE}/v1/users`, async ({ request }) => {
      postedBody = await request.json()
      return HttpResponse.json(UserToJSON(buildUser()))
    }))

    const { emitted, baseElement } = await renderSuspended(UserFormDialog)

    await fireEvent.click(screen.getByRole('button', { name: 'New' }))
    const dialog = within(baseElement as HTMLElement)
    await fireEvent.update(dialog.getByLabelText('First name'), 'Ada')
    await fireEvent.update(dialog.getByLabelText('Last name'), 'Lovelace')
    await fireEvent.update(dialog.getByLabelText('Username'), 'ada-lovelace')
    await fireEvent.update(dialog.getByLabelText('Email'), 'ada@example.com')
    await fireEvent.click(dialog.getByRole('button', { name: 'Submit' }))

    await waitFor(() => {
      expect(emitted().saved).toBeTruthy()
    })
    expect(postedBody).toMatchObject({ password: '' })
  })

  it('edits an existing user, pre-filling the form and omitting the password field', async () => {
    const user = buildUser({ id: 'user-1', firstName: 'Old', lastName: 'Name', email: 'old@example.com', role: 'user' })
    let putBody: unknown
    server.use(
      http.get(`${BASE}/v1/users`, () => HttpResponse.json([user].map(UserToJSON))),
      http.put(`${BASE}/v1/users/user-1`, async ({ request }) => {
        putBody = await request.json()
        return HttpResponse.json(UserToJSON({ ...user, firstName: 'New' }))
      })
    )

    const { emitted, baseElement } = await renderSuspended(UserFormDialog, {
      props: { mode: 'edit', user, open: true }
    })

    const dialog = within(baseElement as HTMLElement)
    expect(dialog.getByLabelText('First name')).toHaveValue('Old')
    expect(dialog.queryByLabelText('Password')).not.toBeInTheDocument()

    await fireEvent.update(dialog.getByLabelText('First name'), 'New')
    await fireEvent.click(dialog.getByRole('button', { name: 'Submit' }))

    await waitFor(() => {
      expect(emitted().saved).toBeTruthy()
    })
    expect(putBody).toEqual({
      email: 'old@example.com',
      username: 'user-1',
      first_name: 'New',
      last_name: 'Name',
      role: 'user'
    })
  })

  it('shows the API field error and keeps the dialog open on failure', async () => {
    server.use(http.post(`${BASE}/v1/users`, () => errorResponse(422, 'validation failed', [
      { location: 'body.email', message: 'already registered' }
    ])))

    const { emitted, baseElement } = await renderSuspended(UserFormDialog)

    await fireEvent.click(screen.getByRole('button', { name: 'New' }))
    const dialog = within(baseElement as HTMLElement)
    await fireEvent.update(dialog.getByLabelText('First name'), 'Ada')
    await fireEvent.update(dialog.getByLabelText('Last name'), 'Lovelace')
    await fireEvent.update(dialog.getByLabelText('Username'), 'ada-lovelace')
    await fireEvent.update(dialog.getByLabelText('Email'), 'taken@example.com')
    await fireEvent.update(dialog.getByLabelText('Password'), 'secret123')
    await fireEvent.click(dialog.getByRole('button', { name: 'Submit' }))

    await waitFor(() => {
      expect(dialog.getByText('already registered')).toBeInTheDocument()
    })
    expect(emitted().saved).toBeFalsy()
  })

  describe('renaming a user', () => {
    const user = () => buildUser({ id: 'user-1', username: 'jane', email: 'jane@example.com' })

    function arrange({ userBasedPaths, shelves }: { userBasedPaths: boolean, shelves: ReturnType<typeof buildShelf>[] }) {
      useState<SettingPageBody | null>('settings').value = buildSettingPageBody({ userBasedPaths })
      const shelfStore = useShelfStore()
      shelfStore.shelves = shelves
      shelfStore.loaded = true
    }

    async function openAndRename(username: string) {
      const { baseElement } = await renderSuspended(UserFormDialog, { props: { mode: 'edit', user: user(), open: true } })
      const dialog = within(baseElement as HTMLElement)
      await fireEvent.update(dialog.getByLabelText('Username'), username)
      return dialog
    }

    it('warns how many of the user\'s shelf URLs change', async () => {
      arrange({
        userBasedPaths: true,
        shelves: [
          buildShelf({ id: 's1', userId: 'user-1', path: 'one' }),
          buildShelf({ id: 's2', userId: 'user-1', path: 'two' }),
          buildShelf({ id: 's3', userId: 'user-2', path: 'someone-else' }),
          buildShelf({ id: 's4', userId: 'user-1', path: '' })
        ]
      })

      const dialog = await openAndRename('janet')

      await waitFor(() => {
        expect(dialog.getByText(/public URL of this user's 2 shelves/)).toBeInTheDocument()
      })
    })

    it('uses the singular for one shelf', async () => {
      arrange({ userBasedPaths: true, shelves: [buildShelf({ id: 's1', userId: 'user-1', path: 'one' })] })

      const dialog = await openAndRename('janet')

      await waitFor(() => {
        expect(dialog.getByText(/public URL of this user's shelf\./)).toBeInTheDocument()
      })
    })

    it('does not warn while user-based paths are off', async () => {
      arrange({ userBasedPaths: false, shelves: [buildShelf({ id: 's1', userId: 'user-1', path: 'one' })] })

      const dialog = await openAndRename('janet')

      expect(dialog.queryByText(/Links shared with the old username stop working/)).not.toBeInTheDocument()
    })

    it('does not warn when the username is unchanged', async () => {
      arrange({ userBasedPaths: true, shelves: [buildShelf({ id: 's1', userId: 'user-1', path: 'one' })] })

      const dialog = await openAndRename('jane')

      expect(dialog.queryByText(/Links shared with the old username stop working/)).not.toBeInTheDocument()
    })

    it('does not warn for a user without shelves', async () => {
      arrange({ userBasedPaths: true, shelves: [] })

      const dialog = await openAndRename('janet')

      expect(dialog.queryByText(/Links shared with the old username stop working/)).not.toBeInTheDocument()
    })

    it('does not warn when creating a user', async () => {
      arrange({ userBasedPaths: true, shelves: [buildShelf({ id: 's1', userId: 'user-1', path: 'one' })] })

      const { baseElement } = await renderSuspended(UserFormDialog, { props: { mode: 'create', open: true } })

      expect(within(baseElement as HTMLElement).queryByText(/Links shared with the old username stop working/)).not.toBeInTheDocument()
    })
  })
})
