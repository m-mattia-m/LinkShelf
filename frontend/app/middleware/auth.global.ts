// Tokens live in localStorage only (no cookie/session), so there's nothing to
// check during SSR - this only enforces on the client, after hydration has
// had a chance to restore a saved session.
export default defineNuxtRouteMiddleware((to) => {
  if (!import.meta.client) return

  const authStore = useAuthStore()
  authStore.init()

  const isAppRoute = to.path.startsWith('/app')
  const isAuthRoute = to.path.startsWith('/auth')

  if (isAppRoute && !authStore.isAuthenticated) {
    return navigateTo({ path: '/auth/sign-in', query: { redirect: to.fullPath } })
  }

  if (isAuthRoute && to.path !== '/auth/callback' && authStore.isAuthenticated) {
    return navigateTo('/app')
  }
})
