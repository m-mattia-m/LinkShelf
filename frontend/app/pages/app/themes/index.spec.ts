import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { screen, waitFor, within } from '@testing-library/vue'
import { beforeEach, describe, expect, it } from 'vitest'
import { ThemeGroupedResponseBodyToJSON } from '~~/api'
import { server } from '../../../../test/mocks/server'
import { errorResponse } from '../../../../test/mocks/handlers'
import { buildTheme } from '../../../../test/mocks/factories'
import { resetOnceCache } from '../../../../test/reset-once-cache'
import ThemesIndexPage from './index.vue'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  resetOnceCache()
})

describe('themes index page', () => {
  it('shows a loading state, then the fetched instance and personal themes', async () => {
    server.use(http.get(`${BASE}/v1/themes`, () => HttpResponse.json(ThemeGroupedResponseBodyToJSON({
      instance: [buildTheme({ id: 'instance-theme', name: 'Instance Look' })],
      mine: [buildTheme({ id: 'mine-theme', name: 'My Custom Look' })]
    }))))

    await renderSuspended(ThemesIndexPage)

    await waitFor(() => {
      expect(screen.getByText('Instance Look')).toBeInTheDocument()
    })

    expect(screen.getByText('My Custom Look')).toBeInTheDocument()
    expect(screen.getByText('Instance themes')).toBeInTheDocument()
  })

  it('shows an empty-state prompt when the user has no themes of their own', async () => {
    server.use(http.get(`${BASE}/v1/themes`, () => HttpResponse.json(ThemeGroupedResponseBodyToJSON({ instance: [], mine: [] }))))

    await renderSuspended(ThemesIndexPage)

    await waitFor(() => {
      expect(screen.getByText('You haven\'t created any themes yet.')).toBeInTheDocument()
    })
  })

  it('deletes a theme after the user confirms the destructive dialog', async () => {
    let deleteCalled = false
    server.use(
      http.get(`${BASE}/v1/themes`, () => HttpResponse.json(ThemeGroupedResponseBodyToJSON({
        instance: [],
        mine: [buildTheme({ id: 'theme-1', name: 'Deletable Theme' })]
      }))),
      http.delete(`${BASE}/v1/themes/theme-1`, () => {
        deleteCalled = true
        return new HttpResponse(null, { status: 204 })
      })
    )

    const { baseElement } = await renderSuspended(ThemesIndexPage)

    await waitFor(() => {
      expect(screen.getByText('Deletable Theme')).toBeInTheDocument()
    })

    const actionsButton = screen.getByRole('button', { name: 'Actions' })
    await actionsButton.click()

    const deleteItem = await within(baseElement as HTMLElement).findByText('Delete')
    await deleteItem.click()

    const confirmButton = await within(baseElement as HTMLElement).findByRole('button', { name: /delete/i, hidden: true })
    await confirmButton.click()

    await waitFor(() => {
      expect(deleteCalled).toBe(true)
    })
  })

  it('stops showing the loading placeholder when the themes cannot be loaded', async () => {
    server.use(http.get(`${BASE}/v1/themes`, () => errorResponse(500, 'boom')))

    await renderSuspended(ThemesIndexPage)

    // The skeleton placeholder must not stay behind when the request fails.
    await waitFor(() => {
      expect(document.querySelector('.animate-pulse')).toBeNull()
    })
  })
})
