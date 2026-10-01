// Sets 'shelf-host' to the shelf domain the frontend is reached on, or null.
export default defineNuxtRouteMiddleware(async (to) => {
  const shelfHost = useState<string | null | undefined>('shelf-host', () => undefined)

  if (shelfHost.value === undefined) {
    if (import.meta.server) {
      shelfHost.value = await resolveShelfHost(
        requestHost(useRequestHeaders(['host', 'x-forwarded-host'])),
        useRuntimeConfig().public.apiBase
      )
    } else {
      // No server state (client-side navigation), so ask the API.
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
