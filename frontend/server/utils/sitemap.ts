const STATIC_PATHS = ['/', '/about', '/cloud']
const LOCALES = ['de', 'de-CH', 'es']

function escapeXml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&apos;')
}

function localized(path: string, code: string): string {
  return path === '/' ? `/${code}` : `/${code}${path}`
}

/** A sitemap listing the static pages and docs pages, with a hreflang alternate per locale. */
export function buildSitemap(origin: string, docsPaths: string[]): string {
  const paths = [...new Set([...STATIC_PATHS, ...docsPaths])].sort()

  const entries = paths.flatMap((path) => {
    const alternates = [
      { lang: 'en', path },
      ...LOCALES.map(code => ({ lang: code, path: localized(path, code) })),
      { lang: 'x-default', path }
    ]
    const links = alternates
      .map(alt => `    <xhtml:link rel="alternate" hreflang="${alt.lang}" href="${escapeXml(origin + alt.path)}"/>`)
      .join('\n')

    return [path, ...LOCALES.map(code => localized(path, code))].map(loc =>
      `  <url>\n    <loc>${escapeXml(origin + loc)}</loc>\n${links}\n  </url>`
    )
  })

  return [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">',
    ...entries,
    '</urlset>',
    ''
  ].join('\n')
}
