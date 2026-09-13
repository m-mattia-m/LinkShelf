import type { Theme, ThemeBase } from '~~/api'

export const useThemeStore = defineStore('themeStore', {
  state: () => ({
    instance: [] as Theme[],
    mine: [] as Theme[],
    loaded: false
  }),

  actions: {
    async fetch(): Promise<void> {
      const api = useApi()
      const grouped = await api.theme.listThemes()
      this.instance = grouped.instance ?? []
      this.mine = grouped.mine ?? []
      this.loaded = true
    },

    async create(themeBase: ThemeBase): Promise<Theme> {
      const api = useApi()
      const created = await api.theme.postCreateTheme({ themeBase })
      await this.fetch()
      return created
    },

    async update(themeId: string, themeBase: ThemeBase): Promise<Theme> {
      const api = useApi()
      const updated = await api.theme.putUpdateTheme({ themeId, themeBase })
      await this.fetch()
      return updated
    },

    async remove(themeId: string): Promise<void> {
      const api = useApi()
      await api.theme.deleteTheme({ themeId })
      await this.fetch()
    }
  }
})
