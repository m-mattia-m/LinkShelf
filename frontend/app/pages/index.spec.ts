import { renderSuspended } from '@nuxt/test-utils/runtime'
import { screen } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import IndexPage from './index.vue'

describe('landing page', () => {
  it('renders the welcome title and description', async () => {
    await renderSuspended(IndexPage)

    expect(screen.getByText('Getting started sharing links.')).toBeInTheDocument()
    expect(screen.getByText(/LinkShelf is an Open Source Linktree alternative/)).toBeInTheDocument()
  })

  it('links "Get started" to the app and "Source code" to GitHub', async () => {
    await renderSuspended(IndexPage)

    expect(screen.getByRole('link', { name: /Get started/ })).toHaveAttribute('href', '/app')
    expect(screen.getByRole('link', { name: /Source code/ })).toHaveAttribute('href', 'https://github.com/m-mattia-m/Linkshelf')
  })

  it('renders the features section', async () => {
    await renderSuspended(IndexPage)

    expect(screen.getByText('Everything you need to organize your links')).toBeInTheDocument()
    expect(screen.getByText('Unlimited collections')).toBeInTheDocument()
    expect(screen.getByText('Custom domains')).toBeInTheDocument()
    expect(screen.getByText('Self-host & scale')).toBeInTheDocument()
  })

  it('renders the closing call-to-action with cloud and self-host links', async () => {
    await renderSuspended(IndexPage)

    expect(screen.getByText('Ready to organize your links?')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: /Start with LinkShelf Cloud/ })).toHaveAttribute('href', '/cloud')
    expect(screen.getByRole('link', { name: /Self-host on GitHub/ })).toHaveAttribute('href', 'https://github.com/m-mattia-m/linkshelf')
  })
})
