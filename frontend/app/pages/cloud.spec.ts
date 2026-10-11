import { renderSuspended } from '@nuxt/test-utils/runtime'
import { screen } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import CloudPage from './cloud.vue'

describe('cloud page', () => {
  it('renders the cloud heading and description', async () => {
    await renderSuspended(CloudPage)

    expect(screen.getByRole('heading', { name: 'LinkShelf Cloud' })).toBeInTheDocument()
    expect(screen.getByText(/managed cloud version/i)).toBeInTheDocument()
  })

  it('links to the cloud account page', async () => {
    await renderSuspended(CloudPage)

    expect(screen.getByRole('link', { name: /open linkshelf cloud/i })).toHaveAttribute('href', 'https://account.linkshelf.cloud')
  })

  it('links to the GitHub sponsors page', async () => {
    await renderSuspended(CloudPage)

    for (const link of screen.getAllByRole('link', { name: 'Sponsor on GitHub' })) {
      expect(link).toHaveAttribute('href', 'https://github.com/sponsors/m-mattia-m')
    }
  })
})
