import type { User, UserBase, UserCreate, UserRequestBodyOnlyPassword } from '~~/api'

export const useUserStore = defineStore('userStore', {
  state: () => ({
    users: [] as User[],
    loaded: false
  }),

  actions: {
    async fetch(): Promise<void> {
      const api = useApi()
      this.users = (await api.user.listUsers()) ?? []
      this.loaded = true
    },

    async create(userCreate: UserCreate): Promise<User> {
      const api = useApi()
      const created = await api.user.postCreateUser({ userCreate })
      await this.fetch()
      return created
    },

    async update(userId: string, userBase: UserBase): Promise<User> {
      const api = useApi()
      const updated = await api.user.putUpdateUser({ userId, userBase })
      await this.fetch()
      return updated
    },

    async patchPassword(userId: string, body: UserRequestBodyOnlyPassword): Promise<void> {
      const api = useApi()
      await api.user.patchUserPassword({ userId, userRequestBodyOnlyPassword: body })
    },

    async remove(userId: string): Promise<void> {
      const api = useApi()
      await api.user.deleteUser({ userId })
      await this.fetch()
    }
  }
})
