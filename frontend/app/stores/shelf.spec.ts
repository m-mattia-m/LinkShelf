import { HttpResponse, http } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'
import { ShelfToJSON } from '~~/api'
import { server } from '../../test/mocks/server'
import { buildShelf } from '../../test/mocks/factories'
import { useShelfStore } from './shelf'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  useShelfStore().$reset()
})

describe('useShelfStore', () => {
  describe('fetch', () => {
    it('loads the list of shelves', async () => {
      const shelves = [buildShelf(), buildShelf()]
      server.use(http.get(`${BASE}/v1/shelves`, () => HttpResponse.json(shelves.map(ShelfToJSON))))
      const store = useShelfStore()

      await store.fetch()

      expect(store.shelves).toHaveLength(2)
      expect(store.loaded).toBe(true)
    })
  })

  describe('create', () => {
    it('creates a shelf and refetches the list', async () => {
      const store = useShelfStore()
      const created = buildShelf({ id: 'shelf-new', path: 'shelf-new' })
      server.use(
        http.post(`${BASE}/v1/shelves`, () => HttpResponse.json(ShelfToJSON(created))),
        http.get(`${BASE}/v1/shelves`, () => HttpResponse.json([ShelfToJSON(created)]))
      )

      const result = await store.create({ path: 'shelf-new', title: 'New Shelf', description: '', domain: '', icon: 'i-lucide-book' })

      expect(result.id).toBe('shelf-new')
      expect(store.shelves).toHaveLength(1)
    })
  })

  describe('update', () => {
    it('updates a shelf and refetches the list', async () => {
      const store = useShelfStore()
      const updated = buildShelf({ id: 'shelf-1', description: 'Updated' })
      server.use(
        http.put(`${BASE}/v1/shelves/:shelfId`, ({ params }) => {
          expect(params.shelfId).toBe('shelf-1')
          return HttpResponse.json(ShelfToJSON(updated))
        }),
        http.get(`${BASE}/v1/shelves`, () => HttpResponse.json([ShelfToJSON(updated)]))
      )

      const result = await store.update('shelf-1', { path: 'shelf-1', title: 'Shelf', description: 'Updated', domain: '', icon: 'i-lucide-book' })

      expect(result.description).toBe('Updated')
      expect(store.shelves).toHaveLength(1)
    })
  })

  describe('remove', () => {
    it('deletes a shelf and refetches the list', async () => {
      const store = useShelfStore()
      let deleteCalled = false
      server.use(
        http.delete(`${BASE}/v1/shelves/:shelfId`, ({ params }) => {
          deleteCalled = true
          expect(params.shelfId).toBe('shelf-1')
          return new HttpResponse(null, { status: 204 })
        }),
        http.get(`${BASE}/v1/shelves`, () => HttpResponse.json([]))
      )

      await store.remove('shelf-1')

      expect(deleteCalled).toBe(true)
      expect(store.loaded).toBe(true)
    })
  })

  describe('getById', () => {
    it('fetches a single shelf without touching store state', async () => {
      server.use(http.get(`${BASE}/v1/shelves/:shelfId`, ({ params }) => HttpResponse.json(ShelfToJSON(buildShelf({ id: params.shelfId as string })))))
      const store = useShelfStore()

      const result = await store.getById('shelf-1')

      expect(result.id).toBe('shelf-1')
      expect(store.shelves).toEqual([])
      expect(store.loaded).toBe(false)
    })
  })
})
