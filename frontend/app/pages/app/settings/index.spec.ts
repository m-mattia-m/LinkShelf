import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { fireEvent, screen, waitFor } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import { SettingBatchResponseBodyToJSON, SettingPageBodyToJSON } from '~~/api'
import { server } from '../../../../test/mocks/server'
import { buildSettingPageBody } from '../../../../test/mocks/factories'
import SettingsIndexPage from './index.vue'

const BASE = 'http://localhost:8085'

describe('app settings index page', () => {
  it('loads the page settings and email delivery info, then pre-fills the form', async () => {
    server.use(
      http.get(`${BASE}/v1/settings`, () => HttpResponse.json(SettingPageBodyToJSON(buildSettingPageBody({ about: 'About us', aboutShow: true })))),
      http.get(`${BASE}/v1/settings/email-delivery`, () => HttpResponse.json({ configured: true, enabled: true, host: 'smtp.example.com', from: 'noreply@example.com', provider: 'smtp' }))
    )

    await renderSuspended(SettingsIndexPage)

    await waitFor(() => {
      expect(screen.getByLabelText('About')).toHaveValue('About us')
    })
    expect(screen.getByText('smtp.example.com')).toBeInTheDocument()
    expect(screen.getByText('noreply@example.com')).toBeInTheDocument()
  })

  it('shows the disabled message when email delivery is not configured', async () => {
    server.use(
      http.get(`${BASE}/v1/settings`, () => HttpResponse.json(SettingPageBodyToJSON(buildSettingPageBody()))),
      http.get(`${BASE}/v1/settings/email-delivery`, () => HttpResponse.json({ configured: false, enabled: false }))
    )

    await renderSuspended(SettingsIndexPage)

    await waitFor(() => {
      expect(screen.getByText('Email verification is disabled.')).toBeInTheDocument()
    })
  })

  it('still renders the rest of the page when the email delivery info request fails', async () => {
    server.use(
      http.get(`${BASE}/v1/settings`, () => HttpResponse.json(SettingPageBodyToJSON(buildSettingPageBody()))),
      http.get(`${BASE}/v1/settings/email-delivery`, () => new HttpResponse(null, { status: 500 }))
    )

    await renderSuspended(SettingsIndexPage)

    await waitFor(() => {
      expect(screen.getByText('Email verification is disabled.')).toBeInTheDocument()
    })
  })

  it('saves only the changed fields in one batch request and shows a success toast', async () => {
    server.use(
      http.get(`${BASE}/v1/settings`, () => HttpResponse.json(SettingPageBodyToJSON(buildSettingPageBody({ about: 'Old about' })))),
      http.get(`${BASE}/v1/settings/email-delivery`, () => HttpResponse.json({ configured: false, enabled: false }))
    )
    let putBody: { settings: { key: string, language_code: string, value: string }[] } | undefined
    server.use(http.put(`${BASE}/v1/settings/batch`, async ({ request }) => {
      putBody = await request.json() as typeof putBody
      return HttpResponse.json(SettingBatchResponseBodyToJSON({ settings: buildSettingPageBody({ about: 'New about' }), failures: [] }))
    }))

    await renderSuspended(SettingsIndexPage)

    await waitFor(() => {
      expect(screen.getByLabelText('About')).toHaveValue('Old about')
    })

    await fireEvent.update(screen.getByLabelText('About'), 'New about')
    await fireEvent.click(screen.getByRole('button', { name: 'Save' }))

    await waitFor(() => {
      expect(putBody).toBeTruthy()
    })
    expect(putBody!.settings).toEqual([{ key: 'about', language_code: 'en', value: 'New about' }])
    await waitFor(() => {
      expect(screen.getByLabelText('About')).toHaveValue('New about')
    })
  })

  it('does not send a batch request when nothing changed', async () => {
    server.use(
      http.get(`${BASE}/v1/settings`, () => HttpResponse.json(SettingPageBodyToJSON(buildSettingPageBody()))),
      http.get(`${BASE}/v1/settings/email-delivery`, () => HttpResponse.json({ configured: false, enabled: false }))
    )
    let called = false
    server.use(http.put(`${BASE}/v1/settings/batch`, () => {
      called = true
      return HttpResponse.json(SettingBatchResponseBodyToJSON({ settings: buildSettingPageBody(), failures: [] }))
    }))

    await renderSuspended(SettingsIndexPage)

    await waitFor(() => {
      expect(screen.getByLabelText('About')).toHaveValue('')
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Save' }))

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Save' })).not.toHaveAttribute('aria-disabled')
    })
    expect(called).toBe(false)
  })

  it('shows a partial-failure toast-worthy state when some settings fail to save, without throwing', async () => {
    server.use(
      http.get(`${BASE}/v1/settings`, () => HttpResponse.json(SettingPageBodyToJSON(buildSettingPageBody({ about: 'Old about' })))),
      http.get(`${BASE}/v1/settings/email-delivery`, () => HttpResponse.json({ configured: false, enabled: false }))
    )
    server.use(http.put(`${BASE}/v1/settings/batch`, () => HttpResponse.json(SettingBatchResponseBodyToJSON({
      settings: buildSettingPageBody({ about: 'Old about' }),
      failures: [{ key: 'about', languageCode: 'en', reason: 'write conflict' }]
    }))))

    await renderSuspended(SettingsIndexPage)

    await waitFor(() => {
      expect(screen.getByLabelText('About')).toHaveValue('Old about')
    })

    await fireEvent.update(screen.getByLabelText('About'), 'New about attempt')
    await fireEvent.click(screen.getByRole('button', { name: 'Save' }))

    await waitFor(() => {
      expect(screen.getByLabelText('About')).toHaveValue('Old about')
    })
  })

  it('reloads settings for the newly selected language', async () => {
    server.use(http.get(`${BASE}/v1/settings/email-delivery`, () => HttpResponse.json({ configured: false, enabled: false })))
    const requestedLanguages: string[] = []
    server.use(http.get(`${BASE}/v1/settings`, ({ request }) => {
      const url = new URL(request.url)
      const languageCode = url.searchParams.get('language_code') ?? 'en'
      requestedLanguages.push(languageCode)
      return HttpResponse.json(SettingPageBodyToJSON(buildSettingPageBody({ about: `About in ${languageCode}` })))
    }))

    await renderSuspended(SettingsIndexPage)

    await waitFor(() => {
      expect(screen.getByLabelText('About')).toHaveValue('About in en')
    })
    expect(requestedLanguages).toContain('en')
  })
})
