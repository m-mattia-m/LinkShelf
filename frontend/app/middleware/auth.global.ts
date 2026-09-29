// Tokens live in localStorage only, so this only runs on the client.
export default defineNuxtRouteMiddleware((to) => {
  if (!import.meta.client) return

  const authStore = useAuthStore()
  authStore.init()

  const isAppRoute = to.path.startsWith('/app')
  const isAuthRoute = to.path.startsWith('/auth')
  // Token-based email links must work even with an unrelated active session.
  const isTokenActionRoute = to.path === '/auth/callback' || to.path === '/auth/verify-email' || to.path === '/auth/set-password' || to.path === '/auth/reset-password'

  if (isAppRoute && !authStore.isAuthenticated) {
    return navigateTo({ path: '/auth/sign-in', query: { redirect: to.fullPath } })
  }

  if (isAuthRoute && !isTokenActionRoute && authStore.isAuthenticated) {
    return navigateTo('/app')
  }
})
