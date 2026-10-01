const SAFE_PROTOCOLS = new Set(['http:', 'https:'])

/** Returns the href for an absolute http(s) URL, otherwise undefined. */
export function safeHref(url: string | null | undefined): string | undefined {
  if (!url) return undefined
  try {
    // The browser parses hrefs the same way, so http(s) is safe.
    return SAFE_PROTOCOLS.has(new URL(url).protocol) ? url : undefined
  } catch {
    return undefined
  }
}
