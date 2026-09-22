import { describe, expect, it } from 'vitest'
import { isRouteReservedPath, publicShelfPath, publicShelfUrl } from './shelfUrl'

describe('publicShelfPath', () => {
  it('is /<path> while user-based paths are off', () => {
    expect(publicShelfPath({ path: 'profile', username: 'alice' }, false)).toBe('/profile')
  })

  it('is /<username>/<path> while user-based paths are on', () => {
    expect(publicShelfPath({ path: 'profile', username: 'alice' }, true)).toBe('/alice/profile')
  })

  it('is empty for a shelf that only has a domain', () => {
    expect(publicShelfPath({ path: '', username: 'alice' }, true)).toBe('')
    expect(publicShelfPath({ path: undefined }, false)).toBe('')
  })
})

describe('isRouteReservedPath', () => {
  it.each(['app', 'auth', 'docs', 'cloud', 'about', 'contact', 'imprint', 'privacy-policy', 'terms-of-use', 'api', 'v1', 'swagger', 'health', 'images'])('reserves %s', (path) => {
    expect(isRouteReservedPath(path)).toBe(true)
  })

  it('ignores case and surrounding whitespace', () => {
    expect(isRouteReservedPath('DOCS')).toBe(true)
    expect(isRouteReservedPath(' app ')).toBe(true)
  })

  it('leaves words that only collide as usernames alone', () => {
    for (const path of ['admin', 'support', 'help', 'profile', 'my-links', 'docs-2']) {
      expect(isRouteReservedPath(path)).toBe(false)
    }
  })
})

describe('publicShelfUrl', () => {
  const origin = 'https://linkshelf.example.com'

  it('is the origin plus /<path> for a path shelf', () => {
    expect(publicShelfUrl({ path: 'profile', username: 'alice' }, { userBasedPaths: false, origin }))
      .toBe('https://linkshelf.example.com/profile')
  })

  it('is the origin plus /<username>/<path> while user-based paths are on', () => {
    expect(publicShelfUrl({ path: 'profile', username: 'alice' }, { userBasedPaths: true, origin }))
      .toBe('https://linkshelf.example.com/alice/profile')
  })

  it('is https:// plus the domain, without a path, for a domain shelf', () => {
    expect(publicShelfUrl({ path: '', domain: 'profile.example.com', username: 'alice' }, { userBasedPaths: false, origin }))
      .toBe('https://profile.example.com')
    expect(publicShelfUrl({ domain: 'profile.example.com:9443' }, { userBasedPaths: true, origin }))
      .toBe('https://profile.example.com:9443')
  })

  it('opens a shelf that has both from before at its path', () => {
    expect(publicShelfUrl({ path: 'profile', domain: 'profile.example.com', username: 'alice' }, { userBasedPaths: false, origin }))
      .toBe('https://linkshelf.example.com/profile')
  })

  it('is empty for a shelf that has neither', () => {
    expect(publicShelfUrl({ path: '', domain: '' }, { userBasedPaths: false, origin })).toBe('')
  })
})
