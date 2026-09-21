// Top-level words the frontend or the backend already answers, so a shelf
// with this path can never be reached while user-based paths are off.
// Mirrors routeReservedNames in backend/internal/domain/username.go (the
// backend additionally reserves the configured assets base path).
const ROUTE_RESERVED_PATHS = new Set([
  'app', 'auth', 'docs', 'cloud', 'about', 'contact', 'imprint', 'privacy-policy', 'terms-of-use',
  'api', 'v1', 'swagger', 'health', 'images'
])

export function isRouteReservedPath(path: string): boolean {
  return ROUTE_RESERVED_PATHS.has(path.trim().toLowerCase())
}

/**
 * The public URL path of a shelf: /<username>/<path> while user-based paths
 * are on, /<path> otherwise. Empty for a shelf that only has a domain.
 */
export function publicShelfPath(shelf: { path?: string, username?: string }, userBasedPaths: boolean): string {
  if (!shelf.path) return ''
  return userBasedPaths ? `/${shelf.username ?? ''}/${shelf.path}` : `/${shelf.path}`
}

/**
 * The address a shelf is opened at: https://<domain> for a shelf that is served
 * on a domain of its own, otherwise the instance's own origin plus the shelf's
 * path. Empty for a shelf that has neither, which the backend doesn't allow
 * anymore.
 *
 * A shelf that has both (created before it had to choose one) opens at its
 * path, the same way the edit form shows it first. A domain is always shown
 * with https:// - putting a shelf on a domain without TLS isn't something to
 * point people towards.
 */
export function publicShelfUrl(
  shelf: { path?: string, domain?: string, username?: string },
  options: { userBasedPaths: boolean, origin: string }
): string {
  const path = publicShelfPath(shelf, options.userBasedPaths)
  if (path) return options.origin + path
  if (shelf.domain) return `https://${shelf.domain}`
  return ''
}
