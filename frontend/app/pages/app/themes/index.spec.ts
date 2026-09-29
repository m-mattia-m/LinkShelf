import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { fireEvent, screen, waitFor, within } from '@testing-library/vue'
import { afterEach, beforeEach, describe, expect, it, onTestFinished, vi } from 'vitest'
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

afterEach(() => {
  vi.restoreAllMocks()
})

function serveMine(...themes: ReturnType<typeof buildTheme>[]) {
  server.use(http.get(`${BASE}/v1/themes`, () => HttpResponse.json(ThemeGroupedResponseBodyToJSON({ instance: [], mine: themes }))))
}

async function openActionsItem(baseElement: Element, themeName: string, item: string) {
  await waitFor(() => {
    expect(screen.getByText(themeName)).toBeInTheDocument()
  })
  await screen.getByRole('button', { name: 'Actions' }).click()
  const menuItem = await within(baseElement as HTMLElement).findByText(item)
  await menuItem.click()
}

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

  it('opens the edit dialog pre-filled with the chosen theme', async () => {
    serveMine(buildTheme({ id: 'theme-1', name: 'Editable Theme', config: '--shelf-text: #ffffff;' }))

    const { baseElement } = await renderSuspended(ThemesIndexPage)
    await openActionsItem(baseElement, 'Editable Theme', 'Edit')

    const dialog = await within(baseElement as HTMLElement).findByRole('dialog', { hidden: true })
    expect(within(dialog).getByText('Edit theme')).toBeInTheDocument()
    expect(within(dialog).getByDisplayValue('Editable Theme')).toBeInTheDocument()
    expect(within(dialog).getByDisplayValue('--shelf-text: #ffffff;')).toBeInTheDocument()
  })

  it('exports a theme as a plain-text file named after it', async () => {
    serveMine(buildTheme({ id: 'theme-1', name: 'My Dark Look!', config: '--shelf-bg: #000000;' }))
    const createObjectURL = vi.fn((_: Blob) => 'blob:theme')
    const revokeObjectURL = vi.fn()
    const original = { createObjectURL: URL.createObjectURL, revokeObjectURL: URL.revokeObjectURL }
    Object.assign(URL, { createObjectURL, revokeObjectURL })
    onTestFinished(() => {
      Object.assign(URL, original)
    })
    let downloaded: string | undefined
    vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(function (this: HTMLAnchorElement) {
      downloaded = this.download
    })

    const { baseElement } = await renderSuspended(ThemesIndexPage)
    await openActionsItem(baseElement, 'My Dark Look!', 'Export')

    await waitFor(() => {
      expect(downloaded).toBe('my-dark-look-.txt')
    })
    expect(await createObjectURL.mock.calls[0]![0].text()).toBe('--shelf-bg: #000000;')
    expect(revokeObjectURL).toHaveBeenCalledWith('blob:theme')
  })

  it('opens the create dialog pre-filled from an imported file', async () => {
    serveMine()

    const { baseElement, container } = await renderSuspended(ThemesIndexPage)
    await waitFor(() => {
      expect(screen.getByText('You haven\'t created any themes yet.')).toBeInTheDocument()
    })

    const input = container.querySelector('input[type="file"]') as HTMLInputElement
    const clickSpy = vi.spyOn(input, 'click').mockImplementation(() => {})
    await fireEvent.click(screen.getByRole('button', { name: 'Import' }))
    expect(clickSpy).toHaveBeenCalled()

    const file = new File(['--shelf-text: #123456;'], 'Imported Look.txt', { type: 'text/plain' })
    Object.defineProperty(input, 'files', { value: [file], configurable: true })
    await fireEvent.change(input)

    const dialog = await within(baseElement as HTMLElement).findByRole('dialog', { hidden: true })
    await waitFor(() => {
      expect(within(dialog).getByDisplayValue('Imported Look')).toBeInTheDocument()
    })
    expect(within(dialog).getByDisplayValue('--shelf-text: #123456;')).toBeInTheDocument()
  })

  it('ignores an import without a selected file', async () => {
    serveMine()

    const { baseElement, container } = await renderSuspended(ThemesIndexPage)
    await waitFor(() => {
      expect(screen.getByText('You haven\'t created any themes yet.')).toBeInTheDocument()
    })

    const input = container.querySelector('input[type="file"]') as HTMLInputElement
    Object.defineProperty(input, 'files', { value: [], configurable: true })
    await fireEvent.change(input)

    expect(within(baseElement as HTMLElement).queryByRole('dialog', { hidden: true })).toBeNull()
  })

  it('keeps the theme and the confirm dialog when deleting fails', async () => {
    serveMine(buildTheme({ id: 'theme-1', name: 'Stubborn Theme' }))
    let deleteCalled = false
    server.use(http.delete(`${BASE}/v1/themes/theme-1`, () => {
      deleteCalled = true
      return errorResponse(500, 'cannot delete')
    }))

    const { baseElement } = await renderSuspended(ThemesIndexPage)
    await openActionsItem(baseElement, 'Stubborn Theme', 'Delete')

    const confirmButton = await within(baseElement as HTMLElement).findByRole('button', { name: /delete/i, hidden: true })
    await confirmButton.click()

    await waitFor(() => {
      expect(deleteCalled).toBe(true)
    })
    await waitFor(() => {
      expect(confirmButton).not.toHaveAttribute('aria-busy', 'true')
    })
    expect(within(baseElement as HTMLElement).getByRole('dialog', { hidden: true })).toBeInTheDocument()
    expect(screen.getByText('Stubborn Theme')).toBeInTheDocument()
  })
})
