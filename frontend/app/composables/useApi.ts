import {
  AuthApi,
  Configuration,
  LinkApi,
  SectionApi,
  SettingApi,
  ShelfApi,
  StatisticApi,
  ThemeApi,
  UserApi
} from '~~/api'

function authorizationHeader(headers: HeadersInit | undefined): string | null {
  if (!headers) return null
  return new Headers(headers).get('Authorization')
}

export function useApi() {
  const runtimeConfig = useRuntimeConfig()
  const authStore = useAuthStore()

  // Only requests that carried a Bearer token are retried after a refresh.
  // Auth endpoints never send one, so a failing refresh can't recurse.
  const customFetch: typeof fetch = async (input, init) => {
    const response = await fetch(input, init)

    const sentAuthorization = authorizationHeader(init?.headers)
    if (response.status !== 401 || sentAuthorization === null || !authStore.refreshToken) {
      return response
    }

    // A request that was in flight at the same time may already have renewed
    // the tokens - only refresh if the token this one used is still current.
    if (sentAuthorization === `Bearer ${authStore.accessToken}`) {
      const refreshed = await authStore.refresh()
      if (!refreshed) return response
    } else if (!authStore.accessToken) {
      return response
    }

    // Set the header on a Headers object so there's exactly one Authorization
    // key; a differently-cased duplicate would send both tokens.
    const headers = new Headers(init?.headers)
    headers.set('Authorization', `Bearer ${authStore.accessToken}`)
    return fetch(input, { ...init, headers: Object.fromEntries(headers.entries()) })
  }

  const configuration = new Configuration({
    basePath: runtimeConfig.public.apiBase,
    accessToken: () => authStore.accessToken ?? '',
    fetchApi: customFetch
  })

  return {
    auth: new AuthApi(configuration),
    shelf: new ShelfApi(configuration),
    section: new SectionApi(configuration),
    link: new LinkApi(configuration),
    setting: new SettingApi(configuration),
    user: new UserApi(configuration),
    statistic: new StatisticApi(configuration),
    theme: new ThemeApi(configuration)
  }
}
