import type { Shelf, ShelfBase } from '~~/api'

export const useShelfStore = defineStore('shelfStore', {
  state: () => ({
    shelves: [] as Shelf[],
    loaded: false
  }),

  actions: {
    async fetch(): Promise<void> {
      const api = useApi()
      this.shelves = (await api.shelf.listShelves()) ?? []
      this.loaded = true
    },

    async create(shelfBase: ShelfBase): Promise<Shelf> {
      const api = useApi()
      const created = await api.shelf.postCreateShelf({ shelfBase })
      await this.fetch()
      return created
    },

    async update(shelfId: string, shelfBase: ShelfBase): Promise<Shelf> {
      const api = useApi()
      const updated = await api.shelf.putUpdateShelf({ shelfId, shelfBase })
      await this.fetch()
      return updated
    },

    async remove(shelfId: string): Promise<void> {
      const api = useApi()
      await api.shelf.deleteShelf({ shelfId })
      await this.fetch()
    },

    async getById(shelfId: string): Promise<Shelf> {
      const api = useApi()
      return api.shelf.getShelfById({ shelfId })
    }
  }
})
