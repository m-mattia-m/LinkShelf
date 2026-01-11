import type {PostCreateShelfRequest, ShelfBase} from "~~/api";
import {ShelfApi} from "~~/api";

export const useShelfStore = defineStore('shelfStore', {
  state: () => ({
    shelves: [] as ShelfBase[],
  }),

  actions: {
    async fetch(): Promise<void> {
      this.shelves = await new ShelfApi().listShelves()
    },

    async create(shelf: PostCreateShelfRequest): Promise<void> {
      const newShelf = await new ShelfApi().postCreateShelf(shelf)
      this.shelves.push(newShelf)
    },
  },
})

