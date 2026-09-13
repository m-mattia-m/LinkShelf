import { HttpResponse, http } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'
import { SectionToJSON } from '~~/api'
import { server } from '../../test/mocks/server'
import { buildSection } from '../../test/mocks/factories'
import { useSectionStore } from './section'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  useSectionStore().$reset()
})

describe('useSectionStore', () => {
  describe('fetch', () => {
    it('loads sections for a shelf', async () => {
      const sections = [buildSection({ shelfId: 'shelf-1' }), buildSection({ shelfId: 'shelf-1' })]
      server.use(http.get(`${BASE}/v1/sections`, ({ request }) => {
        expect(new URL(request.url).searchParams.get('shelfId')).toBe('shelf-1')
        return HttpResponse.json(sections.map(SectionToJSON))
      }))
      const store = useSectionStore()

      await store.fetch('shelf-1')

      expect(store.sections).toHaveLength(2)
      expect(store.loaded).toBe(true)
    })
  })

  describe('create', () => {
    it('creates a section and re-fetches its shelf', async () => {
      const store = useSectionStore()
      const created = buildSection({ id: 'section-new', shelfId: 'shelf-1', title: 'New' })
      server.use(
        http.post(`${BASE}/v1/sections`, () => HttpResponse.json(SectionToJSON(created))),
        http.get(`${BASE}/v1/sections`, ({ request }) => {
          expect(new URL(request.url).searchParams.get('shelfId')).toBe('shelf-1')
          return HttpResponse.json([SectionToJSON(created)])
        })
      )

      const result = await store.create({ shelfId: 'shelf-1', title: 'New' })

      expect(result.id).toBe('section-new')
      expect(store.sections).toHaveLength(1)
    })
  })

  describe('update', () => {
    it('updates a section and re-fetches its shelf', async () => {
      const store = useSectionStore()
      const updated = buildSection({ id: 'section-1', shelfId: 'shelf-1', title: 'Updated' })
      server.use(
        http.put(`${BASE}/v1/sections/:sectionId`, ({ params }) => {
          expect(params.sectionId).toBe('section-1')
          return HttpResponse.json(SectionToJSON(updated))
        }),
        http.get(`${BASE}/v1/sections`, ({ request }) => {
          expect(new URL(request.url).searchParams.get('shelfId')).toBe('shelf-1')
          return HttpResponse.json([SectionToJSON(updated)])
        })
      )

      const result = await store.update('section-1', { shelfId: 'shelf-1', title: 'Updated' })

      expect(result.title).toBe('Updated')
      expect(store.sections).toHaveLength(1)
    })
  })

  describe('remove', () => {
    it('deletes a section and re-fetches the given shelf', async () => {
      const store = useSectionStore()
      let deleteCalled = false
      server.use(
        http.delete(`${BASE}/v1/sections/:sectionId`, ({ params }) => {
          deleteCalled = true
          expect(params.sectionId).toBe('section-1')
          return new HttpResponse(null, { status: 204 })
        }),
        http.get(`${BASE}/v1/sections`, ({ request }) => {
          expect(new URL(request.url).searchParams.get('shelfId')).toBe('shelf-1')
          return HttpResponse.json([])
        })
      )

      await store.remove('section-1', 'shelf-1')

      expect(deleteCalled).toBe(true)
      expect(store.loaded).toBe(true)
    })
  })
})
