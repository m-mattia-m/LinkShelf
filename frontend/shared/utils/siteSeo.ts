// Helpers for the instance's own website (NUXT_PUBLIC_SITE_URL). Everything is off while that is unset.

/** The configured site origin without trailing slashes, or '' when unset or not a valid URL. */
export function siteOrigin(siteUrl: unknown): string {
  if (typeof siteUrl !== 'string') return ''
  const origin = siteUrl.trim().replace(/\/+$/, '')
  if (origin === '') return ''
  try {
    new URL(origin)
    return origin
  } catch {
    return ''
  }
}

/** Whether a request host (as from requestHost) is the host of the configured site URL. */
export function isSiteHost(siteUrl: unknown, host: string): boolean {
  const origin = siteOrigin(siteUrl)
  if (origin === '' || host === '') return false
  return host.toLowerCase() === new URL(origin).host.toLowerCase()
}
