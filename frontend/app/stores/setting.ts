import type { SettingPageBody, SettingUpdateFailure } from '~~/api'

export interface SettingKeyValue {
  key: string
  value: string
}

export type { SettingUpdateFailure }

// Current values of the loaded page, to send only changed keys.
function currentValue(page: SettingPageBody | null, key: string): string {
  switch (key) {
    case 'about': return page?.about ?? ''
    case 'about_show': return String(page?.aboutShow ?? false)
    case 'contact': return page?.contact ?? ''
    case 'contact_show': return String(page?.contactShow ?? false)
    case 'imprint': return page?.imprint ?? ''
    case 'imprint_show': return String(page?.imprintShow ?? false)
    case 'terms_of_use': return page?.termsOfUse ?? ''
    case 'terms_of_use_show': return String(page?.termsOfUseShow ?? false)
    case 'privacy_policy': return page?.privacyPolicy ?? ''
    case 'privacy_policy_show': return String(page?.privacyPolicyShow ?? false)
    case 'redirect_to_dashboard': return String(page?.redirectToDashboard ?? false)
    default: return ''
  }
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

    // Sends changed keys in one request. Never throws; errors are reported as failures.
    async updateMany(languageCode: string, entries: SettingKeyValue[]): Promise<SettingUpdateFailure[]> {
      const api = useApi()

      const changed = entries.filter(entry => currentValue(this.page, entry.key) !== entry.value)
      if (changed.length === 0) {
        return []
      }

      try {
        const result = await api.setting.putUpdateSettingsBatch({
          settingBatchRequestBody: {
            settings: changed.map(entry => ({ key: entry.key, languageCode, value: entry.value }))
          }
        })

        this.page = result.settings
        this.languageCode = languageCode

        return result.failures ?? []
      } catch (err) {
        const { message } = await parseApiError(err)
        return changed.map(entry => ({ key: entry.key, languageCode, reason: message }))
      }
    }
  }
})
