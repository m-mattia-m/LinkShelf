import { renderSuspended } from '@nuxt/test-utils/runtime'
import { fireEvent, waitFor, within } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import ConfirmDialog from './ConfirmDialog.vue'

describe('ConfirmDialog', () => {
  it('does not render its content when closed', async () => {
    const { baseElement } = await renderSuspended(ConfirmDialog, {
      props: { title: 'Delete item?', open: false }
    })

    expect(within(baseElement as HTMLElement).queryByText('Delete item?')).not.toBeInTheDocument()
  })

  it('renders the title, description and default button labels when open', async () => {
    const { baseElement } = await renderSuspended(ConfirmDialog, {
      props: { title: 'Delete item?', description: 'This cannot be undone.', open: true }
    })
    const scope = within(baseElement as HTMLElement)

    expect(await scope.findByText('Delete item?')).toBeInTheDocument()
    expect(scope.getByText('This cannot be undone.')).toBeInTheDocument()
    expect(scope.getByRole('button', { name: 'Cancel' })).toBeInTheDocument()
    expect(scope.getByRole('button', { name: 'Delete' })).toBeInTheDocument()
  })

  it('omits the description paragraph when none is given', async () => {
    const { baseElement } = await renderSuspended(ConfirmDialog, {
      props: { title: 'Delete item?', open: true }
    })
    const scope = within(baseElement as HTMLElement)
    await scope.findByText('Delete item?')

    expect(scope.queryByText('This cannot be undone.')).not.toBeInTheDocument()
  })

  it('renders custom confirm/cancel labels', async () => {
    const { baseElement } = await renderSuspended(ConfirmDialog, {
      props: { title: 'Archive item?', confirmLabel: 'Archive', cancelLabel: 'Keep it', open: true }
    })
    const scope = within(baseElement as HTMLElement)
    await scope.findByText('Archive item?')

    expect(scope.getByRole('button', { name: 'Archive' })).toBeInTheDocument()
    expect(scope.getByRole('button', { name: 'Keep it' })).toBeInTheDocument()
  })

  it('emits "confirm" when the confirm button is clicked, without closing itself', async () => {
    const { baseElement, emitted } = await renderSuspended(ConfirmDialog, {
      props: { title: 'Delete item?', open: true }
    })
    const scope = within(baseElement as HTMLElement)
    const confirmButton = await scope.findByRole('button', { name: 'Delete' })

    await fireEvent.click(confirmButton)

    expect(emitted().confirm).toBeTruthy()
    expect(scope.getByText('Delete item?')).toBeInTheDocument()
  })

  it('closes when the cancel button is clicked', async () => {
    const { baseElement } = await renderSuspended(ConfirmDialog, {
      props: { title: 'Delete item?', open: true }
    })
    const scope = within(baseElement as HTMLElement)
    const cancelButton = await scope.findByRole('button', { name: 'Cancel' })

    await fireEvent.click(cancelButton)

    await waitFor(() => {
      expect(scope.queryByText('Delete item?')).not.toBeInTheDocument()
    })
  })

  it('disables the confirm button while loading', async () => {
    const { baseElement } = await renderSuspended(ConfirmDialog, {
      props: { title: 'Delete item?', open: true, loading: true }
    })
    const scope = within(baseElement as HTMLElement)
    const confirmButton = await scope.findByRole('button', { name: 'Delete' })

    expect(confirmButton).toBeDisabled()
  })

  it('does not disable the confirm button when not loading', async () => {
    const { baseElement } = await renderSuspended(ConfirmDialog, {
      props: { title: 'Delete item?', open: true, loading: false }
    })
    const scope = within(baseElement as HTMLElement)
    const confirmButton = await scope.findByRole('button', { name: 'Delete' })

    expect(confirmButton).not.toBeDisabled()
  })
})
