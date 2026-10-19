import { describe, expect, it } from 'vitest'
import { buildRobots } from './robots'

describe('buildRobots', () => {
  it('blocks the app and auth routes and points to the sitemap on the main host', () => {
    const body = buildRobots('https://example.com', true)

    expect(body).toContain('Disallow: /app/\n')
    expect(body).toContain('Disallow: /app$\n')
    expect(body).toContain('Disallow: /auth/\n')
    expect(body).toContain('Disallow: /auth$\n')
    expect(body).toContain('Sitemap: https://example.com/sitemap.xml\n')
  })

  it('allows everything on other hosts and when no site URL is set', () => {
    expect(buildRobots('https://example.com', false)).toBe('User-agent: *\nDisallow:\n')
    expect(buildRobots('', false)).toBe('User-agent: *\nDisallow:\n')
  })
})
