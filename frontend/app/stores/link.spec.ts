import { HttpResponse, http } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'
import { LinkToJSON } from '~~/api'
import { server } from '../../test/mocks/server'
import { buildLink } from '../../test/mocks/factories'
import { useLinkStore } from './link'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  useLinkStore().$reset()
})

describe('useLinkStore', () => {
  describe('fetch', () => {
    it('loads links for a shelf and records the current shelf', async () => {
      const links = [buildLink({ sectionId: 'section-1' }), buildLink({ sectionId: 'section-2' })]
      server.use(http.get(`${BASE}/v1/links`, ({ request }) => {
        expect(new URL(request.url).searchParams.get('shelfId')).toBe('shelf-1')
        return HttpResponse.json(links.map(LinkToJSON))
      }))
      const store = useLinkStore()

      await store.fetch('shelf-1')

      expect(store.links).toHaveLength(2)
      expect(store.loaded).toBe(true)
      expect(store.currentShelfId).toBe('shelf-1')
    })
  })

  describe('bySectionId', () => {
    it('groups links by their sectionId, preserving order within each group', () => {
      const store = useLinkStore()
      const a = buildLink({ sectionId: 'section-1', title: 'A' })
      const b = buildLink({ sectionId: 'section-2', title: 'B' })
      const c = buildLink({ sectionId: 'section-1', title: 'C' })
      store.links = [a, b, c]

      const grouped = store.bySectionId

      expect(grouped.get('section-1')).toEqual([a, c])
      expect(grouped.get('section-2')).toEqual([b])
    })

    it('returns an empty map when there are no links', () => {
      const store = useLinkStore()

      expect(store.bySectionId.size).toBe(0)
    })
  })

  describe('refetch', () => {
    it('re-fetches links for the current shelf', async () => {
      const store = useLinkStore()
      store.currentShelfId = 'shelf-9'
      let calls = 0
      server.use(http.get(`${BASE}/v1/links`, () => {
        calls += 1
        return HttpResponse.json([])
      }))

      await store.refetch()

      expect(calls).toBe(1)
    })

    it('is a no-op when there is no current shelf', async () => {
      const store = useLinkStore()
      let calls = 0
      server.use(http.get(`${BASE}/v1/links`, () => {
        calls += 1
        return HttpResponse.json([])
      }))

      await store.refetch()

      expect(calls).toBe(0)
      expect(store.loaded).toBe(false)
    })
  })

  describe('create', () => {
    it('creates a link and refetches the current shelf', async () => {
      const store = useLinkStore()
      store.currentShelfId = 'shelf-1'
      const created = buildLink({ id: 'link-new', title: 'New Link' })
      server.use(
        http.post(`${BASE}/v1/links`, () => HttpResponse.json(LinkToJSON(created))),
        http.get(`${BASE}/v1/links`, () => HttpResponse.json([LinkToJSON(created)]))
      )

      const result = await store.create({ sectionId: 'section-1', title: 'New Link', link: 'https://example.com' })

      expect(result.id).toBe('link-new')
      expect(store.links).toHaveLength(1)
    })
  })

  describe('update', () => {
    it('updates a link, scoping the request to the current shelf, and refetches', async () => {
      const store = useLinkStore()
      store.currentShelfId = 'shelf-1'
      const updated = buildLink({ id: 'link-1', title: 'Updated' })
      server.use(
        http.put(`${BASE}/v1/links/:linkId`, ({ params, request }) => {
          expect(params.linkId).toBe('link-1')
          expect(new URL(request.url).searchParams.get('shelfId')).toBe('shelf-1')
          return HttpResponse.json(LinkToJSON(updated))
        }),
        http.get(`${BASE}/v1/links`, () => HttpResponse.json([LinkToJSON(updated)]))
      )

      const result = await store.update('link-1', { sectionId: 'section-1', title: 'Updated', link: 'https://example.com' })

      expect(result.title).toBe('Updated')
      expect(store.links).toHaveLength(1)
    })

    it('omits the shelf scope when there is no current shelf', async () => {
      const store = useLinkStore()
      server.use(
        http.put(`${BASE}/v1/links/:linkId`, ({ request }) => {
          expect(new URL(request.url).searchParams.get('shelfId')).toBeNull()
          return HttpResponse.json(LinkToJSON(buildLink({ id: 'link-1' })))
        }),
        http.get(`${BASE}/v1/links`, () => HttpResponse.json([]))
      )

      await store.update('link-1', { sectionId: 'section-1', title: 'Updated', link: 'https://example.com' })
    })
  })

  describe('remove', () => {
    it('deletes a link and refetches the current shelf', async () => {
      const store = useLinkStore()
      store.currentShelfId = 'shelf-1'
      let deleteCalled = false
      server.use(
        http.delete(`${BASE}/v1/links/:linkId`, ({ params, request }) => {
          deleteCalled = true
          expect(params.linkId).toBe('link-1')
          expect(new URL(request.url).searchParams.get('shelfId')).toBe('shelf-1')
          return new HttpResponse(null, { status: 204 })
        }),
        http.get(`${BASE}/v1/links`, () => HttpResponse.json([]))
      )

      await store.remove('link-1')

      expect(deleteCalled).toBe(true)
    })
  })
})
