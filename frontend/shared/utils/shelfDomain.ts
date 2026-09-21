// A shelf can be served on a domain of its own. These are the rules for what
// such a domain looks like, kept in step with NormalizeDomain / ValidateDomain
// in backend/internal/domain/shelf_domain.go - the backend has the final say,
// this only lets the form and the host lookup say no without a round trip.

const LABEL = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/
// Letters only, so an IP address can never pass, or a punycode international TLD.
const TLD = /^([a-z]{2,}|xn--[a-z0-9]([a-z0-9-]*[a-z0-9])?)$/
// No leading zeros, so "0443" can't dodge the default-port stripping.
const PORT = /^[1-9][0-9]{0,4}$/

const MAX_LENGTH = 253
const MAX_LABEL_LENGTH = 63
const MAX_PORT = 65535

/**
 * Brings a domain into the one form it is stored and compared in: trimmed,
 * lowercased, without a trailing slash or dot, and without :80 and :443 (a
 * browser leaves those out of the Host header, so "example.com:443" and
 * "example.com" are the same site). It never rejects anything -
 * validateShelfDomain judges the result.
 */
export function normalizeShelfDomain(raw: string): string {
  let value = raw.trim().toLowerCase()
  value = value.replace(/\/+$/, '')

  const colon = value.lastIndexOf(':')
  const hasPort = colon >= 0
  const port = hasPort ? value.slice(colon + 1) : ''
  const host = (hasPort ? value.slice(0, colon) : value).replace(/\.+$/, '')

  if (hasPort && port !== '80' && port !== '443') return `${host}:${port}`
  return host
}

/**
 * Checks an already normalized domain: a DNS name with a letters-only (or
 * punycode) top-level domain and an optional port. Returns a message saying
 * what is wrong, or null when it is fine. An empty domain is fine - a shelf
 * without one is simply reached through its path.
 */
export function validateShelfDomain(domain: string): string | null {
  if (domain === '') return null

  if (domain.includes('://') || domain.includes('/')) {
    return 'Enter just the domain, without https:// or a path'
  }

  const colon = domain.lastIndexOf(':')
  const host = colon >= 0 ? domain.slice(0, colon) : domain

  if (colon >= 0) {
    const port = domain.slice(colon + 1)
    if (!PORT.test(port) || Number(port) > MAX_PORT) {
      return 'The port must be a number between 1 and 65535'
    }
  }

  if (host.length > MAX_LENGTH) {
    return `A domain can have at most ${MAX_LENGTH} characters`
  }

  const labels = host.split('.')
  if (labels.length < 2) {
    return 'Please enter a full domain name (e.g. profile.example.com)'
  }

  for (const [index, label] of labels.entries()) {
    if (label.length > MAX_LABEL_LENGTH || !LABEL.test(label)) {
      return 'Please enter a valid domain (e.g. profile.example.com). Use letters, numbers and hyphens only'
    }
    if (index === labels.length - 1 && !TLD.test(label)) {
      return 'The domain must end in a valid top-level domain like .com'
    }
  }

  return null
}
