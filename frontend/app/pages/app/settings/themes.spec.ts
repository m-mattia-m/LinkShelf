import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { fireEvent, screen, waitFor, within } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import { ThemeToJSON } from '~~/api'
import { errorResponse } from '../../../../test/mocks/handlers'
import { server } from '../../../../test/mocks/server'
import { buildTheme } from '../../../../test/mocks/factories'
import ThemesSettingsPage from './themes.vue'

const BASE = 'http://localhost:8085'

describe('app settings themes page', () => {
  it('shows a loading state, then the list of user-created themes', async () => {
    server.use(http.get(`${BASE}/v1/themes/admin`, () => HttpResponse.json([
      buildTheme({ id: 'theme-1', name: 'Vivid', ownerUserId: 'user-2' })
    ].map(ThemeToJSON))))

    await renderSuspended(ThemesSettingsPage)

    await waitFor(() => {
      expect(screen.getByText('Vivid')).toBeInTheDocument()
    })
    expect(screen.getByText('Owner: user-2')).toBeInTheDocument()
  })

  it('shows an empty-state message when there are no user-created themes', async () => {
    server.use(http.get(`${BASE}/v1/themes/admin`, () => HttpResponse.json([])))

    await renderSuspended(ThemesSettingsPage)

    await waitFor(() => {
      expect(screen.getByText('No user-created themes yet.')).toBeInTheDocument()
    })
  })

  it('deletes a theme after the user confirms the destructive dialog', async () => {
    let deleteCalled = false
    server.use(
      http.get(`${BASE}/v1/themes/admin`, () => HttpResponse.json([
        buildTheme({ id: 'theme-1', name: 'Deletable', ownerUserId: 'user-2' })
      ].map(ThemeToJSON))),
      http.delete(`${BASE}/v1/themes/theme-1`, () => {
        deleteCalled = true
        return new HttpResponse(null, { status: 204 })
      })
    )

    const { baseElement } = await renderSuspended(ThemesSettingsPage)

    await waitFor(() => {
      expect(screen.getByText('Deletable')).toBeInTheDocument()
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Delete theme' }))

    const dialog = within(baseElement as HTMLElement)
    const confirmButton = await dialog.findByRole('button', { name: /^Delete$/, hidden: true })
    await fireEvent.click(confirmButton)

    await waitFor(() => {
      expect(deleteCalled).toBe(true)
    })
  })

  it('shows the API error and keeps the theme listed when deletion fails', async () => {
    server.use(
      http.get(`${BASE}/v1/themes/admin`, () => HttpResponse.json([
        buildTheme({ id: 'theme-1', name: 'Sticky', ownerUserId: 'user-2' })
      ].map(ThemeToJSON))),
      http.delete(`${BASE}/v1/themes/theme-1`, () => errorResponse(500, 'boom'))
    )

    const { baseElement } = await renderSuspended(ThemesSettingsPage)

    await waitFor(() => {
      expect(screen.getByText('Sticky')).toBeInTheDocument()
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Delete theme' }))
    const dialog = within(baseElement as HTMLElement)
    const confirmButton = await dialog.findByRole('button', { name: /^Delete$/, hidden: true })
    await fireEvent.click(confirmButton)

    await waitFor(() => {
      expect(confirmButton).not.toHaveAttribute('aria-disabled', 'true')
    })
    expect(screen.getByText('Sticky')).toBeInTheDocument()
  })
})
