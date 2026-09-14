import { renderSuspended } from '@nuxt/test-utils/runtime'
import { screen, waitFor } from '@testing-library/vue'
import { beforeEach, describe, expect, it } from 'vitest'
import AppLayout from './AppLayout.vue'

beforeEach(() => {
  useState('settings').value = null
})

describe('AppLayout', () => {
  it('renders the logo linking to the homepage', async () => {
    await renderSuspended(AppLayout)

    const homeLink = screen.getAllByRole('link').find(link => link.getAttribute('href') === '/')
    expect(homeLink).toBeTruthy()
    expect(homeLink?.querySelector('img[src="/logo.svg"]')).toBeInTheDocument()
  })

  it('renders the primary navigation links', async () => {
    await renderSuspended(AppLayout)

    const dashboardLinks = screen.getAllByRole('link', { name: 'Dashboard' })
    expect(dashboardLinks.some(link => link.getAttribute('href') === '/app')).toBe(true)

    const docsLinks = screen.getAllByRole('link', { name: 'Docs' })
    expect(docsLinks.some(link => link.getAttribute('href') === '/docs')).toBe(true)

    const cloudLinks = screen.getAllByRole('link', { name: 'Cloud' })
    expect(cloudLinks.some(link => link.getAttribute('href') === '/cloud')).toBe(true)

    const homeLinks = screen.getAllByRole('link', { name: 'Home' })
    expect(homeLinks.some(link => link.getAttribute('href') === '/')).toBe(true)
  })

  it('renders a GitHub link in the header', async () => {
    await renderSuspended(AppLayout)

    expect(screen.getByRole('link', { name: 'GitHub' })).toHaveAttribute('href', 'https://github.com/m-mattia-m/LinkShelf')
  })

  it('renders the page content passed via the default slot', async () => {
    await renderSuspended(AppLayout, {
      slots: { default: () => 'Page body content' }
    })

    expect(screen.getByText('Page body content')).toBeInTheDocument()
  })

  it('renders the contributors credit line in the footer', async () => {
    await renderSuspended(AppLayout)

    expect(screen.getByText('contributers')).toBeInTheDocument()
    expect(screen.getByText('contributers').closest('a')).toHaveAttribute('href', 'https://github.com/m-mattia-m/LinkShelf/graphs/contributors')
  })

  it('hides footer links for sections the site owner has not enabled', async () => {
    useState('settings').value = {
      aboutShow: false,
      contactShow: true,
      imprintShow: false,
      termsOfUseShow: true,
      privacyPolicyShow: false
    }

    await renderSuspended(AppLayout)

    expect(screen.getByText('About').closest('a')).toHaveClass('hidden')
    expect(screen.getByText('Contact').closest('a')).not.toHaveClass('hidden')
    expect(screen.getByText('Imprint').closest('a')).toHaveClass('hidden')
    expect(screen.getByText('Terms of use').closest('a')).not.toHaveClass('hidden')
    expect(screen.getByText('Privacy policy').closest('a')).toHaveClass('hidden')
  })

  it('shows footer links when settings have not loaded yet', async () => {
    await renderSuspended(AppLayout)

    expect(screen.getByText('About').closest('a')).toHaveClass('hidden')
    expect(screen.getByText('Contact').closest('a')).toHaveClass('hidden')
  })

  it('redirects to the dashboard on mount when the site is configured to do so', async () => {
    useState('settings').value = { redirectToDashboard: true }
    const router = useRouter()
    const pushSpy = vi.spyOn(router, 'push').mockResolvedValue(undefined)
    pushSpy.mockClear()

    await renderSuspended(AppLayout)

    await waitFor(() => {
      expect(pushSpy).toHaveBeenCalledWith('/app')
    })
  })

  it('does not redirect when the site is not configured to redirect', async () => {
    useState('settings').value = { redirectToDashboard: false }
    const router = useRouter()
    const pushSpy = vi.spyOn(router, 'push').mockResolvedValue(undefined)
    pushSpy.mockClear()

    await renderSuspended(AppLayout)

    expect(pushSpy).not.toHaveBeenCalled()
  })
})
