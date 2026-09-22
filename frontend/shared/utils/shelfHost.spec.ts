import { beforeEach, describe, expect, it, vi } from 'vitest'
import { clearShelfHostCache, requestHost, resolveShelfHost } from './shelfHost'

const API = 'https://api.example.com'

function fetchReturning(status: number) {
  return vi.fn(async () => new Response(status === 200 ? '{}' : null, { status })) as unknown as typeof fetch & ReturnType<typeof vi.fn>
}

beforeEach(() => {
  clearShelfHostCache()
})

describe('requestHost', () => {
  it('prefers X-Forwarded-Host over Host', () => {
    expect(requestHost({ 'host': 'internal:3000', 'x-forwarded-host': 'profile.example.com' })).toBe('profile.example.com')
  })

  it('takes the first of a list of forwarded hosts', () => {
    expect(requestHost({ 'x-forwarded-host': ' profile.example.com , proxy.internal' })).toBe('profile.example.com')
    expect(requestHost({ 'x-forwarded-host': ['profile.example.com, other.example.com'] })).toBe('profile.example.com')
  })

  it('falls back to Host, and to nothing', () => {
    expect(requestHost({ host: 'linkshelf.example.com' })).toBe('linkshelf.example.com')
    expect(requestHost({ 'x-forwarded-host': '', 'host': 'linkshelf.example.com' })).toBe('linkshelf.example.com')
    expect(requestHost({})).toBe('')
  })
})

describe('resolveShelfHost', () => {
  it('asks the API for the normalized domain and returns it when a shelf is served there', async () => {
    const fetchFn = fetchReturning(200)

    const result = await resolveShelfHost('  Profile.Example.COM:443 ', API, { fetchFn })

    expect(result).toBe('profile.example.com')
    expect(fetchFn).toHaveBeenCalledTimes(1)
    expect(vi.mocked(fetchFn).mock.calls[0]![0]).toBe(`${API}/v1/shelves/by-domain/profile.example.com`)
  })

  it('escapes a port in the path', async () => {
    const fetchFn = fetchReturning(200)

    expect(await resolveShelfHost('profile.example.com:9443', `${API}/`, { fetchFn })).toBe('profile.example.com:9443')
    expect(vi.mocked(fetchFn).mock.calls[0]![0]).toBe(`${API}/v1/shelves/by-domain/profile.example.com%3A9443`)
  })

  it('returns null for a host no shelf is served on', async () => {
    expect(await resolveShelfHost('linkshelf.example.com', API, { fetchFn: fetchReturning(404) })).toBeNull()
  })

  it.each(['', 'localhost', 'localhost:3000', '127.0.0.1:3000', '10.0.0.5', '*.example.com', 'https://a.example.com', 'a.example.com/x'])(
    'never asks the API about %j, which can\'t be a domain',
    async (host) => {
      const fetchFn = fetchReturning(200)

      expect(await resolveShelfHost(host, API, { fetchFn })).toBeNull()
      expect(fetchFn).not.toHaveBeenCalled()
    }
  )

  describe('cache', () => {
    it('remembers a hit and a miss for a minute', async () => {
      let now = 1_000_000
      const hit = fetchReturning(200)
      const miss = fetchReturning(404)

      expect(await resolveShelfHost('profile.example.com', API, { fetchFn: hit, now: () => now })).toBe('profile.example.com')
      expect(await resolveShelfHost('linkshelf.example.com', API, { fetchFn: miss, now: () => now })).toBeNull()

      now += 59_000
      expect(await resolveShelfHost('profile.example.com', API, { fetchFn: hit, now: () => now })).toBe('profile.example.com')
      expect(await resolveShelfHost('linkshelf.example.com', API, { fetchFn: miss, now: () => now })).toBeNull()
      expect(hit).toHaveBeenCalledTimes(1)
      expect(miss).toHaveBeenCalledTimes(1)

      now += 2_000
      await resolveShelfHost('profile.example.com', API, { fetchFn: hit, now: () => now })
      await resolveShelfHost('linkshelf.example.com', API, { fetchFn: miss, now: () => now })
      expect(hit).toHaveBeenCalledTimes(2)
      expect(miss).toHaveBeenCalledTimes(2)
    })

    it('treats every spelling of a host as the same entry', async () => {
      const fetchFn = fetchReturning(200)

      await resolveShelfHost('Profile.example.com', API, { fetchFn })
      await resolveShelfHost('profile.example.com.:443/', API, { fetchFn })

      expect(fetchFn).toHaveBeenCalledTimes(1)
    })

    it('does not remember an unreachable backend or an error, so the main site recovers', async () => {
      const failing = vi.fn(async () => {
        throw new Error('connect ECONNREFUSED')
      }) as unknown as typeof fetch
      const broken = fetchReturning(500)
      const working = fetchReturning(200)

      expect(await resolveShelfHost('profile.example.com', API, { fetchFn: failing })).toBeNull()
      expect(await resolveShelfHost('profile.example.com', API, { fetchFn: broken })).toBeNull()
      expect(await resolveShelfHost('profile.example.com', API, { fetchFn: working })).toBe('profile.example.com')
    })

    it('stays bounded, so a stream of made-up hosts cannot fill the memory', async () => {
      let now = 1_000_000
      const fetchFn = fetchReturning(404)

      for (let i = 0; i < 1500; i++) {
        await resolveShelfHost(`h${i}.example.com`, API, { fetchFn, now: () => now })
      }
      // The oldest entries were dropped to make room, the newest are still remembered.
      const before = vi.mocked(fetchFn).mock.calls.length
      await resolveShelfHost('h1499.example.com', API, { fetchFn, now: () => now })
      expect(vi.mocked(fetchFn).mock.calls.length).toBe(before)

      // And an expired cache is swept rather than reset.
      now += 120_000
      await resolveShelfHost('fresh.example.com', API, { fetchFn, now: () => now })
      await resolveShelfHost('fresh.example.com', API, { fetchFn, now: () => now })
      expect(vi.mocked(fetchFn).mock.calls.length).toBe(before + 1)
    })
  })
})
