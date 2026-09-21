import { renderSuspended } from '@nuxt/test-utils/runtime'
import { fireEvent, screen, waitFor, within } from '@testing-library/vue'
import { HttpResponse, http } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'
import { ShelfToJSON } from '~~/api'
import { server } from '../../../../test/mocks/server'
import { errorResponse } from '../../../../test/mocks/handlers'
import { buildShelf } from '../../../../test/mocks/factories'
import { resetOnceCache } from '../../../../test/reset-once-cache'
import { useThemeStore } from '~/stores/theme'
import ShelfIndexPage from './index.vue'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  resetOnceCache()
  useThemeStore().loaded = true
})

describe('shelf index page', () => {
  it('shows a loading state, then the fetched shelves in the table', async () => {
    server.use(http.get(`${BASE}/v1/shelves`, () => HttpResponse.json([
      buildShelf({ id: 'shelf-1', title: 'Bookmarks', description: 'My links', path: 'bookmarks', domain: '' }),
      buildShelf({ id: 'shelf-2', title: 'Work', description: 'Work links', path: 'work', domain: '' })
    ].map(ShelfToJSON))))

    await renderSuspended(ShelfIndexPage)

    await waitFor(() => {
      expect(screen.getByText('Bookmarks')).toBeInTheDocument()
    })
    expect(screen.getByText('Work')).toBeInTheDocument()
  })

  it('shows an empty-state prompt when there are no shelves', async () => {
    server.use(http.get(`${BASE}/v1/shelves`, () => HttpResponse.json([])))

    await renderSuspended(ShelfIndexPage)

    await waitFor(() => {
      expect(screen.getByText('You don\'t have any shelves yet.')).toBeInTheDocument()
    })
  })

  it('links each shelf title to its detail page', async () => {
    server.use(http.get(`${BASE}/v1/shelves`, () => HttpResponse.json([
      buildShelf({ id: 'shelf-1', title: 'Bookmarks', path: 'bookmarks' })
    ].map(ShelfToJSON))))

    await renderSuspended(ShelfIndexPage)

    await waitFor(() => {
      expect(screen.getByRole('link', { name: 'Bookmarks' })).toHaveAttribute('href', '/app/shelf/shelf-1')
    })
  })

  describe('the "Open" action', () => {
    async function openActionFor(shelf: ReturnType<typeof buildShelf>) {
      server.use(http.get(`${BASE}/v1/shelves`, () => HttpResponse.json([shelf].map(ShelfToJSON))))

      const { baseElement } = await renderSuspended(ShelfIndexPage)
      await waitFor(() => {
        expect(screen.getByText(shelf.title)).toBeInTheDocument()
      })

      await screen.getByRole('button', { name: 'Actions' }).click()
      return within(baseElement as HTMLElement).findByRole('menuitem', { name: 'Open' })
    }

    it('opens the public page of a path shelf in a new tab', async () => {
      const item = await openActionFor(buildShelf({ id: 'shelf-1', title: 'Bookmarks', path: 'bookmarks', domain: '' }))

      expect(item).toHaveAttribute('href', `${window.location.origin}/bookmarks`)
      expect(item).toHaveAttribute('target', '_blank')
    })

    it('opens a domain shelf on its own domain', async () => {
      const item = await openActionFor(buildShelf({ id: 'shelf-1', title: 'Profile', path: '', domain: 'profile.example.com' }))

      expect(item).toHaveAttribute('href', 'https://profile.example.com')
      expect(item).toHaveAttribute('target', '_blank')
    })
  })

  it('deletes a shelf after the user confirms the destructive dialog', async () => {
    let deleteCalled = false
    server.use(
      http.get(`${BASE}/v1/shelves`, () => HttpResponse.json([
        buildShelf({ id: 'shelf-1', title: 'Deletable Shelf', path: 'deletable' })
      ].map(ShelfToJSON))),
      http.delete(`${BASE}/v1/shelves/shelf-1`, () => {
        deleteCalled = true
        return new HttpResponse(null, { status: 204 })
      })
    )

    const { baseElement } = await renderSuspended(ShelfIndexPage)

    await waitFor(() => {
      expect(screen.getByText('Deletable Shelf')).toBeInTheDocument()
    })

    const actionsButton = screen.getByRole('button', { name: 'Actions' })
    await actionsButton.click()

    const deleteItem = await within(baseElement as HTMLElement).findByText('Delete')
    await deleteItem.click()

    const confirmButton = await within(baseElement as HTMLElement).findByRole('button', { name: 'Delete', hidden: true })
    await confirmButton.click()

    await waitFor(() => {
      expect(deleteCalled).toBe(true)
    })
  })

  it('edits a shelf through the actions menu', async () => {
    let receivedShelfId: string | undefined
    server.use(
      http.get(`${BASE}/v1/shelves`, () => HttpResponse.json([
        buildShelf({ id: 'shelf-1', title: 'Old Title', path: 'old-path' })
      ].map(ShelfToJSON))),
      http.put(`${BASE}/v1/shelves/:shelfId`, ({ params }) => {
        receivedShelfId = params.shelfId as string
        return HttpResponse.json(ShelfToJSON(buildShelf({ id: 'shelf-1', title: 'New Title', path: 'old-path' })))
      })
    )

    const { baseElement } = await renderSuspended(ShelfIndexPage)

    await waitFor(() => {
      expect(screen.getByText('Old Title')).toBeInTheDocument()
    })

    const actionsButton = screen.getByRole('button', { name: 'Actions' })
    await actionsButton.click()

    const editItem = await within(baseElement as HTMLElement).findByText('Edit')
    await editItem.click()

    const scope = within(baseElement as HTMLElement)
    await fireEvent.update(scope.getByLabelText('Title'), 'New Title')
    await fireEvent.click(scope.getByRole('button', { name: 'Submit' }))

    await waitFor(() => {
      expect(receivedShelfId).toBe('shelf-1')
    })
  })

  it('stops showing the loading placeholder when the shelves cannot be loaded', async () => {
    server.use(http.get(`${BASE}/v1/shelves`, () => errorResponse(500, 'boom')))

    await renderSuspended(ShelfIndexPage)

    // The skeleton placeholder must not stay behind when the request fails.
    await waitFor(() => {
      expect(document.querySelector('.animate-pulse')).toBeNull()
    })
  })
})
