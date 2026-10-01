// Mirrors routeReservedNames in backend/internal/domain/username.go.
const ROUTE_RESERVED_PATHS = new Set([
  'app', 'auth', 'docs', 'cloud', 'about', 'contact', 'imprint', 'privacy-policy', 'terms-of-use',
  'api', 'v1', 'swagger', 'health', 'images'
])

export function isRouteReservedPath(path: string): boolean {
  return ROUTE_RESERVED_PATHS.has(path.trim().toLowerCase())
}

/** The shelf's public path, or empty for domain-only shelves. */
export function publicShelfPath(shelf: { path?: string, username?: string }, userBasedPaths: boolean): string {
  if (!shelf.path) return ''
  return userBasedPaths ? `/${shelf.username ?? ''}/${shelf.path}` : `/${shelf.path}`
}

/** The URL a shelf opens at: its path, or https://<domain> for domain-only shelves. */
export function publicShelfUrl(
  shelf: { path?: string, domain?: string, username?: string },
  options: { userBasedPaths: boolean, origin: string }
): string {
  const path = publicShelfPath(shelf, options.userBasedPaths)
  if (path) return options.origin + path
  if (shelf.domain) return `https://${shelf.domain}`
  return ''
}
