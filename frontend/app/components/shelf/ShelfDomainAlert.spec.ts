import { renderSuspended } from '@nuxt/test-utils/runtime'
import { fireEvent, screen } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import type { DomainStatus } from '~~/api'
import { buildShelf } from '../../../test/mocks/factories'
import ShelfDomainAlert from './ShelfDomainAlert.vue'

const records = [
  { type: 'TXT', name: '_linkshelf-challenge.profile.example.com', value: 'linkshelf-verification=abc' },
  { type: 'CNAME', name: 'profile.example.com', value: 'domains.linkshelf.cloud' }
]

type StatusOverride = Partial<Omit<DomainStatus, 'state'>> & { state?: DomainStatus['state'] }

// "status: null" is a shelf the controller never wrote anything for.
function render(options: { domain?: string, status?: StatusOverride | null } = {}) {
  const domainStatus: DomainStatus | undefined = options.status === null
    ? undefined
    : { state: 'pending', reason: 'txt_missing', message: '', records, ...options.status }

  return renderSuspended(ShelfDomainAlert, {
    props: {
      shelf: buildShelf({
        id: 's1',
        title: 'Profile',
        domain: 'domain' in options ? options.domain : 'profile.example.com',
        domainStatus
      })
    }
  })
}

describe('ShelfDomainAlert', () => {
  it('shows a pending domain on the main screen as a short notice, with the records folded away', async () => {
    await render({})

    const alert = screen.getByTestId('domain-alert')
    expect(alert).toHaveTextContent('Domain verification pending')
    expect(alert).toHaveTextContent('profile.example.com is not reachable yet')
    expect(screen.getByTestId('domain-alert-toggle')).toHaveAttribute('aria-expanded', 'false')
    expect(screen.queryByTestId('domain-status-record')).not.toBeInTheDocument()
  })

  it('unfolds the records, and folds them again', async () => {
    await render({})

    await fireEvent.click(screen.getByRole('button', { name: 'Show DNS records (2)' }))
    expect(screen.getAllByTestId('domain-status-record')).toHaveLength(2)
    expect(screen.getByTestId('domain-alert-toggle')).toHaveAttribute('aria-expanded', 'true')

    await fireEvent.click(screen.getByRole('button', { name: 'Hide DNS records' }))
    expect(screen.queryByTestId('domain-status-record')).not.toBeInTheDocument()
  })

  it('asks for a new look on request', async () => {
    const { emitted } = await render({})

    await fireEvent.click(screen.getByTestId('domain-alert-check'))
    expect(emitted()).toHaveProperty('check')
  })

  it('shows a problem, without a button to check again', async () => {
    await render({ status: { state: 'error', reason: 'taken', records: [] } })

    expect(screen.getByTestId('domain-alert')).toHaveTextContent('There is a problem with your domain')
    expect(screen.queryByTestId('domain-alert-check')).not.toBeInTheDocument()
    expect(screen.queryByTestId('domain-alert-toggle')).not.toBeInTheDocument()
  })

  it('says nothing about a domain which is ready', async () => {
    await render({ status: { state: 'ready', reason: '', records: [] } })

    expect(screen.queryByTestId('domain-alert')).not.toBeInTheDocument()
  })

  it('says nothing without a domain or without a status', async () => {
    await render({ domain: '' })
    expect(screen.queryByTestId('domain-alert')).not.toBeInTheDocument()
  })

  it('says nothing for an instance without a controller', async () => {
    await render({ status: null })
    expect(screen.queryByTestId('domain-alert')).not.toBeInTheDocument()
  })
})
