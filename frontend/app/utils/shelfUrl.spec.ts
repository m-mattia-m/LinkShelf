import { describe, expect, it } from 'vitest'
import { isRouteReservedPath, publicShelfPath } from './shelfUrl'

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
