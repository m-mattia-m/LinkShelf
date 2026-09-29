// Shares via the 'shelf-host' state whether the frontend is reached on a
// shelf's domain (so `/` shows that shelf) or on the instance's own host
// (null). See shared/utils/shelfHost.ts.
export default defineNuxtRouteMiddleware(async (to) => {
  const shelfHost = useState<string | null | undefined>('shelf-host', () => undefined)

  if (shelfHost.value === undefined) {
    if (import.meta.server) {
      shelfHost.value = await resolveShelfHost(
        requestHost(useRequestHeaders(['host', 'x-forwarded-host'])),
        useRuntimeConfig().public.apiBase
      )
    } else {
      // Only reached when the page didn't come from the server (no state was
      // handed over), so ask the API from here, the same way.
      shelfHost.value = await resolveShelfHostInBrowser()
    }
  }

  if (shelfHost.value && to.path === '/') setPageLayout('links')
})

async function resolveShelfHostInBrowser(): Promise<string | null> {
  const domain = normalizeShelfDomain(window.location.host)
  if (domain === '' || validateShelfDomain(domain) !== null) return null

  try {
    await useApi().shelf.getPublicShelfByDomain({ domain })
    return domain
  } catch {
    return null
  }
}
