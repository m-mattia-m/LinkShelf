// Runs after auth.global.ts. Client-only because the role comes from the
// locally stored access token.
export default defineNuxtRouteMiddleware(() => {
  if (!import.meta.client) return

  const authStore = useAuthStore()
  authStore.init()

  if (!authStore.isAdmin) {
    return navigateTo('/app')
  }
})
