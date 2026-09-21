import { renderSuspended } from '@nuxt/test-utils/runtime'
import { screen, waitFor } from '@testing-library/vue'
import { HttpResponse, http } from 'msw'
import { describe, expect, it } from 'vitest'
import { LinkToJSON, PublicShelfToJSON, SectionToJSON } from '~~/api'
import { errorResponse } from '../../../test/mocks/handlers'
import { server } from '../../../test/mocks/server'
import { buildLink, buildSection } from '../../../test/mocks/factories'
import UsernamePathPage from './[linkpath].vue'

const BASE = 'http://localhost:8085'

function buildPublicShelf(overrides: Partial<Record<string, unknown>> = {}) {
  return {
    id: 'shelf-1',
    path: 'profile',
    title: 'Alice Profile',
    description: 'Alice links',
    icon: 'i-lucide-book',
    theme: {},
    ...overrides
  }
}

describe('public link page with a username', () => {
  it('looks the shelf up by username and path, then shows it with its links', async () => {
    let requested: { username?: string, path?: string } = {}
    server.use(
      http.get(`${BASE}/v1/shelves/by-user/:username/:path`, ({ params }) => {
        requested = { username: params.username as string, path: params.path as string }
        return HttpResponse.json(PublicShelfToJSON(buildPublicShelf() as never))
      }),
      http.get(`${BASE}/v1/sections`, () => HttpResponse.json([
        SectionToJSON(buildSection({ id: 'section-1', shelfId: 'shelf-1', title: 'Section A', order: 0 }))
      ])),
      http.get(`${BASE}/v1/links`, () => HttpResponse.json([
        LinkToJSON(buildLink({ id: 'link-1', sectionId: 'section-1', title: 'A Link', link: 'https://example.com', order: 0 }))
      ]))
    )

    await renderSuspended(UsernamePathPage, { route: '/alice/profile' })

    await waitFor(() => {
      expect(screen.getByText('Alice Profile')).toBeInTheDocument()
    })
    expect(requested).toEqual({ username: 'alice', path: 'profile' })
    expect(screen.getByRole('link', { name: /A Link/ })).toHaveAttribute('href', 'https://example.com')
  })

  it('never falls back to the path-only lookup', async () => {
    let pathOnlyCalled = false
    server.use(
      http.get(`${BASE}/v1/shelves/by-path/:path`, () => {
        pathOnlyCalled = true
        return HttpResponse.json(PublicShelfToJSON(buildPublicShelf() as never))
      }),
      http.get(`${BASE}/v1/shelves/by-user/:username/:path`, () => errorResponse(404, 'shelf not found'))
    )

    await renderSuspended(UsernamePathPage, { route: '/alice/profile' })

    await waitFor(() => {
      expect(screen.getByText('This page doesn\'t exist.')).toBeInTheDocument()
    })
    expect(pathOnlyCalled).toBe(false)
  })

  it('shows the not-found state when the username or the path does not exist', async () => {
    server.use(http.get(`${BASE}/v1/shelves/by-user/:username/:path`, () => errorResponse(404, 'shelf not found')))

    await renderSuspended(UsernamePathPage, { route: '/nobody/missing' })

    await waitFor(() => {
      expect(screen.getByText('This page doesn\'t exist.')).toBeInTheDocument()
    })
  })
})
