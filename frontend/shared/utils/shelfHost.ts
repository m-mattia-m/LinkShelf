// A shelf can be served on a domain of its own: when the frontend is reached
// on such a domain, `/` shows that shelf instead of the landing page. This is
// the lookup that decides whether the host of a request is one of those.
//
// It runs on the Nuxt server, once per page request, in front of every render,
// so it is answered from a small in-memory cache: the main host, which is
// almost every request, costs one backend call a minute. Content is never
// cached here, only the yes/no - the shelf itself is fetched fresh by the
// page, so an edit shows up immediately, while a domain that was just added
// or removed takes up to a minute to.

const CACHE_TTL_MS = 60_000
// The Host header is chosen by whoever sends the request, so the cache can't be
// allowed to grow with it.
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

/**
 * The host a request was addressed to, as the visitor typed it. Behind a proxy
 * or load balancer that rewrites Host, the original is in X-Forwarded-Host
 * (possibly a list, first one wins), so that is preferred. Anyone can send that
 * header, but the worst it can do is show a public shelf under another host,
 * so it needs no trusted-proxy list.
 */
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

/**
 * The normalized domain of the shelf served on `host`, or null when the host
 * isn't a shelf's domain - which is the case for the instance's own host, and
 * for anything that couldn't be a domain at all (localhost, an IP, junk), none
 * of which ever reaches the backend.
 *
 * A backend that can't be reached also gives null, and isn't remembered: the
 * main site has to keep working when this lookup can't.
 */
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
