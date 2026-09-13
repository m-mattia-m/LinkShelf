import { renderSuspended } from '@nuxt/test-utils/runtime'
import { fireEvent, screen, waitFor, within } from '@testing-library/vue'
import { HttpResponse, http } from 'msw'
import { beforeEach, describe, expect, it } from 'vitest'
import { ShelfToJSON } from '~~/api'
import type { Shelf } from '~~/api'
import { server } from '../../../test/mocks/server'
import { useThemeStore } from '~/stores/theme'
import ShelfFormDialog from './ShelfFormDialog.vue'

const BASE = 'http://localhost:8085'

function buildShelfProp(overrides: Partial<Shelf> = {}): Shelf {
  return {
    id: 'shelf-1',
    title: 'My Shelf',
    description: 'A shelf',
    domain: '',
    path: 'my-shelf',
    icon: 'i-lucide-book',
    theme: {},
    themeId: '',
    themeMissing: false,
    userId: 'user-1',
    ...overrides
  } as Shelf
}

beforeEach(() => {
  useThemeStore().loaded = true
})

describe('ShelfFormDialog', () => {
  it('shows a "New" trigger button in create mode', async () => {
    await renderSuspended(ShelfFormDialog, { props: { mode: 'create' } })

    expect(screen.getByRole('button', { name: 'New' })).toBeInTheDocument()
  })

  it('shows no trigger button in edit mode', async () => {
    await renderSuspended(ShelfFormDialog, { props: { mode: 'edit', shelf: buildShelfProp() } })

    expect(screen.queryByRole('button', { name: 'New' })).not.toBeInTheDocument()
  })

  it('creates a shelf and emits "saved" with the created shelf', async () => {
    const created = buildShelfProp({ id: 'shelf-new', title: 'Brand New', path: 'brand-new' })
    server.use(http.post(`${BASE}/v1/shelves`, () => HttpResponse.json(ShelfToJSON(created))))

    const { emitted, baseElement } = await renderSuspended(ShelfFormDialog, {
      props: { mode: 'create', open: true }
    })

    await fireEvent.update(within(baseElement as HTMLElement).getByLabelText('Title'), 'Brand New')
    await fireEvent.update(within(baseElement as HTMLElement).getByRole('textbox', { name: 'Path' }), 'brand-new')
    await fireEvent.click(within(baseElement as HTMLElement).getByRole('button', { name: 'Submit' }))

    await waitFor(() => {
      expect(emitted().saved).toBeTruthy()
    })
    expect(emitted().saved![0]).toEqual([created])
    expect(emitted()['update:open']).toBeTruthy()
    const openEvents = emitted()['update:open'] as unknown[][]
    expect(openEvents[openEvents.length - 1]).toEqual([false])
  })

  it('updates an existing shelf in edit mode', async () => {
    const shelf = buildShelfProp({ id: 'shelf-1', title: 'Old Title', path: 'old-path' })
    const updated = buildShelfProp({ id: 'shelf-1', title: 'New Title', path: 'old-path' })
    let receivedShelfId: string | undefined
    server.use(http.put(`${BASE}/v1/shelves/:shelfId`, ({ params }) => {
      receivedShelfId = params.shelfId as string
      return HttpResponse.json(ShelfToJSON(updated))
    }))

    const { emitted, baseElement } = await renderSuspended(ShelfFormDialog, {
      props: { mode: 'edit', shelf, open: true }
    })

    await fireEvent.update(within(baseElement as HTMLElement).getByLabelText('Title'), 'New Title')
    await fireEvent.click(within(baseElement as HTMLElement).getByRole('button', { name: 'Submit' }))

    await waitFor(() => {
      expect(emitted().saved).toBeTruthy()
    })
    expect(receivedShelfId).toBe('shelf-1')
    expect(emitted().saved![0]).toEqual([updated])
  })

  it('does not submit when the form fails validation', async () => {
    let postCalled = false
    server.use(http.post(`${BASE}/v1/shelves`, () => {
      postCalled = true
      return HttpResponse.json(ShelfToJSON(buildShelfProp()))
    }))

    const { emitted, baseElement } = await renderSuspended(ShelfFormDialog, {
      props: { mode: 'create', open: true }
    })

    // Leave Title, Domain and Path all empty - invalid on every count.
    await fireEvent.click(within(baseElement as HTMLElement).getByRole('button', { name: 'Submit' }))

    await new Promise(resolve => setTimeout(resolve, 20))
    expect(postCalled).toBe(false)
    expect(emitted().saved).toBeFalsy()
  })

  it('surfaces a field error from the API and keeps the dialog open', async () => {
    server.use(http.post(`${BASE}/v1/shelves`, () => HttpResponse.json({
      status: 422,
      detail: 'validation failed',
      errors: [{ location: 'body.title', message: 'already taken' }]
    }, { status: 422 })))

    const { emitted, baseElement } = await renderSuspended(ShelfFormDialog, {
      props: { mode: 'create', open: true }
    })

    await fireEvent.update(within(baseElement as HTMLElement).getByLabelText('Title'), 'Brand New')
    await fireEvent.update(within(baseElement as HTMLElement).getByRole('textbox', { name: 'Path' }), 'brand-new')
    await fireEvent.click(within(baseElement as HTMLElement).getByRole('button', { name: 'Submit' }))

    await waitFor(() => {
      expect(within(baseElement as HTMLElement).getByText('already taken')).toBeInTheDocument()
    })
    expect(emitted().saved).toBeFalsy()
  })
})
