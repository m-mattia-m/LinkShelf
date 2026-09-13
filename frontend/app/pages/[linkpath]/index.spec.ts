import { renderSuspended } from '@nuxt/test-utils/runtime'
import { screen, waitFor } from '@testing-library/vue'
import { HttpResponse, http } from 'msw'
import { describe, expect, it } from 'vitest'
import { LinkToJSON, PublicShelfToJSON, SectionToJSON } from '~~/api'
import { server } from '../../../test/mocks/server'
import { buildLink, buildSection } from '../../../test/mocks/factories'
import LinkpathIndexPage from './index.vue'

const BASE = 'http://localhost:8085'

function buildPublicShelf(overrides: Partial<Record<string, unknown>> = {}) {
  return {
    id: 'shelf-1',
    path: 'my-shelf',
    title: 'My Shelf',
    description: 'A public shelf',
    icon: 'i-lucide-book',
    theme: {},
    ...overrides
  }
}

describe('public link page', () => {
  it('shows a loading state, then the shelf and its links once loaded', async () => {
    server.use(
      http.get(`${BASE}/v1/shelves/by-path/my-shelf`, () => HttpResponse.json(PublicShelfToJSON(buildPublicShelf() as never))),
      http.get(`${BASE}/v1/sections`, () => HttpResponse.json([
        SectionToJSON(buildSection({ id: 'section-1', shelfId: 'shelf-1', title: 'Section A', order: 0 }))
      ])),
      http.get(`${BASE}/v1/links`, () => HttpResponse.json([
        LinkToJSON(buildLink({ id: 'link-1', sectionId: 'section-1', title: 'A Link', link: 'https://example.com', order: 0 }))
      ]))
    )

    await renderSuspended(LinkpathIndexPage, { route: '/my-shelf' })

    await waitFor(() => {
      expect(screen.getByText('My Shelf')).toBeInTheDocument()
    })
    expect(screen.getByText('A public shelf')).toBeInTheDocument()
    expect(screen.getByText('Section A')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /A Link/ })).toHaveAttribute('href', 'https://example.com')
  })

  it('requests the shelf using the "linkpath" route param', async () => {
    let requestedPath: string | undefined
    server.use(
      http.get(`${BASE}/v1/shelves/by-path/:path`, ({ params }) => {
        requestedPath = params.path as string
        return HttpResponse.json(PublicShelfToJSON(buildPublicShelf({ path: params.path as string }) as never))
      }),
      http.get(`${BASE}/v1/sections`, () => HttpResponse.json([])),
      http.get(`${BASE}/v1/links`, () => HttpResponse.json([]))
    )

    await renderSuspended(LinkpathIndexPage, { route: '/some-other-shelf' })

    await waitFor(() => {
      expect(requestedPath).toBe('some-other-shelf')
    })
  })

  it('hides sections that have no links', async () => {
    server.use(
      http.get(`${BASE}/v1/shelves/by-path/my-shelf`, () => HttpResponse.json(PublicShelfToJSON(buildPublicShelf() as never))),
      http.get(`${BASE}/v1/sections`, () => HttpResponse.json([
        SectionToJSON(buildSection({ id: 'section-1', shelfId: 'shelf-1', title: 'Has Links', order: 0 })),
        SectionToJSON(buildSection({ id: 'section-2', shelfId: 'shelf-1', title: 'Empty Section', order: 1 }))
      ])),
      http.get(`${BASE}/v1/links`, () => HttpResponse.json([
        LinkToJSON(buildLink({ id: 'link-1', sectionId: 'section-1', title: 'A Link', order: 0 }))
      ]))
    )

    await renderSuspended(LinkpathIndexPage, { route: '/my-shelf' })

    await waitFor(() => {
      expect(screen.getByText('Has Links')).toBeInTheDocument()
    })
    expect(screen.queryByText('Empty Section')).not.toBeInTheDocument()
  })

  it('shows an empty-state message when the shelf has no links at all', async () => {
    server.use(
      http.get(`${BASE}/v1/shelves/by-path/my-shelf`, () => HttpResponse.json(PublicShelfToJSON(buildPublicShelf() as never))),
      http.get(`${BASE}/v1/sections`, () => HttpResponse.json([])),
      http.get(`${BASE}/v1/links`, () => HttpResponse.json([]))
    )

    await renderSuspended(LinkpathIndexPage, { route: '/my-shelf' })

    await waitFor(() => {
      expect(screen.getByText('Nothing here yet.')).toBeInTheDocument()
    })
  })

  it('shows a not-found state when no shelf exists at the given path', async () => {
    server.use(http.get(`${BASE}/v1/shelves/by-path/missing-shelf`, () => new HttpResponse(null, { status: 404 })))

    await renderSuspended(LinkpathIndexPage, { route: '/missing-shelf' })

    await waitFor(() => {
      expect(screen.getByText('This page doesn\'t exist.')).toBeInTheDocument()
    })
    expect(screen.getByText('The link you followed may be broken, or the page may have been removed.')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Go to LinkShelf' })).toHaveAttribute('href', '/')
  })

  it('applies a link\'s own color as its background when it isn\'t the DB default black', async () => {
    server.use(
      http.get(`${BASE}/v1/shelves/by-path/my-shelf`, () => HttpResponse.json(PublicShelfToJSON(buildPublicShelf() as never))),
      http.get(`${BASE}/v1/sections`, () => HttpResponse.json([
        SectionToJSON(buildSection({ id: 'section-1', shelfId: 'shelf-1', title: 'Section A', order: 0 }))
      ])),
      http.get(`${BASE}/v1/links`, () => HttpResponse.json([
        LinkToJSON(buildLink({ id: 'link-1', sectionId: 'section-1', title: 'Colored Link', color: '#588157', order: 0 })),
        LinkToJSON(buildLink({ id: 'link-2', sectionId: 'section-1', title: 'Default Link', color: '#000000', order: 1 }))
      ]))
    )

    await renderSuspended(LinkpathIndexPage, { route: '/my-shelf' })

    await waitFor(() => {
      expect(screen.getByRole('link', { name: /Colored Link/ })).toBeInTheDocument()
    })
    expect(screen.getByRole('link', { name: /Colored Link/ })).toHaveStyle({ backgroundColor: '#588157' })
    expect(screen.getByRole('link', { name: /Default Link/ })).not.toHaveStyle({ backgroundColor: '#000000' })
  })
})
