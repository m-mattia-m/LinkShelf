import { renderSuspended } from '@nuxt/test-utils/runtime'
import { screen, waitFor } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import ImprintPage from './imprint.vue'

describe('imprint page', () => {
  it('renders the markdown "imprint" content from the page settings', async () => {
    useState('settings').value = { imprint: '# Legal notice\n\nSome imprint text.' }

    await renderSuspended(ImprintPage)

    await waitFor(() => {
      expect(screen.getByText('Legal notice')).toBeInTheDocument()
    })
    expect(screen.getByText('Some imprint text.')).toBeInTheDocument()
  })

  it('renders an empty article without error when there is no content yet', async () => {
    useState('settings').value = { imprint: '' }

    const { container } = await renderSuspended(ImprintPage)

    expect(container.querySelector('article')).toBeInTheDocument()
  })
})
