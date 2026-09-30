const SAFE_PROTOCOLS = new Set(['http:', 'https:'])

/**
 * The href to render for a user-supplied link, or undefined when it must not
 * be clickable. Only absolute http(s) URLs pass: Vue binds :href verbatim and
 * does not filter schemes, so a stored "javascript:..." value would run as
 * script on this origin when clicked. The backend already rejects those; this
 * is the second line of defence for anything that gets past it.
 */
export function safeHref(url: string | null | undefined): string | undefined {
  if (!url) return undefined
  try {
    // The browser resolves an href with this same URL parser, so a value
    // whose parsed protocol is http(s) is safe to bind unchanged.
    return SAFE_PROTOCOLS.has(new URL(url).protocol) ? url : undefined
  } catch {
    return undefined
  }
}
