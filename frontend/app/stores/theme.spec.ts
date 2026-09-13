import { HttpResponse, http } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'
import { ThemeGroupedResponseBodyToJSON, ThemeToJSON } from '~~/api'
import { server } from '../../test/mocks/server'
import { buildTheme, buildThemeGrouped } from '../../test/mocks/factories'
import { useThemeStore } from './theme'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  useThemeStore().$reset()
})

describe('useThemeStore', () => {
  describe('fetch', () => {
    it('splits themes into instance and mine groups', async () => {
      const grouped = buildThemeGrouped({
        instance: [buildTheme({ scope: 'instance' })],
        mine: [buildTheme({ scope: 'user' })]
      })
      server.use(http.get(`${BASE}/v1/themes`, () => HttpResponse.json(ThemeGroupedResponseBodyToJSON(grouped))))
      const store = useThemeStore()

      await store.fetch()

      expect(store.instance).toHaveLength(1)
      expect(store.mine).toHaveLength(1)
      expect(store.loaded).toBe(true)
    })

    it('defaults both groups to an empty array when missing from the response', async () => {
      server.use(http.get(`${BASE}/v1/themes`, () => HttpResponse.json({})))
      const store = useThemeStore()

      await store.fetch()

      expect(store.instance).toEqual([])
      expect(store.mine).toEqual([])
      expect(store.loaded).toBe(true)
    })
  })

  describe('create', () => {
    it('creates a theme and refetches the grouped list', async () => {
      const store = useThemeStore()
      const created = buildTheme({ id: 'theme-new', name: 'New Theme' })
      server.use(
        http.post(`${BASE}/v1/themes`, () => HttpResponse.json(ThemeToJSON(created))),
        http.get(`${BASE}/v1/themes`, () => HttpResponse.json(ThemeGroupedResponseBodyToJSON({ instance: [], mine: [created] })))
      )

      const result = await store.create({ name: 'New Theme', config: '{}' })

      expect(result.id).toBe('theme-new')
      expect(store.mine).toHaveLength(1)
    })
  })

  describe('update', () => {
    it('updates a theme and refetches the grouped list', async () => {
      const store = useThemeStore()
      const updated = buildTheme({ id: 'theme-1', name: 'Updated' })
      server.use(
        http.put(`${BASE}/v1/themes/:themeId`, ({ params }) => {
          expect(params.themeId).toBe('theme-1')
          return HttpResponse.json(ThemeToJSON(updated))
        }),
        http.get(`${BASE}/v1/themes`, () => HttpResponse.json(ThemeGroupedResponseBodyToJSON({ instance: [], mine: [updated] })))
      )

      const result = await store.update('theme-1', { name: 'Updated', config: '{}' })

      expect(result.name).toBe('Updated')
      expect(store.mine).toHaveLength(1)
    })
  })

  describe('remove', () => {
    it('deletes a theme and refetches the grouped list', async () => {
      const store = useThemeStore()
      let deleteCalled = false
      server.use(
        http.delete(`${BASE}/v1/themes/:themeId`, ({ params }) => {
          deleteCalled = true
          expect(params.themeId).toBe('theme-1')
          return new HttpResponse(null, { status: 204 })
        }),
        http.get(`${BASE}/v1/themes`, () => HttpResponse.json(ThemeGroupedResponseBodyToJSON({ instance: [], mine: [] })))
      )

      await store.remove('theme-1')

      expect(deleteCalled).toBe(true)
      expect(store.loaded).toBe(true)
    })
  })
})
