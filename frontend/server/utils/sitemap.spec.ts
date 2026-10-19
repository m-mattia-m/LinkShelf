import { describe, expect, it } from 'vitest'
import { buildSitemap } from './sitemap'

const locs = (xml: string) => [...xml.matchAll(/<loc>([^<]*)<\/loc>/g)].map(m => m[1])

describe('buildSitemap', () => {
  const xml = buildSitemap('https://example.com', ['/docs', '/docs/self-hosting/docker'])

  it('lists the static and docs pages as absolute URLs', () => {
    const urls = locs(xml)

    expect(urls).toContain('https://example.com/')
    expect(urls).toContain('https://example.com/about')
    expect(urls).toContain('https://example.com/cloud')
    expect(urls).toContain('https://example.com/docs')
    expect(urls).toContain('https://example.com/docs/self-hosting/docker')
    expect(urls).not.toContain('https://example.com/imprint')
  })

  it('adds locale-prefixed URLs with hreflang alternates', () => {
    expect(locs(xml)).toContain('https://example.com/de-CH/about')
    expect(locs(xml)).toContain('https://example.com/es')
    expect(xml).toContain('hreflang="de-CH" href="https://example.com/de-CH/about"')
    expect(xml).toContain('hreflang="x-default" href="https://example.com/about"')
  })

  it('has no duplicate URLs', () => {
    const urls = locs(buildSitemap('https://example.com', ['/about', '/docs', '/docs']))

    expect(new Set(urls).size).toBe(urls.length)
  })

  it('escapes XML special characters', () => {
    const out = buildSitemap('https://example.com', ['/docs/a&b<c>'])

    expect(out).toContain('/docs/a&amp;b&lt;c&gt;')
    expect(out).not.toContain('a&b')
  })

  it('declares the sitemap namespace and no invented fields', () => {
    expect(xml).toContain('<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"')
    expect(xml).not.toMatch(/lastmod|priority|changefreq/)
  })
})
