/** robots.txt body: restrictive on the instance's own website, allow-all everywhere else. */
export function buildRobots(origin: string, isMainHost: boolean): string {
  if (!isMainHost) {
    // As the previous static file: shelves opt out with their own noindex.
    return 'User-agent: *\nDisallow:\n'
  }

  // "/app/" plus "/app$" rather than "/app", which would also block shelf paths such as /apple-links.
  return [
    'User-agent: *',
    'Disallow: /app/',
    'Disallow: /app$',
    'Disallow: /auth/',
    'Disallow: /auth$',
    '',
    `Sitemap: ${origin}/sitemap.xml`,
    ''
  ].join('\n')
}
