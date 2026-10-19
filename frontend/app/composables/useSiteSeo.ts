// Whether this request is for the instance's own website (not a shelf on a custom domain),
// plus a helper for absolute URLs. Everything is off unless NUXT_PUBLIC_SITE_URL is set.
export function useSiteSeo() {
  const origin = siteOrigin(useRuntimeConfig().public.siteUrl)
  const host = import.meta.server
    ? requestHost(useRequestHeaders(['host', 'x-forwarded-host']))
    : window.location.host
  const isMainHost = isSiteHost(origin, host)

  return {
    origin,
    isMainHost,
    absolute: (path: string) => `${origin}${path.startsWith('/') ? path : `/${path}`}`
  }
}
