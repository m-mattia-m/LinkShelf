// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: [
    '@nuxt/eslint',
    '@nuxt/ui',
    '@nuxt/hints',
    '@nuxt/image',
    '@nuxtjs/i18n',
    '@nuxt/icon',
    '@nuxtjs/mdc',
    '@nuxt/content',
    '@pinia/nuxt',
    '@nuxtjs/plausible',
    'nuxt-security'
  ],

  devtools: {
    enabled: true
  },

  css: ['~/assets/css/main.css'],

  content: {
    database: {
      type: 'sqlite',
      filename: ':memory:'
    }
  },

  ui: {
    theme: {
      colors: ['primary', 'secondary', 'tertiary', 'success', 'info', 'warning', 'error']
    }
  },

  runtimeConfig: {
    public: {
      apiBase: 'http://localhost:8085'
    }
  },

  // "/" is deliberately not prerendered (or cached per path): when the frontend
  // is reached on the domain of a shelf it shows that shelf, so what "/" renders
  // depends on the host and a static copy would serve one host's page to all.

  compatibilityDate: '2025-01-15',

  typescript: {
    tsConfig: {
      compilerOptions: {
        types: ['vitest/globals', '@testing-library/jest-dom']
      }
    }
  },

  eslint: {
    config: {
      stylistic: {
        commaDangle: 'never',
        braceStyle: '1tbs'
      }
    }
  },

  i18n: {
    defaultLocale: 'en',
    locales: [
      { code: 'en', name: 'English', file: 'en.json' },
      { code: 'de', name: 'Deutsch', file: 'de.json' },
      { code: 'de-CH', name: 'Schwiizerdütsch', file: 'de-CH.json' },
      { code: 'es', name: 'Español', file: 'es.json' }
    ],
    detectBrowserLanguage: false
  },

  // Disabled by default: instance owners opt in via NUXT_PUBLIC_PLAUSIBLE_ENABLED.
  // `domain` is left unset here so it falls back to window.location.hostname
  // (see @nuxtjs/plausible), which is always correct for a self-hosted deployment.
  plausible: {
    enabled: false,
    proxyBaseEndpoint: '/_plausible'
  },

  // Security headers. The part that matters most is the CSP's script-src: a
  // per-request nonce plus 'strict-dynamic' means only scripts Nuxt itself
  // rendered (and what they load) run. Browsers then ignore 'unsafe-inline'
  // and https:, so injected inline scripts and javascript: URLs are blocked -
  // this is what limits the damage of any XSS that slips past the backend
  // and safeHref. Everything else is loosened where the default would break
  // a feature of this app.
  security: {
    headers: {
      contentSecurityPolicy: {
        'base-uri': ['\'none\''],
        'object-src': ['\'none\''],
        'script-src-attr': ['\'none\''],
        // 'wasm-unsafe-eval': @nuxt/content queries its SQLite database in
        // the browser through WebAssembly.
        'script-src': ['\'self\'', 'https:', '\'unsafe-inline\'', '\'strict-dynamic\'', '\'nonce-{{nonce}}\'', '\'wasm-unsafe-eval\''],
        // Nuxt UI and themes set inline styles.
        'style-src': ['\'self\'', 'https:', '\'unsafe-inline\''],
        // Theme backgrounds, the backend's /images and QR codes (blob:) can
        // come from anywhere the instance admin configured, over http or https.
        'img-src': ['\'self\'', 'data:', 'blob:', 'https:', 'http:'],
        'font-src': ['\'self\'', 'https:', 'data:'],
        'form-action': ['\'self\''],
        'frame-ancestors': ['\'self\''],
        // Left out on purpose: connect-src (the API base URL and Plausible
        // host are only known at runtime) and upgrade-insecure-requests
        // (it would break instances whose API is served over plain http).
        'upgrade-insecure-requests': false
      },
      // require-corp/credentialless would block cross-origin theme images.
      crossOriginEmbedderPolicy: false,
      // The browser default: the API and outgoing links keep seeing the
      // origin, as before.
      referrerPolicy: 'strict-origin-when-cross-origin'
    },
    // The API is the Go backend; these would only police Nuxt's own few
    // server routes (e.g. the Plausible proxy) and could reject valid traffic.
    rateLimiter: false,
    requestSizeLimiter: false,
    xssValidator: false,
    // Keep console output in production builds, as before.
    removeLoggers: false
  }
})
