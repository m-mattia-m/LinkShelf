import {
  AuthApi,
  Configuration,
  LinkApi,
  SectionApi,
  SettingApi,
  ShelfApi,
  StatisticApi,
  UserApi
} from '~~/api'

function hasAuthorizationHeader(headers: HeadersInit | undefined): boolean {
  if (!headers) return false
  return new Headers(headers).has('Authorization')
}

export function useApi() {
  const runtimeConfig = useRuntimeConfig()
  const authStore = useAuthStore()

  // Only requests that already carried a Bearer token (i.e. protected
  // endpoints) are eligible for a refresh-and-retry - auth endpoints
  // (login/refresh/logout/oidc) never send one, so this can't recurse into
  // itself when a refresh attempt is what 401s.
  const customFetch: typeof fetch = async (input, init) => {
    const response = await fetch(input, init)

    if (response.status !== 401 || !hasAuthorizationHeader(init?.headers) || !authStore.refreshToken) {
      return response
    }

    const refreshed = await authStore.refresh()
    if (!refreshed) return response

    const headers = new Headers(init?.headers)
    headers.set('Authorization', `Bearer ${authStore.accessToken}`)
    return fetch(input, { ...init, headers })
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
    statistic: new StatisticApi(configuration)
  }
}
