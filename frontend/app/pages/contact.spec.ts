import { renderSuspended } from '@nuxt/test-utils/runtime'
import { screen, waitFor } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import ContactPage from './contact.vue'

describe('contact page', () => {
  it('renders the markdown "contact" content from the page settings', async () => {
    useState('settings').value = { contact: '# Get in touch\n\nSome contact text.' }

    await renderSuspended(ContactPage)

    await waitFor(() => {
      expect(screen.getByText('Get in touch')).toBeInTheDocument()
    })
    expect(screen.getByText('Some contact text.')).toBeInTheDocument()
  })

  it('renders an empty article without error when there is no content yet', async () => {
    useState('settings').value = { contact: '' }

    const { container } = await renderSuspended(ContactPage)

    expect(container.querySelector('article')).toBeInTheDocument()
  })
})
