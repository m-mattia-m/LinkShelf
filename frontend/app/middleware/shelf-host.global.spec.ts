import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { PublicShelfToJSON } from '~~/api'
import { server } from '../../test/mocks/server'
import shelfHostMiddleware from './shelf-host.global'

const BASE = 'http://localhost:8085'

const { setPageLayoutMock } = vi.hoisted(() => ({ setPageLayoutMock: vi.fn() }))
mockNuxtImport('setPageLayout', () => setPageLayoutMock)

const originalUrl = window.location.href

function visit(url: string) {
  // happy-dom's own way of "navigating" the test window.
  (window as unknown as { happyDOM: { setURL: (url: string) => void } }).happyDOM.setURL(url)
}

function route(path: string) {
  return { path, fullPath: path } as never
}

function shelfHostState() {
  return useState<string | null | undefined>('shelf-host')
}

function mockByDomain(status: 200 | 404 | 500, onRequest?: (domain: string) => void) {
  server.use(http.get(`${BASE}/v1/shelves/by-domain/:domain`, ({ params }) => {
    onRequest?.(params.domain as string)
    if (status === 200) {
      return HttpResponse.json(PublicShelfToJSON({ id: 'shelf-1', title: 'Profile', description: '', icon: '', path: '', theme: {} } as never))
    }
    return new HttpResponse(null, { status })
  }))
}

beforeEach(() => {
  setPageLayoutMock.mockClear()
  clearNuxtState('shelf-host')
})

afterEach(() => {
  visit(originalUrl)
})

describe('shelf-host.global middleware', () => {
  describe('when the server handed over what it found', () => {
    it('keeps a shelf domain and does not ask again', async () => {
      const requested = vi.fn()
      mockByDomain(200, requested)
      shelfHostState().value = 'profile.example.com'

      await shelfHostMiddleware(route('/'), {} as never)

      expect(shelfHostState().value).toBe('profile.example.com')
      expect(requested).not.toHaveBeenCalled()
    })

    it('keeps null for the instance\'s own host and does not ask again', async () => {
      const requested = vi.fn()
      mockByDomain(200, requested)
      shelfHostState().value = null

      await shelfHostMiddleware(route('/'), {} as never)

      expect(shelfHostState().value).toBeNull()
      expect(requested).not.toHaveBeenCalled()
    })
  })

  describe('layout', () => {
    it('switches "/" on a shelf host to the layout a shelf page uses', async () => {
      shelfHostState().value = 'profile.example.com'

      await shelfHostMiddleware(route('/'), {} as never)

      expect(setPageLayoutMock).toHaveBeenCalledWith('links')
    })

    it('leaves every other page on a shelf host alone', async () => {
      shelfHostState().value = 'profile.example.com'

      await shelfHostMiddleware(route('/app'), {} as never)
      await shelfHostMiddleware(route('/some-path'), {} as never)

      expect(setPageLayoutMock).not.toHaveBeenCalled()
    })

    it('leaves "/" on the instance\'s own host alone', async () => {
      shelfHostState().value = null

      await shelfHostMiddleware(route('/'), {} as never)

      expect(setPageLayoutMock).not.toHaveBeenCalled()
    })
  })

  describe('when the page did not come from the server', () => {
    it('does not ask the API about a host that cannot be a domain', async () => {
      const requested = vi.fn()
      mockByDomain(200, requested)

      await shelfHostMiddleware(route('/'), {} as never)

      // The test window is on localhost.
      expect(shelfHostState().value).toBeNull()
      expect(requested).not.toHaveBeenCalled()
    })

    it('finds out from the API whether the host is a shelf domain', async () => {
      const requested = vi.fn()
      mockByDomain(200, requested)
      visit('https://Profile.Example.com/')

      await shelfHostMiddleware(route('/'), {} as never)

      expect(requested).toHaveBeenCalledWith('profile.example.com')
      expect(shelfHostState().value).toBe('profile.example.com')
      expect(setPageLayoutMock).toHaveBeenCalledWith('links')
    })

    it('includes a port in the domain it asks about', async () => {
      const requested = vi.fn()
      mockByDomain(200, requested)
      visit('https://profile.example.com:9443/')

      await shelfHostMiddleware(route('/'), {} as never)

      expect(requested).toHaveBeenCalledWith('profile.example.com:9443')
      expect(shelfHostState().value).toBe('profile.example.com:9443')
    })

    it('treats a host no shelf is served on as the instance\'s own', async () => {
      mockByDomain(404)
      visit('https://linkshelf.example.com/')

      await shelfHostMiddleware(route('/'), {} as never)

      expect(shelfHostState().value).toBeNull()
      expect(setPageLayoutMock).not.toHaveBeenCalled()
    })

    it('treats an API that fails as the instance\'s own host, so the site keeps working', async () => {
      mockByDomain(500)
      visit('https://linkshelf.example.com/')

      await shelfHostMiddleware(route('/'), {} as never)

      expect(shelfHostState().value).toBeNull()
    })
  })
})
