// Tokens live in localStorage only (no cookie/session), so there's nothing to
// check during SSR - this only enforces on the client, after hydration has
// had a chance to restore a saved session.
export default defineNuxtRouteMiddleware((to) => {
  if (!import.meta.client) return

  const authStore = useAuthStore()
  authStore.init()

  const isAppRoute = to.path.startsWith('/app')
  const isAuthRoute = to.path.startsWith('/auth')
  // These complete a token-based email link, which must work regardless of
  // whether the browser happens to have an unrelated active session (e.g. an
  // admin testing an invite, or a user checking the link on a device where
  // they're logged into a different account) - same reasoning that already
  // exempts the OIDC callback below.
  const isTokenActionRoute = to.path === '/auth/callback' || to.path === '/auth/verify-email' || to.path === '/auth/set-password'

  if (isAppRoute && !authStore.isAuthenticated) {
    return navigateTo({ path: '/auth/sign-in', query: { redirect: to.fullPath } })
  }

  if (isAuthRoute && !isTokenActionRoute && authStore.isAuthenticated) {
    return navigateTo('/app')
  }
})
