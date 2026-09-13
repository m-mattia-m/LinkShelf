import { renderSuspended } from '@nuxt/test-utils/runtime'
import { screen, waitFor } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import AboutPage from './about.vue'

describe('about page', () => {
  it('renders the markdown "about" content from the page settings', async () => {
    useState('settings').value = { about: '# Hello there\n\nSome about text.' }

    await renderSuspended(AboutPage)

    await waitFor(() => {
      expect(screen.getByText('Hello there')).toBeInTheDocument()
    })
    expect(screen.getByText('Some about text.')).toBeInTheDocument()
  })

  it('renders an empty article without error when there is no content yet', async () => {
    useState('settings').value = { about: '' }

    const { container } = await renderSuspended(AboutPage)

    expect(container.querySelector('article')).toBeInTheDocument()
  })
})
