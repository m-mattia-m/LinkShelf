// Mirrors backend/internal/domain/shelf_domain.go; the backend has the final say.

const LABEL = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/
// Letters only, so an IP address can never pass, or a punycode international TLD.
const TLD = /^([a-z]{2,}|xn--[a-z0-9]([a-z0-9-]*[a-z0-9])?)$/
// No leading zeros, so "0443" can't dodge the default-port stripping.
const PORT = /^[1-9][0-9]{0,4}$/

const MAX_LENGTH = 253
const MAX_LABEL_LENGTH = 63
const MAX_PORT = 65535

/** Lowercases a domain and strips trailing slash/dot and default ports. */
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

/** Validates a normalized domain. Returns an i18n error key or null. */
export function validateShelfDomain(domain: string): string | null {
  if (domain === '') return null

  if (domain.includes('://') || domain.includes('/')) {
    return 'validation.domain.withScheme'
  }

  const colon = domain.lastIndexOf(':')
  const host = colon >= 0 ? domain.slice(0, colon) : domain

  if (colon >= 0) {
    const port = domain.slice(colon + 1)
    if (!PORT.test(port) || Number(port) > MAX_PORT) {
      return 'validation.domain.port'
    }
  }

  if (host.length > MAX_LENGTH) {
    return 'validation.domain.tooLong'
  }

  const labels = host.split('.')
  if (labels.length < 2) {
    return 'validation.domain.notFull'
  }

  for (const [index, label] of labels.entries()) {
    if (label.length > MAX_LABEL_LENGTH || !LABEL.test(label)) {
      return 'validation.domain.invalidLabel'
    }
    if (index === labels.length - 1 && !TLD.test(label)) {
      return 'validation.domain.tld'
    }
  }

  return null
}
