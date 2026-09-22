import { mountSuspended, renderSuspended } from '@nuxt/test-utils/runtime'
import { fireEvent, screen, waitFor, within } from '@testing-library/vue'
import { HttpResponse, http } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'
import draggable from 'vuedraggable'
import { LinkOrderResponseBodyToJSON, LinkToJSON, SectionOrderResponseBodyToJSON, SectionToJSON, ShelfToJSON } from '~~/api'
import type { SettingPageBody } from '~~/api'
import { server } from '../../../../test/mocks/server'
import { buildLink, buildSection, buildSettingPageBody, buildShelf, buildTokenPair } from '../../../../test/mocks/factories'
import ShelfDetailPage from './[shelfId].vue'

const BASE = 'http://localhost:8085'
// The instance's own origin, which a path shelf's public URL starts with.
const ORIGIN = window.location.origin

function mockShelfDetail() {
  server.use(
    http.get(`${BASE}/v1/shelves/shelf-1`, () => HttpResponse.json(ShelfToJSON(buildShelf({
      id: 'shelf-1',
      title: 'My Shelf',
      description: 'A shelf',
      path: 'my-shelf',
      domain: ''
    })))),
    http.get(`${BASE}/v1/sections`, () => HttpResponse.json([])),
    http.get(`${BASE}/v1/links`, () => HttpResponse.json([]))
  )
}

beforeEach(() => {
  // The "/app/*" global auth middleware runs on any real navigation
  // (including the `route:` option below, since it's an actual router
  // navigation) and redirects to /auth/sign-in when unauthenticated.
  useAuthStore().setTokens(buildTokenPair())
  useState<SettingPageBody | null>('settings').value = buildSettingPageBody({ userBasedPaths: false })
  mockShelfDetail()
})

// Nuxt UI's alert has no ARIA role, so "no alert" is checked by its titles.
function expectNoUrlAlert() {
  for (const title of ['This shelf\'s URL now includes your username', 'This shelf\'s URL no longer includes a username', 'This path can\'t be reached']) {
    expect(screen.queryByText(title)).not.toBeInTheDocument()
  }
}

function mockShelfWithPath(path: string, { username = 'alice', createdWithUserBasedPaths = false } = {}) {
  server.use(http.get(`${BASE}/v1/shelves/shelf-1`, () => HttpResponse.json(ShelfToJSON(buildShelf({
    id: 'shelf-1',
    title: 'My Shelf',
    description: 'A shelf',
    path,
    username,
    createdWithUserBasedPaths,
    domain: ''
  })))))
}

