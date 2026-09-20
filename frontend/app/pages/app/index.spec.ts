import { renderSuspended } from '@nuxt/test-utils/runtime'
import { HttpResponse, http } from 'msw'
import { screen, waitFor } from '@testing-library/vue'
import { beforeEach, describe, expect, it } from 'vitest'
import { ShelfToJSON, StatisticToJSON } from '~~/api'
import { server } from '../../../test/mocks/server'
import { errorResponse } from '../../../test/mocks/handlers'
import { buildShelf, buildStatistic } from '../../../test/mocks/factories'
import { resetOnceCache } from '../../../test/reset-once-cache'
import DashboardPage from './index.vue'

const BASE = 'http://localhost:8085'

beforeEach(() => {
  resetOnceCache()
})

describe('app dashboard page', () => {
  it('shows a loading state, then the fetched statistics and recent shelves', async () => {
    server.use(
      http.get(`${BASE}/v1/statistics`, () => HttpResponse.json(StatisticToJSON(buildStatistic({ shelfNumber: 4, sectionNumber: 7, linkNumber: 12 })))),
      http.get(`${BASE}/v1/shelves`, () => HttpResponse.json([
        ShelfToJSON({ id: 'shelf-1', title: 'My Links', path: 'my-links', description: 'A shelf', domain: '', icon: 'i-lucide-book', theme: {}, themeId: '', themeMissing: false, userId: 'user-1' })
      ]))
    )

    await renderSuspended(DashboardPage)

    await waitFor(() => {
      expect(screen.getByText('My Links')).toBeInTheDocument()
    })

    expect(screen.getByText('4')).toBeInTheDocument()
    expect(screen.getByText('7')).toBeInTheDocument()
    expect(screen.getByText('12')).toBeInTheDocument()
    expect(screen.getByText('A shelf')).toBeInTheDocument()
  })

  it('shows an empty-state message when the user has no shelves', async () => {
    server.use(
      http.get(`${BASE}/v1/statistics`, () => HttpResponse.json(StatisticToJSON(buildStatistic({ shelfNumber: 0, sectionNumber: 0, linkNumber: 0 })))),
      http.get(`${BASE}/v1/shelves`, () => HttpResponse.json([]))
    )

    await renderSuspended(DashboardPage)

    await waitFor(() => {
      expect(screen.getByText('You don\'t have any shelves yet.')).toBeInTheDocument()
    })
  })

  it('only shows the first five recent shelves', async () => {
    server.use(
      http.get(`${BASE}/v1/statistics`, () => HttpResponse.json(StatisticToJSON(buildStatistic()))),
      http.get(`${BASE}/v1/shelves`, () => HttpResponse.json(
        Array.from({ length: 7 }, (_, i) => ShelfToJSON(buildShelf({ id: `shelf-${i}`, title: `Shelf ${i}`, description: '', theme: {}, themeId: '', themeMissing: false, userId: 'user-1' })))
      ))
    )

    await renderSuspended(DashboardPage)

    await waitFor(() => {
      expect(screen.getByText('Shelf 0')).toBeInTheDocument()
    })
    expect(screen.getByText('Shelf 4')).toBeInTheDocument()
    expect(screen.queryByText('Shelf 5')).not.toBeInTheDocument()
  })

  it('stops showing the loading placeholder when the data cannot be loaded', async () => {
    server.use(
      http.get(`${BASE}/v1/statistics`, () => errorResponse(500, 'boom')),
      http.get(`${BASE}/v1/shelves`, () => HttpResponse.json([]))
    )

    await renderSuspended(DashboardPage)

    // The skeleton placeholder must not stay behind when the request fails.
    await waitFor(() => {
      expect(document.querySelector('.animate-pulse')).toBeNull()
    })
  })
})
