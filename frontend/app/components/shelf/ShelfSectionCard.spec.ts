import { mountSuspended, renderSuspended } from '@nuxt/test-utils/runtime'
import { fireEvent, screen, waitFor, within } from '@testing-library/vue'
import { HttpResponse, http } from 'msw'
import { describe, expect, it } from 'vitest'
import draggable from 'vuedraggable'
import type { Link, Section } from '~~/api'
import { server } from '../../../test/mocks/server'
import { buildLink, buildSection } from '../../../test/mocks/factories'
import ShelfSectionCard from './ShelfSectionCard.vue'

const BASE = 'http://localhost:8085'

function buildSectionProp(overrides: Partial<Section> = {}): Section {
  return buildSection({ id: 'section-1', shelfId: 'shelf-1', title: 'My Section', order: 0, ...overrides })
}

describe('ShelfSectionCard', () => {
  it('shows an empty state and a shortcut to add the first link when there are no links', async () => {
    await renderSuspended(ShelfSectionCard, {
      props: { section: buildSectionProp(), links: [] }
    })

    expect(screen.getByText('No links yet.')).toBeInTheDocument()

    await fireEvent.click(screen.getByRole('button', { name: 'Add the first link' }))

    expect(screen.getByLabelText('Title')).toHaveValue('')
  })

  it('renders each link with its title and href', async () => {
    const links: Link[] = [
      buildLink({ id: 'link-1', title: 'Example', link: 'https://example.com', order: 0 }),
      buildLink({ id: 'link-2', title: 'Other', link: 'https://other.com', order: 1 })
    ]

    await renderSuspended(ShelfSectionCard, {
      props: { section: buildSectionProp(), links }
    })

    expect(screen.getByText('Example')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'https://example.com' })).toHaveAttribute('href', 'https://example.com')
    expect(screen.getByText('Other')).toBeInTheDocument()
  })

  it('opens the edit dialog pre-filled when editing a link', async () => {
    const link = buildLink({ id: 'link-1', title: 'Example', link: 'https://example.com', order: 0 })

    await renderSuspended(ShelfSectionCard, {
      props: { section: buildSectionProp(), links: [link] }
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Edit link' }))

    expect(screen.getByLabelText('Title')).toHaveValue('Example')
    expect(screen.getByLabelText('URL')).toHaveValue('https://example.com')
  })

  it('deletes a link after confirming', async () => {
    const link = buildLink({ id: 'link-1', title: 'Deletable', link: 'https://example.com', order: 0 })
    let deleteCalled = false
    server.use(http.delete(`${BASE}/v1/links/link-1`, () => {
      deleteCalled = true
      return new HttpResponse(null, { status: 204 })
    }))

    const { baseElement } = await renderSuspended(ShelfSectionCard, {
      props: { section: buildSectionProp(), links: [link] }
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Delete link' }))

    const scope = within(baseElement as HTMLElement)
    expect(scope.getByText('Delete “Deletable”? This cannot be undone.')).toBeInTheDocument()
    await fireEvent.click(scope.getByRole('button', { name: 'Delete', hidden: true }))

    await waitFor(() => {
      expect(deleteCalled).toBe(true)
    })
  })

  describe('renaming a section', () => {
    it('renames the section when Enter is pressed', async () => {
      let receivedBody: unknown
      server.use(http.put(`${BASE}/v1/sections/section-1`, async ({ request }) => {
        receivedBody = await request.json()
        return HttpResponse.json({ id: 'section-1', shelfId: 'shelf-1', title: 'Renamed', order: 0 })
      }))

      await renderSuspended(ShelfSectionCard, {
        props: { section: buildSectionProp({ title: 'My Section' }), links: [] }
      })

      await fireEvent.click(screen.getByRole('button', { name: 'Rename section' }))
      const input = screen.getByDisplayValue('My Section')
      await fireEvent.update(input, 'Renamed')
      await fireEvent.keyUp(input, { key: 'Enter' })

      await waitFor(() => {
        expect(receivedBody).toMatchObject({ title: 'Renamed', shelfId: 'shelf-1' })
      })
      await waitFor(() => {
        expect(screen.queryByDisplayValue('Renamed')).not.toBeInTheDocument()
      })
    })

    it('cancels renaming on Escape without calling the API', async () => {
      let putCalled = false
      server.use(http.put(`${BASE}/v1/sections/section-1`, () => {
        putCalled = true
        return HttpResponse.json({ id: 'section-1', shelfId: 'shelf-1', title: 'My Section', order: 0 })
      }))

      await renderSuspended(ShelfSectionCard, {
        props: { section: buildSectionProp({ title: 'My Section' }), links: [] }
      })

      await fireEvent.click(screen.getByRole('button', { name: 'Rename section' }))
      const input = screen.getByDisplayValue('My Section')
      await fireEvent.update(input, 'Ignored')
      await fireEvent.keyUp(input, { key: 'Escape' })

      expect(screen.queryByDisplayValue('Ignored')).not.toBeInTheDocument()
      expect(screen.getByText('My Section')).toBeInTheDocument()
      expect(putCalled).toBe(false)
    })
  })

  it('deletes the section after confirming, mentioning the link count', async () => {
    let deleteCalled = false
    server.use(http.delete(`${BASE}/v1/sections/section-1`, () => {
      deleteCalled = true
      return new HttpResponse(null, { status: 204 })
    }))
    const links = [buildLink({ id: 'link-1', order: 0 }), buildLink({ id: 'link-2', order: 1 })]

    const { baseElement } = await renderSuspended(ShelfSectionCard, {
      props: { section: buildSectionProp({ title: 'Doomed Section' }), links }
    })

    await fireEvent.click(screen.getByRole('button', { name: 'Delete section' }))

    const scope = within(baseElement as HTMLElement)
    expect(scope.getByText('This also deletes 2 link(s) in “Doomed Section”. This cannot be undone.')).toBeInTheDocument()
    await fireEvent.click(scope.getByRole('button', { name: 'Delete', hidden: true }))

    await waitFor(() => {
      expect(deleteCalled).toBe(true)
    })
  })

  // Simulating a real HTML5 drag-and-drop gesture is unreliable in
  // jsdom/happy-dom, so these exercise vuedraggable's actual wiring by
  // triggering its emitted events directly instead (mountSuspended, rather
  // than renderSuspended, so wrapper.findComponent() is available).
  describe('drag-and-drop reordering', () => {
    it('emits "reordered" when the draggable list reports a drag end', async () => {
      const links = [buildLink({ id: 'link-1', order: 0 }), buildLink({ id: 'link-2', order: 1 })]
      const wrapper = await mountSuspended(ShelfSectionCard, {
        props: { section: buildSectionProp(), links }
      })

      await wrapper.findComponent(draggable).vm.$emit('end')

      expect(wrapper.emitted('reordered')).toBeTruthy()
    })

    it('updates the "links" v-model when the draggable list reports a new order', async () => {
      const link1 = buildLink({ id: 'link-1', order: 0 })
      const link2 = buildLink({ id: 'link-2', order: 1 })
      const wrapper = await mountSuspended(ShelfSectionCard, {
        props: { section: buildSectionProp(), links: [link1, link2] }
      })

      const reordered = [link2, link1]
      await wrapper.findComponent(draggable).vm.$emit('update:modelValue', reordered)

      const events = wrapper.emitted('update:links')
      expect(events).toBeTruthy()
      expect(events![events!.length - 1]).toEqual([reordered])
    })
  })
})
