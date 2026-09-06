import type { TokenPair, User, UserCreate } from '~~/api'

const ACCESS_TOKEN_KEY = 'linkshelf.accessToken'
const REFRESH_TOKEN_KEY = 'linkshelf.refreshToken'
const OIDC_STATE_KEY = 'linkshelf.oidcState'

interface AccessTokenClaims {
  sub: string
  role: string
  exp: number
}

function decodeAccessToken(token: string): AccessTokenClaims | null {
  try {
    const payload = token.split('.')[1]
    if (!payload) return null
    const base64 = payload.replace(/-/g, '+').replace(/_/g, '/')
    return JSON.parse(atob(base64)) as AccessTokenClaims
  } catch {
    return null
  }
}

export const useAuthStore = defineStore('authStore', {
  state: () => ({
    accessToken: null as string | null,
    refreshToken: null as string | null,
    user: null as User | null,
    initialized: false
  }),

  getters: {
    isAuthenticated: (state) => !!state.accessToken,
    claims: (state) => (state.accessToken ? decodeAccessToken(state.accessToken) : null),
    userId(): string | null {
      return this.claims?.sub ?? null
    },
    role(): string | null {
      return this.claims?.role ?? null
    },
    isAdmin(): boolean {
      return this.role === 'admin'
    }
  },

  actions: {
    /**
     * Restores tokens saved by a previous session. Client-only: tokens live
     * in localStorage, never in a cookie, so there's nothing to restore
     * during SSR.
     */
    init() {
      if (this.initialized || !import.meta.client) return
      this.initialized = true
      this.accessToken = localStorage.getItem(ACCESS_TOKEN_KEY)
      this.refreshToken = localStorage.getItem(REFRESH_TOKEN_KEY)
    },

    setTokens(tokens: TokenPair) {
      this.accessToken = tokens.accessToken
      this.refreshToken = tokens.refreshToken
      if (import.meta.client) {
        localStorage.setItem(ACCESS_TOKEN_KEY, tokens.accessToken)
        localStorage.setItem(REFRESH_TOKEN_KEY, tokens.refreshToken)
      }
    },

    clear() {
      this.accessToken = null
      this.refreshToken = null
      this.user = null
      if (import.meta.client) {
        localStorage.removeItem(ACCESS_TOKEN_KEY)
        localStorage.removeItem(REFRESH_TOKEN_KEY)
      }
    },

    async login(email: string, password: string): Promise<void> {
      const api = useApi()
      const tokens = await api.auth.postLogin({ loginRequest: { email, password } })
      this.setTokens(tokens)
      await this.fetchUser()
    },

    async register(userCreate: UserCreate): Promise<void> {
      const api = useApi()
      await api.user.postCreateUser({ userCreate })
      await this.login(userCreate.email, userCreate.password)
    },

    async fetchUser(): Promise<void> {
      if (!this.accessToken) return
      const api = useApi()
      this.user = await api.user.getCurrentUser()
    },

    /**
     * Exchanges the refresh token for a new pair (single-use rotation on the
     * backend - the old refresh token stops working the moment this call
     * succeeds). Clears the session on failure.
     */
    async refresh(): Promise<boolean> {
      if (!this.refreshToken) return false
      try {
        const api = useApi()
        const tokens = await api.auth.postRefresh({ refreshRequest: { refreshToken: this.refreshToken } })
        this.setTokens(tokens)
        return true
      } catch {
        this.clear()
        return false
      }
    },

    async logout(): Promise<void> {
      if (this.refreshToken) {
        try {
          const api = useApi()
          await api.auth.postLogout({ refreshRequest: { refreshToken: this.refreshToken } })
        } catch {
          // Best effort - still clear the local session below.
        }
      }
      this.clear()
    },

    /**
     * Starts an OIDC login by redirecting the browser to the provider. The
     * state is stashed so the callback page can pass it back untouched.
     */
    async startOidcLogin(): Promise<void> {
      const api = useApi()
      const { authorizationUrl, state } = await api.auth.getOidcLogin()
      if (import.meta.client) {
        sessionStorage.setItem(OIDC_STATE_KEY, state)
      }
      window.location.href = authorizationUrl
    },

    /**
     * Completes an OIDC login/link. When the caller is already authenticated,
     * this links the external identity to the current account instead of
     * logging in as a different one - matching the backend's callback
     * behavior, which branches on whether a Bearer token was sent.
     */
    async completeOidcLogin(code: string, state: string): Promise<void> {
      const api = useApi()
      const tokens = await api.auth.postOidcCallback({
        oidcCallbackRequest: { code, state },
        authorization: this.accessToken ? `Bearer ${this.accessToken}` : undefined
      })
      this.setTokens(tokens)
      await this.fetchUser()
    },

    consumeStoredOidcState(): string | null {
      if (!import.meta.client) return null
      const state = sessionStorage.getItem(OIDC_STATE_KEY)
      sessionStorage.removeItem(OIDC_STATE_KEY)
      return state
    }
  }
})
