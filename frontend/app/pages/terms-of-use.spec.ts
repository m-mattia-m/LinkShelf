import { renderSuspended } from '@nuxt/test-utils/runtime'
import { screen, waitFor } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import TermsOfUsePage from './terms-of-use.vue'

describe('terms of use page', () => {
  it('renders the markdown "termsOfUse" content from the page settings', async () => {
    useState('settings').value = { termsOfUse: '# Terms\n\nSome terms text.' }

    await renderSuspended(TermsOfUsePage)

    await waitFor(() => {
      expect(screen.getByText('Terms')).toBeInTheDocument()
    })
    expect(screen.getByText('Some terms text.')).toBeInTheDocument()
  })

  it('renders an empty article without error when there is no content yet', async () => {
    useState('settings').value = { termsOfUse: '' }

    const { container } = await renderSuspended(TermsOfUsePage)

    expect(container.querySelector('article')).toBeInTheDocument()
  })
})