describe('shelf detail page', () => {
  it('shows the shelf title, description and path once loaded', async () => {
    await renderSuspended(ShelfDetailPage, { route: '/app/shelf/shelf-1' })

    await waitFor(() => {
      expect(screen.getByText('My Shelf')).toBeInTheDocument()
    })
    expect(screen.getByText('A shelf')).toBeInTheDocument()
    expect(screen.getByText(`${ORIGIN}/my-shelf`)).toBeInTheDocument()
  })

  describe('public URL and the alert at the top', () => {
    it('shows no alert for an ordinary path while user-based paths are off', async () => {
      await renderSuspended(ShelfDetailPage, { route: '/app/shelf/shelf-1' })

      await waitFor(() => {
        expect(screen.getByText('My Shelf')).toBeInTheDocument()
      })
      expectNoUrlAlert()
      expect(screen.getByRole('link', { name: 'Open' })).toHaveAttribute('href', `${ORIGIN}/my-shelf`)
    })

    it('puts the owner\'s username in the URL and explains the change for a shelf created before user-based paths were on', async () => {
      useState<SettingPageBody | null>('settings').value = buildSettingPageBody({ userBasedPaths: true })
      mockShelfWithPath('my-shelf', { username: 'alice', createdWithUserBasedPaths: false })

      await renderSuspended(ShelfDetailPage, { route: '/app/shelf/shelf-1' })

      await waitFor(() => {
        expect(screen.getByText('This shelf\'s URL now includes your username')).toBeInTheDocument()
      })
      expect(screen.getByText(`${ORIGIN}/alice/my-shelf`)).toBeInTheDocument()
      expect(screen.getByText(/Links without the username, like \/my-shelf, no longer work/)).toBeInTheDocument()
      expect(screen.getByRole('link', { name: 'Open' })).toHaveAttribute('href', `${ORIGIN}/alice/my-shelf`)
    })

    it('shows no alert for a shelf created while user-based paths were already on', async () => {
      useState<SettingPageBody | null>('settings').value = buildSettingPageBody({ userBasedPaths: true })
      mockShelfWithPath('my-shelf', { username: 'alice', createdWithUserBasedPaths: true })

      await renderSuspended(ShelfDetailPage, { route: '/app/shelf/shelf-1' })

      await waitFor(() => {
        expect(screen.getByText(`${ORIGIN}/alice/my-shelf`)).toBeInTheDocument()
      })
      expectNoUrlAlert()
      // The URL itself still has the username in it.
      expect(screen.getByRole('link', { name: 'Open' })).toHaveAttribute('href', `${ORIGIN}/alice/my-shelf`)
    })

    it('warns that a reserved path cannot be reached', async () => {
      mockShelfWithPath('docs')

      await renderSuspended(ShelfDetailPage, { route: '/app/shelf/shelf-1' })

      await waitFor(() => {
        expect(screen.getByText('This path can\'t be reached')).toBeInTheDocument()
      })
    })

    it('shows a domain shelf at its own domain, without a path', async () => {
      server.use(http.get(`${BASE}/v1/shelves/shelf-1`, () => HttpResponse.json(ShelfToJSON(buildShelf({
        id: 'shelf-1',
        title: 'My Shelf',
        description: 'A shelf',
        path: '',
        domain: 'profile.example.com:9443'
      })))))

      await renderSuspended(ShelfDetailPage, { route: '/app/shelf/shelf-1' })

      await waitFor(() => {
        expect(screen.getByText('https://profile.example.com:9443')).toBeInTheDocument()
      })
      expectNoUrlAlert()
      const open = screen.getByRole('link', { name: 'Open' })
      expect(open).toHaveAttribute('href', 'https://profile.example.com:9443')
      expect(open).toHaveAttribute('target', '_blank')
    })

    it('shows the alert above the title', async () => {
      mockShelfWithPath('docs')

      await renderSuspended(ShelfDetailPage, { route: '/app/shelf/shelf-1' })

      const alert = await screen.findByText('This path can\'t be reached')
      const title = screen.getByRole('heading', { name: 'My Shelf' })
      expect(alert.compareDocumentPosition(title) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
    })
  })

  it('shows a not-found state when the shelf cannot be loaded', async () => {
    server.use(http.get(`${BASE}/v1/shelves/shelf-1`, () => new HttpResponse(null, { status: 404 })))

    await renderSuspended(ShelfDetailPage, { route: '/app/shelf/shelf-1' })

    await waitFor(() => {
      expect(screen.getByText('This shelf could not be found.')).toBeInTheDocument()
    })
    expect(screen.getByRole('link', { name: 'Back to shelves' })).toHaveAttribute('href', '/app/shelf')
  })

  it('shows an empty-sections prompt when the shelf has no sections', async () => {
    await renderSuspended(ShelfDetailPage, { route: '/app/shelf/shelf-1' })

    await waitFor(() => {
      expect(screen.getByText('This shelf doesn\'t have any sections yet.')).toBeInTheDocument()
    })
  })

  it('renders each section with its links', async () => {
    server.use(
      http.get(`${BASE}/v1/sections`, () => HttpResponse.json([
        SectionToJSON(buildSection({ id: 'section-1', shelfId: 'shelf-1', title: 'First Section', order: 0 }))
      ])),
      http.get(`${BASE}/v1/links`, () => HttpResponse.json([
        LinkToJSON(buildLink({ id: 'link-1', sectionId: 'section-1', title: 'A Link', link: 'https://example.com', order: 0 }))
      ]))
    )

    await renderSuspended(ShelfDetailPage, { route: '/app/shelf/shelf-1' })

    await waitFor(() => {
      expect(screen.getByText('First Section')).toBeInTheDocument()
    })
    expect(screen.getByText('A Link')).toBeInTheDocument()
  })

  it('creates a new section', async () => {
    let receivedBody: unknown
    server.use(http.post(`${BASE}/v1/sections`, async ({ request }) => {
      receivedBody = await request.json()
      return HttpResponse.json(SectionToJSON(buildSection({ id: 'section-new', shelfId: 'shelf-1', title: 'New Section', order: 0 })))
    }))

    await renderSuspended(ShelfDetailPage, { route: '/app/shelf/shelf-1' })

    await waitFor(() => {
      expect(screen.getByText('My Shelf')).toBeInTheDocument()
    })

    await fireEvent.update(screen.getByPlaceholderText('Section title'), 'New Section')
    await fireEvent.click(screen.getByRole('button', { name: 'New section' }))

    await waitFor(() => {
      expect(receivedBody).toMatchObject({ title: 'New Section', shelfId: 'shelf-1' })
    })
  })

  it('deletes the shelf and navigates back to the shelf list after confirming', async () => {
    let deleteCalled = false
    server.use(http.delete(`${BASE}/v1/shelves/shelf-1`, () => {
      deleteCalled = true
      return new HttpResponse(null, { status: 204 })
    }))

    const { baseElement, container } = await renderSuspended(ShelfDetailPage, { route: '/app/shelf/shelf-1' })

    await waitFor(() => {
      expect(screen.getByText('My Shelf')).toBeInTheDocument()
    })

    // The page's own "Delete" trigger always renders before the (closed)
    // confirm dialog's own "Delete" button in DOM order, so take the first
    // match rather than risk an ambiguous getByRole.
    // The page's own outline "Delete" trigger stays mounted even once the
    // dialog opens (unlike a dropdown menu item, which closes), so scoping
    // just to `container` still leaves both it and the dialog's own solid
    // "Delete" button matching - scope precisely by role first instead.
    await fireEvent.click(within(container as HTMLElement).getByRole('button', { name: 'Delete' }))

    const dialog = await within(baseElement as HTMLElement).findByRole('dialog', { hidden: true })
    const confirmButton = within(dialog).getByRole('button', { name: 'Delete' })
    await fireEvent.click(confirmButton)

    await waitFor(() => {
      expect(deleteCalled).toBe(true)
    })
  })

  it('updates the displayed title after editing the shelf', async () => {
    server.use(http.put(`${BASE}/v1/shelves/shelf-1`, () => HttpResponse.json(ShelfToJSON(buildShelf({
      id: 'shelf-1',
      title: 'Renamed Shelf',
      path: 'my-shelf',
      domain: ''
    })))))

    const { baseElement } = await renderSuspended(ShelfDetailPage, { route: '/app/shelf/shelf-1' })

    await waitFor(() => {
      expect(screen.getByText('My Shelf')).toBeInTheDocument()
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Edit' }))

    const scope = within(baseElement as HTMLElement)
    await fireEvent.update(scope.getByLabelText('Title'), 'Renamed Shelf')
    await fireEvent.click(scope.getByRole('button', { name: 'Submit' }))

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Renamed Shelf' })).toBeInTheDocument()
    })
  })

  it('enables and runs "Save order" once a drag reorders sections, then disables it again', async () => {
    server.use(
      http.get(`${BASE}/v1/sections`, () => HttpResponse.json([
        SectionToJSON(buildSection({ id: 'section-1', shelfId: 'shelf-1', title: 'Section A', order: 0 })),
        SectionToJSON(buildSection({ id: 'section-2', shelfId: 'shelf-1', title: 'Section B', order: 1 }))
      ])),
      http.get(`${BASE}/v1/links`, () => HttpResponse.json([
        LinkToJSON(buildLink({ id: 'link-1', sectionId: 'section-1', title: 'A Link', order: 0 }))
      ]))
    )
    let sectionsReorderCalled = false
    let linksReorderCalled = false
    server.use(
      http.put(`${BASE}/v1/sections/reorder`, () => {
        sectionsReorderCalled = true
        return HttpResponse.json(SectionOrderResponseBodyToJSON({ failures: [] }))
      }),
      http.put(`${BASE}/v1/links/reorder`, () => {
        linksReorderCalled = true
        return HttpResponse.json(LinkOrderResponseBodyToJSON({ failures: [] }))
      })
    )

    // mountSuspended (not renderSuspended) so wrapper.findComponent() can
    // reach the draggable list directly - simulating a real HTML5 drag
    // gesture is unreliable in jsdom, so the drag-end event is triggered
    // directly instead. Its result isn't attached to document.body, so
    // queries are scoped to wrapper.element rather than the global screen.
    // The page's template has no single root (separate v-if/v-else-if
    // branches for loading/not-found/loaded), so wrapper.element itself
    // changes once loading finishes - re-derive the scope on each use
    // rather than caching it, or it keeps pointing at the discarded skeleton.
    const wrapper = await mountSuspended(ShelfDetailPage, { route: '/app/shelf/shelf-1' })
    // waitFor's MutationObserver watches document.body by default, but this
    // tree is detached from it - fall back to a plain wait for the initial
    // shelf/section/link fetches to resolve instead of polling for them.
    await new Promise(resolve => setTimeout(resolve, 300))

    // Not .toBeInTheDocument() - this tree is intentionally detached from
    // window.document (see the note above), which that matcher requires.
    expect(within(wrapper.element as HTMLElement).getByText('Section A')).toBeTruthy()

    const saveOrderButton = within(wrapper.element as HTMLElement).getByRole('button', { name: 'Save order' })
    expect(saveOrderButton).toBeDisabled()

    await wrapper.findComponent(draggable).vm.$emit('end')
    await waitFor(() => {
      expect(saveOrderButton).not.toBeDisabled()
    })

    await fireEvent.click(saveOrderButton)

    await waitFor(() => {
      expect(sectionsReorderCalled).toBe(true)
    })
    expect(linksReorderCalled).toBe(true)
    await waitFor(() => {
      expect(saveOrderButton).toBeDisabled()
    })
  })
})
