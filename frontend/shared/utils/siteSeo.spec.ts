import { describe, expect, it } from 'vitest'
import { isSiteHost, siteOrigin } from './siteSeo'

describe('siteOrigin', () => {
  it('strips whitespace and trailing slashes', () => {
    expect(siteOrigin('  https://example.com// ')).toBe('https://example.com')
  })

  it('is empty when unset or invalid', () => {
    expect(siteOrigin('')).toBe('')
    expect(siteOrigin(undefined)).toBe('')
    expect(siteOrigin('not a url')).toBe('')
  })
})

describe('isSiteHost', () => {
  it('matches the host of the site URL', () => {
    expect(isSiteHost('https://example.com/', 'example.com')).toBe(true)
    expect(isSiteHost('https://example.com', 'EXAMPLE.com')).toBe(true)
    expect(isSiteHost('http://localhost:3000', 'localhost:3000')).toBe(true)
  })

  it('rejects other hosts', () => {
    expect(isSiteHost('https://example.com', 'shelf.example.org')).toBe(false)
    expect(isSiteHost('http://localhost:3000', 'localhost')).toBe(false)
  })

  it('is never true when the site URL is unset', () => {
    expect(isSiteHost('', 'example.com')).toBe(false)
    expect(isSiteHost(undefined, '')).toBe(false)
  })
})
