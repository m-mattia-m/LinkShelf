import { renderSuspended } from '@nuxt/test-utils/runtime'
import { screen } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import CloudPage from './cloud.vue'

describe('cloud page', () => {
  it('renders the "Cloud" heading and explanatory text', async () => {
    await renderSuspended(CloudPage)

    expect(screen.getByRole('heading', { name: 'Cloud' })).toBeInTheDocument()
    expect(screen.getByText(/this plan is not available yet/i)).toBeInTheDocument()
  })

  it('links to the GitHub sponsors page', async () => {
    await renderSuspended(CloudPage)

    expect(screen.getByRole('link', { name: 'Sponsoring' })).toHaveAttribute('href', 'https://github.com/sponsors/m-mattia-m')
  })
})
