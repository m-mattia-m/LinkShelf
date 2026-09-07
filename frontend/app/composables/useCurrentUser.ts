import type { User } from '~~/api'

export function useCurrentUser() {
  const authStore = useAuthStore()

  async function ensureUserId(): Promise<string> {
    authStore.init()
    if (!authStore.userId) {
      throw new Error('Not authenticated')
    }
    return authStore.userId
  }

  async function ensureUser(): Promise<User> {
    authStore.init()
    if (!authStore.user) {
      await authStore.fetchUser()
    }
    if (!authStore.user) {
      throw new Error('Not authenticated')
    }
    return authStore.user
  }

  return {
    userId: computed(() => authStore.userId),
    user: computed(() => authStore.user),
    ensureUserId,
    ensureUser
  }
}
