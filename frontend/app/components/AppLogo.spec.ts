import { renderSuspended } from '@nuxt/test-utils/runtime'
import { describe, expect, it } from 'vitest'
import AppLogo from './AppLogo.vue'

describe('AppLogo', () => {
  it('renders the logo as a replaceable svg asset', async () => {
    const { container } = await renderSuspended(AppLogo)

    const img = container.querySelector('img')
    expect(img).toBeInTheDocument()
    expect(img).toHaveAttribute('src', '/logo.svg')
    expect(img).toHaveAttribute('alt')
  })
})
