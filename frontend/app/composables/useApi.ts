import {
  Configuration,
  LinkApi,
  SectionApi,
  SettingApi,
  ShelfApi,
  UserApi
} from '~~/api'

let configuration: Configuration | undefined

function getConfiguration(): Configuration {
  if (!configuration) {
    const runtimeConfig = useRuntimeConfig()
    configuration = new Configuration({ basePath: runtimeConfig.public.apiBase })
  }
  return configuration
}

export function useApi() {
  const config = getConfiguration()

  return {
    shelf: new ShelfApi(config),
    section: new SectionApi(config),
    link: new LinkApi(config),
    setting: new SettingApi(config),
    user: new UserApi(config)
  }
}
