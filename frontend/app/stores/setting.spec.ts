import { HttpResponse, http } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'
import { SettingBatchResponseBodyToJSON, SettingPageBodyToJSON } from '~~/api'
import { errorResponse } from '../../test/mocks/handlers'
import { server } from '../../test/mocks/server'
import { buildSettingPageBody } from '../../test/mocks/factories'
import { useSettingStore } from './setting'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  useSettingStore().$reset()
})

describe('useSettingStore', () => {
  describe('fetch', () => {
    it('loads the settings page for a language', async () => {
      const page = buildSettingPageBody({ about: 'About us' })
      server.use(http.get(`${BASE}/v1/settings`, ({ request }) => {
        expect(new URL(request.url).searchParams.get('language_code')).toBe('en')
        return HttpResponse.json(SettingPageBodyToJSON(page))
      }))
      const store = useSettingStore()

      await store.fetch('en')

      expect(store.page?.about).toBe('About us')
      expect(store.languageCode).toBe('en')
      expect(store.loaded).toBe(true)
    })
  })

  describe('updateMany', () => {
    it('sends nothing and returns no failures when no entry differs from the current page', async () => {
      const store = useSettingStore()
      store.page = buildSettingPageBody({ about: 'Existing', aboutShow: true })
      let calls = 0
      server.use(http.put(`${BASE}/v1/settings/batch`, () => {
        calls += 1
        return HttpResponse.json(SettingBatchResponseBodyToJSON({ settings: buildSettingPageBody(), failures: [] }))
      }))

      const result = await store.updateMany('en', [
        { key: 'about', value: 'Existing' },
        { key: 'about_show', value: 'true' }
      ])

      expect(result).toEqual([])
      expect(calls).toBe(0)
    })

    it('treats an unset page as all-blank defaults for comparison', async () => {
      const store = useSettingStore()
      let calls = 0
      server.use(http.put(`${BASE}/v1/settings/batch`, () => {
        calls += 1
        return HttpResponse.json(SettingBatchResponseBodyToJSON({ settings: buildSettingPageBody(), failures: [] }))
      }))

      const result = await store.updateMany('en', [
        { key: 'about', value: '' },
        { key: 'redirect_to_dashboard', value: 'false' }
      ])

      expect(result).toEqual([])
      expect(calls).toBe(0)
    })

    it('sends only the changed entries and applies the returned settings', async () => {
      const store = useSettingStore()
      store.page = buildSettingPageBody({ about: 'Old about', contact: 'Old contact' })
      const savedPage = buildSettingPageBody({ about: 'New about', contact: 'Old contact' })
      let receivedBody: unknown
      server.use(http.put(`${BASE}/v1/settings/batch`, async ({ request }) => {
        receivedBody = await request.json()
        return HttpResponse.json(SettingBatchResponseBodyToJSON({ settings: savedPage, failures: [] }))
      }))

      const result = await store.updateMany('de', [
        { key: 'about', value: 'New about' },
        { key: 'contact', value: 'Old contact' }
      ])

      expect(result).toEqual([])
      expect(receivedBody).toEqual({
        settings: [{ key: 'about', language_code: 'de', value: 'New about' }]
      })
      expect(store.page?.about).toBe('New about')
      expect(store.languageCode).toBe('de')
    })

    it('treats an unrecognized key as blank so any non-empty value counts as changed', async () => {
      const store = useSettingStore()
      let receivedBody: unknown
      server.use(http.put(`${BASE}/v1/settings/batch`, async ({ request }) => {
        receivedBody = await request.json()
        return HttpResponse.json(SettingBatchResponseBodyToJSON({ settings: buildSettingPageBody(), failures: [] }))
      }))

      await store.updateMany('en', [{ key: 'not_a_real_key', value: 'x' }])

      expect(receivedBody).toEqual({
        settings: [{ key: 'not_a_real_key', language_code: 'en', value: 'x' }]
      })
    })

    it('returns per-item failures reported by the server alongside a successful save', async () => {
      const store = useSettingStore()
      store.page = buildSettingPageBody({ about: 'Old' })
      server.use(http.put(`${BASE}/v1/settings/batch`, () => HttpResponse.json(SettingBatchResponseBodyToJSON({
        settings: buildSettingPageBody({ about: 'New' }),
        failures: [{ key: 'about', languageCode: 'en', reason: 'rejected by moderation' }]
      }))))

      const result = await store.updateMany('en', [{ key: 'about', value: 'New' }])

      expect(result).toEqual([{ key: 'about', languageCode: 'en', reason: 'rejected by moderation' }])
      expect(store.page?.about).toBe('New')
    })

    it('maps a request-level failure onto every changed entry without mutating stored state', async () => {
      const store = useSettingStore()
      store.page = buildSettingPageBody({ about: 'Old', contact: 'Old contact' })
      server.use(http.put(`${BASE}/v1/settings/batch`, () => errorResponse(500, 'internal server error')))

      const result = await store.updateMany('en', [
        { key: 'about', value: 'New' },
        { key: 'contact', value: 'New contact' }
      ])

      expect(result).toEqual([
        { key: 'about', languageCode: 'en', reason: 'internal server error' },
        { key: 'contact', languageCode: 'en', reason: 'internal server error' }
      ])
      expect(store.page?.about).toBe('Old')
      expect(store.languageCode).toBe('en')
    })
  })
})
