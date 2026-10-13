import { renderSuspended } from '@nuxt/test-utils/runtime'
import { fireEvent, screen } from '@testing-library/vue'
import { describe, expect, it, vi } from 'vitest'
import type { DomainStatus } from '~~/api'
import ShelfDomainStatus from './ShelfDomainStatus.vue'

function render(status: Partial<DomainStatus>) {
  return renderSuspended(ShelfDomainStatus, {
    props: { status: { state: 'pending', reason: '', message: '', records: [], ...status } }
  })
}

describe('ShelfDomainStatus', () => {
  it('shows the state as a badge', async () => {
    await render({ state: 'ready' })

    expect(screen.getByTestId('domain-status-state')).toHaveTextContent('Ready')
  })

  it('translates a known reason', async () => {
    await render({ reason: 'txt_missing', message: 'ignored' })

    expect(screen.getByTestId('domain-status-reason')).toHaveTextContent('The verification record was not found yet.')
  })

  it('falls back to the message for a reason it does not know', async () => {
    await render({ state: 'error', reason: 'something_new', message: 'Quota exceeded at the registrar' })

    expect(screen.getByTestId('domain-status-reason')).toHaveTextContent('Quota exceeded at the registrar')
  })

  it('shows an unknown state as it is', async () => {
    await render({ state: 'weird' as DomainStatus['state'] })

    expect(screen.getByTestId('domain-status-state')).toHaveTextContent('weird')
  })

  it('lists the DNS records as they are', async () => {
    await render({
      records: [
        { type: 'TXT', name: '_linkshelf-challenge.profile.example.com', value: 'linkshelf-verification=abc' },
        { type: 'CNAME', name: 'profile.example.com', value: 'domains.linkshelf.cloud' }
      ]
    })

    const records = screen.getAllByTestId('domain-status-record')
    expect(records).toHaveLength(2)
    expect(records[0]).toHaveTextContent('TXT')
    expect(records[0]).toHaveTextContent('_linkshelf-challenge.profile.example.com')
    expect(records[0]).toHaveTextContent('linkshelf-verification=abc')
    expect(records[1]).toHaveTextContent('domains.linkshelf.cloud')
  })

  it('shows no record list without records', async () => {
    await render({ records: [] })

    expect(screen.queryByTestId('domain-status-record')).not.toBeInTheDocument()
  })

  it('keeps type, name and value apart, so they can be told apart and copied one by one', async () => {
    await render({
      records: [{ type: 'TXT', name: '_linkshelf-challenge.profile.example.com', value: 'linkshelf-verification=abc' }]
    })

    // Never run together as "TXTname_linkshelf-challenge...": each is its own element, the type a badge
    // next to the title and name and value with a label of their own.
    expect(screen.getByTestId('domain-status-record-type')).toHaveTextContent(/^TXT$/)
    expect(screen.getByTestId('domain-status-record-name')).toHaveTextContent(/^_linkshelf-challenge\.profile\.example\.com$/)
    expect(screen.getByTestId('domain-status-record-value')).toHaveTextContent(/^linkshelf-verification=abc$/)
    expect(screen.getByText('Name')).toBeInTheDocument()
    expect(screen.getByText('Value')).toBeInTheDocument()
  })

  it('says what each record is for, and warns that a CNAME is not a TXT', async () => {
    await render({
      records: [
        { type: 'TXT', name: '_linkshelf-challenge.profile.example.com', value: 'v' },
        { type: 'CNAME', name: 'profile.example.com', value: 'domains.linkshelf.cloud' }
      ]
    })

    const [verify, route] = screen.getAllByTestId('domain-status-record')
    expect(verify).toHaveTextContent('1. Prove that the domain is yours')
    expect(verify).toHaveTextContent('A CNAME record does not work for this one')
    expect(route).toHaveTextContent('2. Send the traffic to LinkShelf')
  })

  it('titles a record of an unknown type with its type', async () => {
    await render({ records: [{ type: 'SRV', name: 'a', value: 'b' }] })

    expect(screen.getByTestId('domain-status-record')).toHaveTextContent('1. SRV')
  })

  it('copies a name and a value on their own', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', { value: { writeText }, configurable: true })
    await render({ records: [{ type: 'TXT', name: '_linkshelf-challenge.a.example.com', value: 'linkshelf-verification=abc' }] })

    await fireEvent.click(screen.getByRole('button', { name: 'Copy: Name' }))
    await fireEvent.click(screen.getByRole('button', { name: 'Copy: Value' }))

    expect(writeText).toHaveBeenNthCalledWith(1, '_linkshelf-challenge.a.example.com')
    expect(writeText).toHaveBeenNthCalledWith(2, 'linkshelf-verification=abc')
  })

  it('has no state badge of its own when it is embedded', async () => {
    await renderSuspended(ShelfDomainStatus, {
      props: { embedded: true, status: { state: 'pending', reason: 'txt_missing', message: '', records: [] } }
    })

    expect(screen.queryByTestId('domain-status-state')).not.toBeInTheDocument()
    expect(screen.getByTestId('domain-status-reason')).toHaveTextContent('The verification record was not found yet.')
  })
})
