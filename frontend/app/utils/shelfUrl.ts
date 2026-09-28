// Top-level words the frontend or backend already routes, so a shelf with
// this path is unreachable while user-based paths are off. Mirrors
// routeReservedNames in backend/internal/domain/username.go.
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
 * The address a shelf is opened at: the instance's origin plus the shelf's
 * path, or https://<domain> for a domain-only shelf. A shelf with both opens
 * at its path.
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
