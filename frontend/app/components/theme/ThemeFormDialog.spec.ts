import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { fireEvent, screen, waitFor, within } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import { ThemeGroupedResponseBodyToJSON, ThemeToJSON } from '~~/api'
import { errorResponse } from '../../../test/mocks/handlers'
import { server } from '../../../test/mocks/server'
import { buildTheme, buildThemeGrouped } from '../../../test/mocks/factories'
import ThemeFormDialog from './ThemeFormDialog.vue'

const BASE = 'http://localhost:8085'

describe('ThemeFormDialog', () => {
  it('creates a theme from the entered fields and emits saved', async () => {
    let postedBody: unknown
    server.use(
      http.get(`${BASE}/v1/themes`, () => HttpResponse.json(ThemeGroupedResponseBodyToJSON(buildThemeGrouped()))),
      http.post(`${BASE}/v1/themes`, async ({ request }) => {
        postedBody = await request.json()
        return HttpResponse.json(ThemeToJSON(buildTheme({ id: 'theme-new', name: 'New Theme', config: '--shelf-bg: #000;' })))
      })
    )

    const { emitted, baseElement } = await renderSuspended(ThemeFormDialog)

    await fireEvent.click(screen.getByRole('button', { name: 'New theme' }))
    const dialog = within(baseElement as HTMLElement)
    await fireEvent.update(dialog.getByLabelText('Name'), 'New Theme')
    await fireEvent.update(dialog.getByLabelText('Config'), '--shelf-bg: #000;')
    await fireEvent.click(dialog.getByRole('button', { name: 'Save' }))

    await waitFor(() => {
      expect(emitted().saved).toBeTruthy()
    })
    expect(postedBody).toEqual({ name: 'New Theme', config: '--shelf-bg: #000;' })
    expect((emitted().saved![0] as unknown[])[0]).toMatchObject({ id: 'theme-new', name: 'New Theme' })
  })

  it('does not submit and stays open when required fields are empty', async () => {
    let called = false
    server.use(http.post(`${BASE}/v1/themes`, () => {
      called = true
      return HttpResponse.json(ThemeToJSON(buildTheme()))
    }))

    const { emitted, baseElement } = await renderSuspended(ThemeFormDialog)

    await fireEvent.click(screen.getByRole('button', { name: 'New theme' }))
    const dialog = within(baseElement as HTMLElement)
    await fireEvent.click(dialog.getByRole('button', { name: 'Save' }))

    expect(called).toBe(false)
    expect(emitted().saved).toBeFalsy()
    expect(dialog.getByLabelText('Name')).toBeInTheDocument()
  })

  it('shows the field error from the API and leaves the dialog open on failure', async () => {
    server.use(http.post(`${BASE}/v1/themes`, () => errorResponse(422, 'validation failed', [
      { location: 'body.name', message: 'already taken' }
    ])))

    const { emitted, baseElement } = await renderSuspended(ThemeFormDialog)

    await fireEvent.click(screen.getByRole('button', { name: 'New theme' }))
    const dialog = within(baseElement as HTMLElement)
    await fireEvent.update(dialog.getByLabelText('Name'), 'Duplicate')
    await fireEvent.update(dialog.getByLabelText('Config'), '--x: 1;')
    await fireEvent.click(dialog.getByRole('button', { name: 'Save' }))

    await waitFor(() => {
      expect(dialog.getByText('already taken')).toBeInTheDocument()
    })
    expect(emitted().saved).toBeFalsy()
  })

  it('edits an existing theme, pre-filling the form and sending the update', async () => {
    const theme = buildTheme({ id: 'theme-1', name: 'Old Name', config: '--x: 1;' })
    let putBody: unknown
    server.use(
      http.get(`${BASE}/v1/themes`, () => HttpResponse.json(ThemeGroupedResponseBodyToJSON(buildThemeGrouped()))),
      http.put(`${BASE}/v1/themes/theme-1`, async ({ request }) => {
        putBody = await request.json()
        return HttpResponse.json(ThemeToJSON({ ...theme, name: 'Updated Name' }))
      })
    )

    const { emitted, baseElement } = await renderSuspended(ThemeFormDialog, {
      props: { mode: 'edit', theme, open: true }
    })

    const dialog = within(baseElement as HTMLElement)
    expect(dialog.getByLabelText('Name')).toHaveValue('Old Name')

    await fireEvent.update(dialog.getByLabelText('Name'), 'Updated Name')
    await fireEvent.click(dialog.getByRole('button', { name: 'Save' }))

    await waitFor(() => {
      expect(emitted().saved).toBeTruthy()
    })
    expect(putBody).toMatchObject({ name: 'Updated Name' })
  })
})
