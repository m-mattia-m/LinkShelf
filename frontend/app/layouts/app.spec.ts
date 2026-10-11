import { renderSuspended } from '@nuxt/test-utils/runtime'
import { screen } from '@testing-library/vue'
import { beforeEach, describe, expect, it } from 'vitest'
import AppLayout from './app.vue'

beforeEach(() => {
  useState('settings').value = null
  useAuthStore().user = { firstName: 'Ada', lastName: 'Lovelace', email: 'ada@example.com' } as never
})

describe('app layout', () => {
  it('shows configured sidebar links that open in a new tab', async () => {
    useState('settings').value = {
      sidebarLinks: [{ name: 'Status', url: 'https://status.example.com', icon: 'i-lucide-activity' }]
    }
    await renderSuspended(AppLayout)

    const link = screen.getAllByRole('link', { name: 'Status' })[0]
    expect(link).toHaveAttribute('href', 'https://status.example.com')
    expect(link).toHaveAttribute('target', '_blank')
  })

  it('skips links with an unsafe url', async () => {
    useState('settings').value = {
      sidebarLinks: [{ name: 'Evil', url: 'javascript:alert(1)', icon: '' }]
    }
    await renderSuspended(AppLayout)

    expect(screen.queryByRole('link', { name: 'Evil' })).not.toBeInTheDocument()
  })
})
