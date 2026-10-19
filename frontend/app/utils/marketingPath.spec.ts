import { describe, expect, it } from 'vitest'
import { isMarketingPath, stripLocalePrefix } from './marketingPath'

describe('stripLocalePrefix', () => {
  it('removes a locale prefix', () => {
    expect(stripLocalePrefix('/de/about')).toBe('/about')
    expect(stripLocalePrefix('/de-CH/docs/x')).toBe('/docs/x')
    expect(stripLocalePrefix('/es')).toBe('/')
  })

  it('keeps other paths, including ones that merely start with a locale code', () => {
    expect(stripLocalePrefix('/about')).toBe('/about')
    expect(stripLocalePrefix('/design')).toBe('/design')
    expect(stripLocalePrefix('/')).toBe('/')
  })
})

describe('isMarketingPath', () => {
  it.each(['/', '/about', '/cloud', '/contact', '/imprint', '/privacy-policy', '/terms-of-use', '/docs', '/docs/x', '/docs/self-hosting/docker', '/de', '/de/about', '/es/docs/faq', '/about/'])(
    'treats %s as a marketing page', (path) => {
      expect(isMarketingPath(path)).toBe(true)
    }
  )

  it.each(['/app', '/app/shelves', '/auth/sign-in', '/de/app', '/links', '/docsx', '/anything-else'])(
    'does not treat %s as a marketing page', (path) => {
      expect(isMarketingPath(path)).toBe(false)
    }
  )
})
