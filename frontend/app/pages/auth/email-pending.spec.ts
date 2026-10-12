import { renderSuspended } from '@nuxt/test-utils/runtime'
import { screen } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import EmailPendingPage from './email-pending.vue'

describe('email pending page', () => {
  it('explains that verification is pending and links to sign in', async () => {
    await renderSuspended(EmailPendingPage, { route: { path: '/auth/email-pending' } })

    expect(screen.getByText('Verify your email')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Sign in' })).toHaveAttribute('href', '/auth/sign-in')
  })
})
