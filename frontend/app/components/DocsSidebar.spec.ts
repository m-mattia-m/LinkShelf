import { mockNuxtImport, renderSuspended } from '@nuxt/test-utils/runtime'
import { screen } from '@testing-library/vue'
import { beforeEach, describe, expect, it } from 'vitest'
import DocsSidebar from './DocsSidebar.vue'

const docsState = vi.hoisted(() => ({ items: [] as Record<string, unknown>[] }))

mockNuxtImport('queryCollection', () => {
  return () => ({
    all: () => Promise.resolve(docsState.items)
  })
})

// useAsyncData caches its result under an auto-generated key that is stable
// across renders within the same Nuxt app (one app per test file here) -
// clear it before every test so each one actually re-invokes the mocked
// queryCollection() above instead of reusing a previous test's cached data.
beforeEach(() => {
  clearNuxtData()
})

describe('DocsSidebar', () => {
  it('renders top-level pages and groups nested pages under their section', async () => {
    docsState.items = [
      { path: '/docs', title: 'Introduction', navigation: true, meta: { icon: 'i-lucide-house' } },
      { path: '/docs/guide', title: 'Guide', navigation: true, meta: { icon: 'i-lucide-book', order: 1 } },
      { path: '/docs/guide/install', title: 'Install', navigation: true, meta: {} },
      { path: '/docs/guide/config', title: 'Configuration', navigation: true, meta: {} }
    ]

    await renderSuspended(DocsSidebar)

    expect(screen.getByText('Introduction')).toBeInTheDocument()
    expect(screen.getByText('Guide')).toBeInTheDocument()
    expect(screen.getByText('Install')).toBeInTheDocument()
    expect(screen.getByText('Configuration')).toBeInTheDocument()
  })

  it('excludes pages that are not marked for navigation', async () => {
    docsState.items = [
      { path: '/docs', title: 'Introduction', navigation: true, meta: {} },
      { path: '/docs/hidden-page', title: 'Hidden Page', navigation: false, meta: {} }
    ]

    await renderSuspended(DocsSidebar)

    expect(screen.getByText('Introduction')).toBeInTheDocument()
    expect(screen.queryByText('Hidden Page')).not.toBeInTheDocument()
  })

  it('sorts sibling pages by their "order" meta, falling back to alphabetical order', async () => {
    docsState.items = [
      { path: '/docs', title: 'Introduction', navigation: true, meta: {} },
      { path: '/docs/zeta', title: 'Zeta Section', navigation: true, meta: { order: 1 } },
      { path: '/docs/alpha', title: 'Alpha Section', navigation: true, meta: { order: 2 } }
    ]

    await renderSuspended(DocsSidebar)

    const labels = screen.getAllByText(/Introduction|Zeta Section|Alpha Section/).map(el => el.textContent)
    expect(labels).toEqual(['Introduction', 'Zeta Section', 'Alpha Section'])
  })

  it('creates a section entry from nested pages even without an explicit section-root page', async () => {
    docsState.items = [
      { path: '/docs', title: 'Introduction', navigation: true, meta: {} },
      { path: '/docs/api/users', title: 'Users API', navigation: true, meta: {} }
    ]

    await renderSuspended(DocsSidebar)

    // With no dedicated "/docs/api" page, the "api" section falls back to
    // its only child's title - so "Users API" appears both as the section's
    // accordion trigger and as the nested link inside it.
    expect(screen.getByRole('button', { name: 'Users API' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Users API' })).toBeInTheDocument()
  })

  it('only marks the link of the current page as active, not "/docs" on every docs page', async () => {
    docsState.items = [
      { path: '/docs', title: 'Introduction', navigation: true, meta: {} },
      { path: '/docs/guide', title: 'Guide', navigation: true, meta: { order: 1 } },
      { path: '/docs/guide/install', title: 'Install', navigation: true, meta: {} }
    ]

    await renderSuspended(DocsSidebar, { route: '/docs/guide/install' })

    expect(screen.getByRole('link', { name: 'Install' })).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('link', { name: 'Introduction' })).not.toHaveAttribute('aria-current')
  })
})
