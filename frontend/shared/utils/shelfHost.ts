// Resolves whether a request's host is a shelf domain; results are cached for a minute.

const CACHE_TTL_MS = 60_000
// The Host header is client-controlled, so the cache size is capped.
const CACHE_MAX_ENTRIES = 1024
const REQUEST_TIMEOUT_MS = 3_000

interface CacheEntry {
  // The normalized domain when a shelf is served on it, null when none is.
  domain: string | null
  expires: number
}

const cache = new Map<string, CacheEntry>()

interface ResolveOptions {
  fetchFn?: typeof fetch
  now?: () => number
}

export function clearShelfHostCache() {
  cache.clear()
}

/** The request host, preferring X-Forwarded-Host. */
export function requestHost(headers: Record<string, string | string[] | undefined>): string {
  const pick = (name: string) => {
    const value = headers[name]
    return (Array.isArray(value) ? value[0] : value)?.split(',')[0]?.trim() ?? ''
  }
  return pick('x-forwarded-host') || pick('host')
}

function remember(host: string, domain: string | null, now: number) {
  if (cache.size >= CACHE_MAX_ENTRIES) {
    for (const [key, entry] of cache) {
      if (entry.expires <= now) cache.delete(key)
    }
    // Still full of live entries: start over rather than refuse to cache.
    if (cache.size >= CACHE_MAX_ENTRIES) cache.clear()
  }
  cache.set(host, { domain, expires: now + CACHE_TTL_MS })
}

/** The normalized shelf domain served on host, or null. Backend errors aren't cached. */
export async function resolveShelfHost(host: string, apiBase: string, options: ResolveOptions = {}): Promise<string | null> {
  const fetchFn = options.fetchFn ?? fetch
  const now = options.now ?? Date.now

  const domain = normalizeShelfDomain(host)
  if (domain === '' || validateShelfDomain(domain) !== null) return null

  const cached = cache.get(domain)
  if (cached && cached.expires > now()) return cached.domain

  try {
    const response = await fetchFn(
      `${apiBase.replace(/\/+$/, '')}/v1/shelves/by-domain/${encodeURIComponent(domain)}`,
      { headers: { accept: 'application/json' }, signal: AbortSignal.timeout(REQUEST_TIMEOUT_MS) }
    )

    if (response.status === 404) {
      remember(domain, null, now())
      return null
    }
    if (!response.ok) return null

    remember(domain, domain, now())
    return domain
  } catch {
    return null
  }
}
