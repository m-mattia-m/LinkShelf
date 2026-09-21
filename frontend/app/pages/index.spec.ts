import { renderSuspended } from '@nuxt/test-utils/runtime'
import { screen, waitFor } from '@testing-library/vue'
import { HttpResponse, http } from 'msw'
import { afterEach, describe, expect, it } from 'vitest'
import { LinkToJSON, PublicShelfToJSON, SectionToJSON } from '~~/api'
import { server } from '../../test/mocks/server'
import { buildLink, buildSection } from '../../test/mocks/factories'
import IndexPage from './index.vue'

const BASE = 'http://localhost:8085'

describe('landing page', () => {
  it('renders the welcome title and description', async () => {
    await renderSuspended(IndexPage)

    expect(screen.getByText('Getting started sharing links.')).toBeInTheDocument()
    expect(screen.getByText(/LinkShelf is an Open Source Linktree alternative/)).toBeInTheDocument()
  })

  it('links "Get started" to the app and "Source code" to GitHub', async () => {
    await renderSuspended(IndexPage)

    expect(screen.getByRole('link', { name: /Get started/ })).toHaveAttribute('href', '/app')
    expect(screen.getByRole('link', { name: /Source code/ })).toHaveAttribute('href', 'https://github.com/m-mattia-m/Linkshelf')
  })

  it('renders the features section', async () => {
    await renderSuspended(IndexPage)

    expect(screen.getByText('Everything you need to organize your links')).toBeInTheDocument()
    expect(screen.getByText('Unlimited collections')).toBeInTheDocument()
    expect(screen.getByText('Custom domains')).toBeInTheDocument()
    expect(screen.getByText('Self-host & scale')).toBeInTheDocument()
  })

  it('renders the closing call-to-action with cloud and self-host links', async () => {
    await renderSuspended(IndexPage)

    expect(screen.getByText('Ready to organize your links?')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Start with LinkShelf Cloud/ })).toHaveAttribute('href', '/cloud')
    expect(screen.getByRole('link', { name: /Self-host on GitHub/ })).toHaveAttribute('href', 'https://github.com/m-mattia-m/linkshelf')
  })
})

// When the frontend is reached on the domain of a shelf, "/" shows that shelf
// instead. middleware/shelf-host.global.ts decides that and hands the domain
// over as the 'shelf-host' state.
describe('landing page on the domain of a shelf', () => {
  afterEach(() => {
    clearNuxtState('shelf-host')
  })

  function serveShelfOn(domain: string, onRequest?: (domain: string) => void) {
    server.use(
      http.get(`${BASE}/v1/shelves/by-domain/:domain`, ({ params }) => {
        onRequest?.(params.domain as string)
        return HttpResponse.json(PublicShelfToJSON({
          id: 'shelf-1', title: 'Profile shelf', description: `Served on ${domain}`, icon: '', path: '', theme: {}
        } as never))
      }),
      http.get(`${BASE}/v1/sections`, () => HttpResponse.json([
        SectionToJSON(buildSection({ id: 'section-1', shelfId: 'shelf-1', title: 'Section A', order: 0 }))
      ])),
      http.get(`${BASE}/v1/links`, () => HttpResponse.json([
        LinkToJSON(buildLink({ id: 'link-1', sectionId: 'section-1', title: 'A Link', link: 'https://example.com', order: 0 }))
      ]))
    )
  }

  it('shows the shelf of that domain instead of the landing page', async () => {
    let requestedDomain: string | undefined
    serveShelfOn('profile.example.com', (domain) => {
      requestedDomain = domain
    })
    useState<string | null>('shelf-host').value = 'profile.example.com'

    await renderSuspended(IndexPage)

    await waitFor(() => {
      expect(screen.getByText('Profile shelf')).toBeInTheDocument()
    })
    expect(requestedDomain).toBe('profile.example.com')
    expect(screen.getByText('Served on profile.example.com')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /A Link/ })).toHaveAttribute('href', 'https://example.com')
    expect(screen.queryByText('Getting started sharing links.')).not.toBeInTheDocument()
  })

  it('asks for the domain including its port', async () => {
    let requestedDomain: string | undefined
    serveShelfOn('profile.example.com:9443', (domain) => {
      requestedDomain = domain
    })
    useState<string | null>('shelf-host').value = 'profile.example.com:9443'

    await renderSuspended(IndexPage)

    await waitFor(() => {
      expect(screen.getByText('Profile shelf')).toBeInTheDocument()
    })
    expect(requestedDomain).toBe('profile.example.com:9443')
  })

  it('shows the not-found page when the shelf is gone by the time it is loaded', async () => {
    server.use(http.get(`${BASE}/v1/shelves/by-domain/:domain`, () => new HttpResponse(null, { status: 404 })))
    useState<string | null>('shelf-host').value = 'profile.example.com'

    await renderSuspended(IndexPage)

    await waitFor(() => {
      expect(screen.getByText('This page doesn\'t exist.')).toBeInTheDocument()
    })
  })

  it('still shows the landing page on the instance\'s own host', async () => {
    useState<string | null>('shelf-host').value = null

    await renderSuspended(IndexPage)

    expect(screen.getByText('Getting started sharing links.')).toBeInTheDocument()
  })
})
