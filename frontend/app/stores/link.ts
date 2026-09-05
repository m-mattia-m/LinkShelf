import type { Link, LinkBase } from '~~/api'

export const useLinkStore = defineStore('linkStore', {
  state: () => ({
    links: [] as Link[],
    loaded: false,
    currentShelfId: null as string | null
  }),

  getters: {
    bySectionId: (state) => {
      const map = new Map<string, Link[]>()
      for (const link of state.links) {
        const list = map.get(link.sectionId) ?? []
        list.push(link)
        map.set(link.sectionId, list)
      }
      return map
    }
  },

  actions: {
    async fetch(shelfId: string): Promise<void> {
      const api = useApi()
      this.links = (await api.link.getLinks({ shelfId })) ?? []
      this.loaded = true
      this.currentShelfId = shelfId
    },

    async refetch(): Promise<void> {
      if (this.currentShelfId) {
        await this.fetch(this.currentShelfId)
      }
    },

    async create(linkBase: LinkBase): Promise<Link> {
      const api = useApi()
      const created = await api.link.postCreateLink({ linkBase })
      await this.refetch()
      return created
    },

    async update(linkId: string, linkBase: LinkBase): Promise<Link> {
      const api = useApi()
      const updated = await api.link.putUpdateLink({ linkId, linkBase, shelfId: this.currentShelfId ?? undefined })
      await this.refetch()
      return updated
    },

    async remove(linkId: string): Promise<void> {
      const api = useApi()
      await api.link.deleteLink({ linkId, shelfId: this.currentShelfId ?? undefined })
      await this.refetch()
    }
  }
})
