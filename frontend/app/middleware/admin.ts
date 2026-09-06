// Runs after the global auth middleware, which has already guaranteed the
// caller is authenticated (or redirected them to sign-in) by this point.
// Client-only for the same reason as auth.global.ts: role comes from the
// locally-stored access token, which SSR never has.
export default defineNuxtRouteMiddleware(() => {
  if (!import.meta.client) return

  const authStore = useAuthStore()
  authStore.init()

  if (!authStore.isAdmin) {
    return navigateTo('/app')
  }
})
