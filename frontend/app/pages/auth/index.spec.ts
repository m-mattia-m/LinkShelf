import { renderSuspended } from '@nuxt/test-utils/runtime'
import { waitFor } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import AuthIndexPage from './index.vue'

describe('auth index page', () => {
  it('redirects to the sign-in page', async () => {
    await renderSuspended(AuthIndexPage)

    const route = useRoute()
    await waitFor(() => {
      expect(route.fullPath).toBe('/auth/sign-in')
    })
  })
})
