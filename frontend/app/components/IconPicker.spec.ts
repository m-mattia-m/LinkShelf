import { registerEndpoint, renderSuspended } from '@nuxt/test-utils/runtime'
import { fireEvent, screen, waitFor, within } from '@testing-library/vue'
import { beforeEach, describe, expect, it } from 'vitest'
import IconPicker from './IconPicker.vue'

beforeEach(() => {
  useState<string[]>('icon-picker-names').value = []
})

async function openPicker(baseElement: HTMLElement) {
  const trigger = screen.getByRole('button')
  await fireEvent.click(trigger)
  return within(baseElement)
}

describe('IconPicker', () => {
  it('shows the placeholder text and no clear button when no icon is selected', async () => {
    await renderSuspended(IconPicker, { props: { modelValue: '' } })

    expect(screen.getByText('No icon')).toBeInTheDocument()
    expect(screen.queryByLabelText('Clear icon')).not.toBeInTheDocument()
  })

  it('shows the selected icon name and a clear button when an icon is selected', async () => {
    await renderSuspended(IconPicker, { props: { modelValue: 'i-lucide-anchor' } })

    expect(screen.getByText('i-lucide-anchor')).toBeInTheDocument()
    expect(screen.getByLabelText('Clear icon')).toBeInTheDocument()
  })

  it('emits an empty string when the clear button is clicked', async () => {
    const { emitted } = await renderSuspended(IconPicker, { props: { modelValue: 'i-lucide-anchor' } })

    await fireEvent.click(screen.getByLabelText('Clear icon'))

    expect(emitted()['update:modelValue']?.at(-1)).toEqual([''])
  })

  it('fetches and displays the icon list when opened', async () => {
    registerEndpoint('/icon-names.json', () => ({
      'lucide': ['i-lucide-anchor', 'i-lucide-apple'],
      'simple-icons': ['i-simple-icons-github']
    }))

    const { baseElement } = await renderSuspended(IconPicker, { props: { modelValue: '' } })
    const scope = await openPicker(baseElement as HTMLElement)

    expect(await scope.findByTitle('i-lucide-anchor')).toBeInTheDocument()
    expect(scope.getByTitle('i-lucide-apple')).toBeInTheDocument()
    expect(scope.getByTitle('i-simple-icons-github')).toBeInTheDocument()
  })

  it('shows an error message when the icon list fails to load', async () => {
    registerEndpoint('/icon-names.json', {
      method: 'GET',
      handler: () => {
        throw createError({ statusCode: 500, statusMessage: 'Server error' })
      }
    })

    const { baseElement } = await renderSuspended(IconPicker, { props: { modelValue: '' } })
    const scope = await openPicker(baseElement as HTMLElement)

    expect(await scope.findByText('Could not load the icon list. You can still type an icon name directly.')).toBeInTheDocument()
  })

  it('filters the icon list by the search input', async () => {
    useState<string[]>('icon-picker-names').value = ['i-lucide-anchor', 'i-lucide-apple', 'i-lucide-banana']

    const { baseElement } = await renderSuspended(IconPicker, { props: { modelValue: '' } })
    const scope = await openPicker(baseElement as HTMLElement)
    const search = await scope.findByPlaceholderText('Search icons...')

    await fireEvent.update(search, 'ban')

    expect(scope.getByTitle('i-lucide-banana')).toBeInTheDocument()
    expect(scope.queryByTitle('i-lucide-anchor')).not.toBeInTheDocument()
    expect(scope.queryByTitle('i-lucide-apple')).not.toBeInTheDocument()
  })

  it('shows a "no icons found" message when the search matches nothing', async () => {
    useState<string[]>('icon-picker-names').value = ['i-lucide-anchor']

    const { baseElement } = await renderSuspended(IconPicker, { props: { modelValue: '' } })
    const scope = await openPicker(baseElement as HTMLElement)
    const search = await scope.findByPlaceholderText('Search icons...')

    await fireEvent.update(search, 'zzz-no-match')

    expect(scope.getByText('No icons found for "zzz-no-match".')).toBeInTheDocument()
  })

  it('caps the visible results at 120 and mentions the truncation', async () => {
    useState<string[]>('icon-picker-names').value = Array.from({ length: 150 }, (_, i) => `i-lucide-icon-${i}`)

    const { baseElement } = await renderSuspended(IconPicker, { props: { modelValue: '' } })
    const scope = await openPicker(baseElement as HTMLElement)

    await scope.findByTitle('i-lucide-icon-0')
    expect(scope.getByText('Showing 120 of 150 - keep typing to narrow it down.')).toBeInTheDocument()
  })

  it('selects an icon, emits it and closes the popover', async () => {
    useState<string[]>('icon-picker-names').value = ['i-lucide-anchor']

    const { baseElement, emitted } = await renderSuspended(IconPicker, { props: { modelValue: '' } })
    const scope = await openPicker(baseElement as HTMLElement)
    const iconButton = await scope.findByTitle('i-lucide-anchor')

    await fireEvent.click(iconButton)

    expect(emitted()['update:modelValue']?.at(-1)).toEqual(['i-lucide-anchor'])
    await waitFor(() => {
      expect(scope.queryByPlaceholderText('Search icons...')).not.toBeInTheDocument()
    })
  })
})
