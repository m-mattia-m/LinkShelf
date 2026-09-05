import type { SettingPageBody } from '~~/api'

export interface SettingKeyValue {
  key: string
  value: string
}

export interface SettingUpdateFailure {
  key: string
  message: string
}

export const useSettingStore = defineStore('settingStore', {
  state: () => ({
    page: null as SettingPageBody | null,
    languageCode: 'en',
    loaded: false
  }),

  actions: {
    async fetch(languageCode: string): Promise<void> {
      const api = useApi()
      this.languageCode = languageCode
      this.page = await api.setting.getPageSettings({ languageCode })
      this.loaded = true
    },

    // Settings are stored one key/language row at a time on the backend, so a
    // single "Save" here fires one PUT per changed key sequentially and
    // collects any failures instead of aborting on the first one.
    async updateMany(languageCode: string, entries: SettingKeyValue[]): Promise<SettingUpdateFailure[]> {
      const api = useApi()
      const failures: SettingUpdateFailure[] = []

      for (const entry of entries) {
        try {
          this.page = await api.setting.putUpdateSetting({
            setting: {
              key: entry.key,
              languageCode,
              value: entry.value
            }
          })
        } catch (err) {
          const { message } = await parseApiError(err)
          failures.push({ key: entry.key, message })
        }
      }

      await this.fetch(languageCode)

      return failures
    }
  }
})
