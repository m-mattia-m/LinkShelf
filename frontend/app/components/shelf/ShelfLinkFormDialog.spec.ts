import { renderSuspended } from '@nuxt/test-utils/runtime'
import { fireEvent, waitFor, within } from '@testing-library/vue'
import { HttpResponse, http } from 'msw'
import { describe, expect, it } from 'vitest'
import { LinkToJSON } from '~~/api'
import type { Link } from '~~/api'
import { server } from '../../../test/mocks/server'
import { buildLink } from '../../../test/mocks/factories'
import ShelfLinkFormDialog from './ShelfLinkFormDialog.vue'

const BASE = 'http://localhost:8085'

describe('ShelfLinkFormDialog', () => {
  it('initializes its fields from the "link" prop in edit mode', async () => {
    const link = buildLink({ id: 'link-1', title: 'My Link', link: 'https://example.com/page', icon: 'i-lucide-star', color: '#588157' })

    const { baseElement } = await renderSuspended(ShelfLinkFormDialog, {
      props: { mode: 'edit', sectionId: 'section-1', link, open: true }
    })

    expect(within(baseElement as HTMLElement).getByLabelText('Title')).toHaveValue('My Link')
    expect(within(baseElement as HTMLElement).getByLabelText('URL')).toHaveValue('https://example.com/page')
  })

  it('starts with empty fields in create mode', async () => {
    const { baseElement } = await renderSuspended(ShelfLinkFormDialog, {
      props: { mode: 'create', sectionId: 'section-1', open: true }
    })

    expect(within(baseElement as HTMLElement).getByLabelText('Title')).toHaveValue('')
    expect(within(baseElement as HTMLElement).getByLabelText('URL')).toHaveValue('')
    expect(within(baseElement as HTMLElement).getByRole('button', { name: 'Set a color' })).toBeInTheDocument()
  })

  it('creates a link and emits "saved"', async () => {
    let receivedBody: unknown
    server.use(http.post(`${BASE}/v1/links`, async ({ request }) => {
      receivedBody = await request.json()
      return HttpResponse.json(LinkToJSON(buildLink({ id: 'link-new', title: 'New Link', link: 'https://example.com', sectionId: 'section-1' })))
    }))

    const { emitted, baseElement } = await renderSuspended(ShelfLinkFormDialog, {
      props: { mode: 'create', sectionId: 'section-1', open: true }
    })

    await fireEvent.update(within(baseElement as HTMLElement).getByLabelText('Title'), 'New Link')
    await fireEvent.update(within(baseElement as HTMLElement).getByLabelText('URL'), 'https://example.com')
    await fireEvent.click(within(baseElement as HTMLElement).getByRole('button', { name: 'Submit' }))

    await waitFor(() => {
      expect(emitted().saved).toBeTruthy()
    })
    expect(receivedBody).toMatchObject({ title: 'New Link', link: 'https://example.com', sectionId: 'section-1' })
    const openEvents = emitted()['update:open'] as unknown[][]
    expect(openEvents[openEvents.length - 1]).toEqual([false])
  })

  it('updates an existing link and emits "saved"', async () => {
    const link: Link = buildLink({ id: 'link-1', title: 'Old', link: 'https://old.example.com', sectionId: 'section-1' })
    let receivedLinkId: string | undefined
    server.use(http.put(`${BASE}/v1/links/:linkId`, ({ params }) => {
      receivedLinkId = params.linkId as string
      return HttpResponse.json(LinkToJSON(buildLink({ ...link, title: 'Updated' })))
    }))

    const { emitted, baseElement } = await renderSuspended(ShelfLinkFormDialog, {
      props: { mode: 'edit', sectionId: 'section-1', link, open: true }
    })

    await fireEvent.update(within(baseElement as HTMLElement).getByLabelText('Title'), 'Updated')
    await fireEvent.click(within(baseElement as HTMLElement).getByRole('button', { name: 'Submit' }))

    await waitFor(() => {
      expect(emitted().saved).toBeTruthy()
    })
    expect(receivedLinkId).toBe('link-1')
  })

  it('does not submit when title or URL is missing', async () => {
    let postCalled = false
    server.use(http.post(`${BASE}/v1/links`, () => {
      postCalled = true
      return HttpResponse.json(LinkToJSON(buildLink()))
    }))

    const { emitted, baseElement } = await renderSuspended(ShelfLinkFormDialog, {
      props: { mode: 'create', sectionId: 'section-1', open: true }
    })

    await fireEvent.click(within(baseElement as HTMLElement).getByRole('button', { name: 'Submit' }))

    await new Promise(resolve => setTimeout(resolve, 20))
    expect(postCalled).toBe(false)
    expect(emitted().saved).toBeFalsy()
  })

  // Not covered: typing an invalid hex string into the color field to
  // exercise the color schema's regex-rejection branch. UColorPicker's own
  // v-model wiring (its writable `pickedColor` computed, synced on mount via
  // vueuse's watchPausable) normalizes any unparsable color string back to a
  // valid hex (e.g. "#FFFFFF") before the form's own validation ever sees
  // the bad value, so this branch can't be reached by simulating typing.

  it('clears a set color back to the "Set a color" button', async () => {
    const { baseElement } = await renderSuspended(ShelfLinkFormDialog, {
      props: { mode: 'create', sectionId: 'section-1', open: true }
    })
    const scope = within(baseElement as HTMLElement)

    await fireEvent.click(scope.getByRole('button', { name: 'Set a color' }))
    expect(scope.getByPlaceholderText('#000000')).toBeInTheDocument()

    await fireEvent.click(scope.getByRole('button', { name: 'Clear color' }))

    expect(scope.getByRole('button', { name: 'Set a color' })).toBeInTheDocument()
  })
})
