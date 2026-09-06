import type { User } from '~~/api'

// The UUID of the dev user seeded by backend/migrations/0003_dashboard.up.sql.
// Real authentication isn't built yet - swap this out once it exists and every
// caller of useCurrentUser() keeps working unchanged.
const MOCK_USER_ID = '018f1a3e-0000-7000-8000-000000000001'

async function fetchCurrentUserId(): Promise<string> {
  // Mocks an authentication lookup.
  return MOCK_USER_ID
}

export function useCurrentUser() {
  const userId = useState<string | null>('current-user-id', () => null)
  const user = useState<User | null>('current-user', () => null)

  async function ensureUserId(): Promise<string> {
    if (!userId.value) {
      userId.value = await fetchCurrentUserId()
    }
    return userId.value
  }

  async function ensureUser(): Promise<User> {
    await ensureUserId()
    if (!user.value) {
      const api = useApi()
      user.value = await api.user.getUserById({ userId: userId.value! })
    }
    return user.value
  }

  return { userId, user, ensureUserId, ensureUser }
}
