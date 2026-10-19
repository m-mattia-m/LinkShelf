const LOCALE_PREFIX = /^\/(?:de-CH|de|es)(?=\/|$)/

/** The path without its locale prefix ('/de/about' -> '/about'). */
export function stripLocalePrefix(path: string): string {
  return path.replace(LOCALE_PREFIX, '') || '/'
}

const MARKETING_PATHS = new Set(['/', '/about', '/cloud', '/contact', '/imprint', '/privacy-policy', '/terms-of-use'])

/** Whether the path is a page of the instance's own website that is rendered on the server. */
export function isMarketingPath(path: string): boolean {
  const stripped = stripLocalePrefix(path.split(/[?#]/)[0] ?? '').replace(/(.)\/+$/, '$1')
  return MARKETING_PATHS.has(stripped) || stripped === '/docs' || stripped.startsWith('/docs/')
}
