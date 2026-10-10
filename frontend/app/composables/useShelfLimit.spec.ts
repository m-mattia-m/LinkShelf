import { beforeEach, describe, expect, it } from 'vitest'
import type { SettingPageBody } from '~~/api'
import { buildShelf, buildTokenPair, buildUser } from '../../test/mocks/factories'
import { useShelfLimit } from './useShelfLimit'

function setUp(maxShelves: number | null, ownShelves: number, upgradeUrl = '') {
  const authStore = useAuthStore()
  authStore.setTokens(buildTokenPair())
  authStore.user = buildUser({ id: 'user-1', maxShelves })
  useShelfStore().shelves = [
    ...Array.from({ length: ownShelves }, () => buildShelf({ userId: 'user-1' })),
    buildShelf({ userId: 'someone-else' })
  ]
  useState('settings').value = { upgradeUrl } as unknown as SettingPageBody
}

beforeEach(() => {
  useAuthStore().$reset()
  useShelfStore().$reset()
  useState('settings').value = null
})

describe('useShelfLimit', () => {
  it('is never reached without a limit', () => {
    setUp(null, 50)
    const { maxShelves, reached } = useShelfLimit()

    expect(maxShelves.value).toBeNull()
    expect(reached.value).toBe(false)
  })

  it('counts only the caller\'s own shelves', () => {
    setUp(5, 2)

    expect(useShelfLimit().count.value).toBe(2)
  })

  it('is reached at the limit, and for a limit of zero', () => {
    setUp(2, 2)
    expect(useShelfLimit().reached.value).toBe(true)

    setUp(0, 0)
    expect(useShelfLimit().reached.value).toBe(true)
  })

  it('is not reached below the limit', () => {
    setUp(3, 2)

    expect(useShelfLimit().reached.value).toBe(false)
  })

  it('only exposes an http(s) upgrade url', () => {
    setUp(1, 1, 'https://account.example.com')
    expect(useShelfLimit().upgradeUrl.value).toBe('https://account.example.com')

    setUp(1, 1, 'javascript:alert(1)')
    expect(useShelfLimit().upgradeUrl.value).toBeUndefined()

    setUp(1, 1, '')
    expect(useShelfLimit().upgradeUrl.value).toBeUndefined()
  })
})
