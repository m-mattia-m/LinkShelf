import { describe, expect, it } from 'vitest'
import { normalizeShelfDomain, validateShelfDomain } from './shelfDomain'

// The cases mirror Test_Unit_NormalizeDomain and Test_Unit_ValidateDomain in
// backend/internal/domain/shelf_domain_test.go: both sides have to agree.
describe('normalizeShelfDomain', () => {
  it.each([
    ['profile.example.com', 'profile.example.com'],
    ['  Profile.Example.COM  ', 'profile.example.com'],
    ['profile.example.com.', 'profile.example.com'],
    ['profile.example.com/', 'profile.example.com'],
    ['profile.example.com./', 'profile.example.com'],
    ['profile.example.com:443', 'profile.example.com'],
    ['profile.example.com:80', 'profile.example.com'],
    ['profile.example.com.:443/', 'profile.example.com'],
    ['profile.example.com:9443', 'profile.example.com:9443'],
    ['PROFILE.example.com:9443/', 'profile.example.com:9443'],
    ['profile.example.com:0443', 'profile.example.com:0443'],
    ['profile.example.com:', 'profile.example.com:'],
    ['https://profile.example.com', 'https://profile.example.com'],
    ['', ''],
    ['   ', '']
  ])('turns %j into %j', (input, expected) => {
    expect(normalizeShelfDomain(input)).toBe(expected)
  })

  it('is idempotent', () => {
    for (const input of ['A.b.COM.:443/', 'x.y.io:9443', 'junk', 'profile.example.com:0443']) {
      const once = normalizeShelfDomain(input)
      expect(normalizeShelfDomain(once)).toBe(once)
    }
  })
})

describe('validateShelfDomain', () => {
  it.each([
    'example.com',
    'profile.example.com',
    'a.b.c.d.example.co.uk',
    'profile.example.com:9443',
    'profile.example.com:1',
    'profile.example.com:65535',
    'my-profile.example.com',
    '123.example.com',
    'profile.example.xn--p1ai',
    'xn--bcher-kva.example.com',
    `${'a'.repeat(63)}.example.com`,
    ''
  ])('accepts %j', (domain) => {
    expect(validateShelfDomain(domain)).toBeNull()
  })

  it.each([
    ['localhost', 'single label'],
    ['example', 'single label'],
    ['1.2.3.4', 'IPv4 address'],
    ['192.168.0.1:8080', 'IPv4 address with a port'],
    ['*.example.com', 'wildcard'],
    ['https://profile.example.com', 'scheme'],
    ['profile.example.com/path', 'path'],
    ['profile.example.com:0', 'port zero'],
    ['profile.example.com:65536', 'port too big'],
    ['profile.example.com:0443', 'port with a leading zero'],
    ['profile.example.com:abc', 'port not a number'],
    ['profile.example.com:', 'empty port'],
    ['profile.example.com:1:2', 'two ports'],
    ['profile_x.example.com', 'underscore'],
    ['-profile.example.com', 'leading hyphen'],
    ['profile-.example.com', 'trailing hyphen'],
    ['profile..example.com', 'empty label'],
    ['.example.com', 'leading dot'],
    ['profile.example.c', 'one-letter TLD'],
    ['profile.example.c0m', 'TLD with a digit'],
    ['profile.example.123', 'numeric TLD'],
    ['profile.exämple.com', 'unicode label'],
    ['profile example.com', 'whitespace'],
    [`${'a'.repeat(64)}.example.com`, 'label over 63 characters'],
    [`${'a.'.repeat(130)}com`, 'name over 253 characters']
  ])('rejects %j (%s)', (domain) => {
    expect(validateShelfDomain(domain)).toEqual(expect.any(String))
  })

  it('says what is wrong in a way a person can act on', () => {
    expect(validateShelfDomain('https://a.example.com')).toContain('without https://')
    expect(validateShelfDomain('example')).toContain('full domain name')
    expect(validateShelfDomain('a.example.com:99999')).toContain('port')
    expect(validateShelfDomain('a.example.c0m')).toContain('top-level domain')
  })
})
