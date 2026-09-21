import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { fireEvent, screen, waitFor, within } from '@testing-library/vue'
import { beforeEach, describe, expect, it } from 'vitest'
import { ShelfToJSON, UserToJSON } from '~~/api'
import type { SettingPageBody } from '~~/api'
import { errorResponse } from '../../../test/mocks/handlers'
import { server } from '../../../test/mocks/server'
import { buildSettingPageBody, buildShelf, buildUser } from '../../../test/mocks/factories'
import { useShelfStore } from '~/stores/shelf'
import ProfilePage from './profile.vue'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  useShelfStore().$reset()
  useState<SettingPageBody | null>('settings').value = buildSettingPageBody({ userBasedPaths: false })
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
    expect(screen.getByLabelText('Username')).toHaveValue('user-1')
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
      expect(putBody).toEqual({ email: 'jane@example.com', username: 'user-1', first_name: 'Janet', last_name: 'Doe', role: undefined })
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

  describe('renaming the username', () => {
    const TOKEN = 'header.eyJzdWIiOiJ1c2VyLTEiLCJyb2xlIjoidXNlciIsImV4cCI6OTk5OTk5OTk5OX0.signature'

    // Serves the profile, the user's shelves, and records what gets saved.
    function arrange({ userBasedPaths, shelves }: { userBasedPaths: boolean, shelves: ReturnType<typeof buildShelf>[] }) {
      useState<SettingPageBody | null>('settings').value = buildSettingPageBody({ userBasedPaths })
      useAuthStore().setTokens({ accessToken: TOKEN, refreshToken: 'refresh-1' })
      const saved: unknown[] = []
      server.use(
        http.get(`${BASE}/v1/users/me`, () => HttpResponse.json(UserToJSON(buildUser({ id: 'user-1', username: 'jane' })))),
        http.get(`${BASE}/v1/shelves`, () => HttpResponse.json(shelves.map(ShelfToJSON))),
        http.put(`${BASE}/v1/users/user-1`, async ({ request }) => {
          const body = await request.json() as { username: string }
          saved.push(body)
          return HttpResponse.json(UserToJSON(buildUser({ id: 'user-1', username: body.username })))
        })
      )
      return saved
    }

    async function renameTo(username: string) {
      await waitFor(() => {
        expect(screen.getByLabelText('Username')).toHaveValue('jane')
      })
      await fireEvent.update(screen.getByLabelText('Username'), username)
      await fireEvent.click(screen.getByRole('button', { name: 'Save' }))
    }

    it('asks for confirmation first, with the number of affected shelves, before changing the URLs', async () => {
      const saved = arrange({
        userBasedPaths: true,
        shelves: [
          buildShelf({ id: 's1', userId: 'user-1', path: 'one' }),
          buildShelf({ id: 's2', userId: 'user-1', path: 'two' }),
          // Someone else's shelf, and one without a path, must not be counted.
          buildShelf({ id: 's3', userId: 'user-2', path: 'three' }),
          buildShelf({ id: 's4', userId: 'user-1', path: '' })
        ]
      })

      const { baseElement } = await renderSuspended(ProfilePage)
      await renameTo('janet')

      const dialog = within(baseElement as HTMLElement)
      await waitFor(() => {
        expect(dialog.getByText('Change your username?')).toBeInTheDocument()
      })
      expect(dialog.getByText(/your 2 shelves to \/janet\//)).toBeInTheDocument()
      expect(saved).toEqual([])

      await fireEvent.click(dialog.getByRole('button', { name: 'Change username' }))

      await waitFor(() => {
        expect(saved).toHaveLength(1)
      })
      expect(saved[0]).toMatchObject({ username: 'janet' })
      expect(useAuthStore().user?.username).toBe('janet')
    })

    it('uses the singular wording for one shelf', async () => {
      arrange({ userBasedPaths: true, shelves: [buildShelf({ id: 's1', userId: 'user-1', path: 'one' })] })

      const { baseElement } = await renderSuspended(ProfilePage)
      await renameTo('janet')

      await waitFor(() => {
        expect(within(baseElement as HTMLElement).getByText(/of your shelf to \/janet\//)).toBeInTheDocument()
      })
    })

    it('saves nothing when the confirmation is cancelled', async () => {
      const saved = arrange({ userBasedPaths: true, shelves: [buildShelf({ id: 's1', userId: 'user-1', path: 'one' })] })

      const { baseElement } = await renderSuspended(ProfilePage)
      await renameTo('janet')

      const dialog = within(baseElement as HTMLElement)
      await waitFor(() => {
        expect(dialog.getByText('Change your username?')).toBeInTheDocument()
      })
      await fireEvent.click(dialog.getByRole('button', { name: 'Cancel' }))

      await waitFor(() => {
        expect(dialog.queryByText('Change your username?')).not.toBeInTheDocument()
      })
      expect(saved).toEqual([])
    })

    it('saves straight away while user-based paths are off, since no URL changes', async () => {
      const saved = arrange({ userBasedPaths: false, shelves: [buildShelf({ id: 's1', userId: 'user-1', path: 'one' })] })

      const { baseElement } = await renderSuspended(ProfilePage)
      await renameTo('janet')

      await waitFor(() => {
        expect(saved).toHaveLength(1)
      })
      expect(within(baseElement as HTMLElement).queryByText('Change your username?')).not.toBeInTheDocument()
    })

    it('saves straight away when there is no shelf whose URL could break', async () => {
      const saved = arrange({ userBasedPaths: true, shelves: [buildShelf({ id: 's1', userId: 'user-2', path: 'not-mine' })] })

      const { baseElement } = await renderSuspended(ProfilePage)
      await renameTo('janet')

      await waitFor(() => {
        expect(saved).toHaveLength(1)
      })
      expect(within(baseElement as HTMLElement).queryByText('Change your username?')).not.toBeInTheDocument()
    })

    it('does not ask when the username is unchanged', async () => {
      const saved = arrange({ userBasedPaths: true, shelves: [buildShelf({ id: 's1', userId: 'user-1', path: 'one' })] })

      const { baseElement } = await renderSuspended(ProfilePage)
      await waitFor(() => {
        expect(screen.getByLabelText('Username')).toHaveValue('jane')
      })
      await fireEvent.update(screen.getByLabelText('First name'), 'Janet')
      await fireEvent.click(screen.getByRole('button', { name: 'Save' }))

      await waitFor(() => {
        expect(saved).toHaveLength(1)
      })
      expect(within(baseElement as HTMLElement).queryByText('Change your username?')).not.toBeInTheDocument()
    })

    it('validates the username format before asking or saving', async () => {
      const saved = arrange({ userBasedPaths: true, shelves: [buildShelf({ id: 's1', userId: 'user-1', path: 'one' })] })

      const { baseElement } = await renderSuspended(ProfilePage)
      await renameTo('Not Valid')

      await waitFor(() => {
        expect(screen.getByText(/Lowercase letters, numbers and hyphens only/)).toBeInTheDocument()
      })
      expect(within(baseElement as HTMLElement).queryByText('Change your username?')).not.toBeInTheDocument()
      expect(saved).toEqual([])
    })
  })
})
