// When the frontend is reached on the domain of a shelf, `/` shows that shelf
// instead of the landing page. This finds out which host it is being reached
// on and shares the answer through the 'shelf-host' state: the shelf's domain,
// or null for the instance's own host. See shared/utils/shelfHost.ts.
//
// Everything else on such a host falls through to the normal pages.
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
