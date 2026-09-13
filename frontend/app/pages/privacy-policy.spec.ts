import { renderSuspended } from '@nuxt/test-utils/runtime'
import { screen, waitFor } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import PrivacyPolicyPage from './privacy-policy.vue'

describe('privacy policy page', () => {
  it('renders the markdown "privacyPolicy" content from the page settings', async () => {
    useState('settings').value = { privacyPolicy: '# Privacy\n\nSome privacy text.' }

    await renderSuspended(PrivacyPolicyPage)

    await waitFor(() => {
      expect(screen.getByText('Privacy')).toBeInTheDocument()
    })
    expect(screen.getByText('Some privacy text.')).toBeInTheDocument()
  })

  it('renders an empty article without error when there is no content yet', async () => {
    useState('settings').value = { privacyPolicy: '' }

    const { container } = await renderSuspended(PrivacyPolicyPage)

    expect(container.querySelector('article')).toBeInTheDocument()
  })
})
