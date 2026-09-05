import type { Section, SectionBase } from '~~/api'

export const useSectionStore = defineStore('sectionStore', {
  state: () => ({
    sections: [] as Section[],
    loaded: false
  }),

  actions: {
    async fetch(shelfId: string): Promise<void> {
      const api = useApi()
      this.sections = (await api.section.getSections({ shelfId })) ?? []
      this.loaded = true
    },

    async create(sectionBase: SectionBase): Promise<Section> {
      const api = useApi()
      const created = await api.section.postCreateSection({ sectionBase })
      await this.fetch(sectionBase.shelfId)
      return created
    },

    async update(sectionId: string, sectionBase: SectionBase): Promise<Section> {
      const api = useApi()
      const updated = await api.section.putUpdateSection({ sectionId, sectionBase })
      await this.fetch(sectionBase.shelfId)
      return updated
    },

    async remove(sectionId: string, shelfId: string): Promise<void> {
      const api = useApi()
      await api.section.deleteSection({ sectionId })
      await this.fetch(shelfId)
    }
  }
})
