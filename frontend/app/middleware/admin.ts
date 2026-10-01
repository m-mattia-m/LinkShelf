// Client-only: the role comes from the stored access token.
export default defineNuxtRouteMiddleware(() => {
  if (!import.meta.client) return

  const authStore = useAuthStore()
  authStore.init()

  if (!authStore.isAdmin) {
    return navigateTo('/app')
  }
})
