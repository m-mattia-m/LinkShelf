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

  // "/" depends on the host (shelf domains), so it's not prerendered.

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

  // Opt in via NUXT_PUBLIC_PLAUSIBLE_ENABLED; domain defaults to the current hostname.
  plausible: {
    enabled: false,
    proxyBaseEndpoint: '/_plausible'
  },

  // CSP: a nonce plus 'strict-dynamic' only allows scripts rendered by Nuxt.
  security: {
    headers: {
      contentSecurityPolicy: {
        'base-uri': ['\'none\''],
        'object-src': ['\'none\''],
        'script-src-attr': ['\'none\''],
        // @nuxt/content runs SQLite via WebAssembly.
        'script-src': ['\'self\'', 'https:', '\'unsafe-inline\'', '\'strict-dynamic\'', '\'nonce-{{nonce}}\'', '\'wasm-unsafe-eval\''],
        // Nuxt UI and themes set inline styles.
        'style-src': ['\'self\'', 'https:', '\'unsafe-inline\''],
        // Images can come from any configured host, or blob: for QR codes.
        'img-src': ['\'self\'', 'data:', 'blob:', 'https:', 'http:'],
        'font-src': ['\'self\'', 'https:', 'data:'],
        'form-action': ['\'self\''],
        'frame-ancestors': ['\'self\''],
        // connect-src and upgrade-insecure-requests are omitted: the API URL is only known at runtime.
        'upgrade-insecure-requests': false
      },
      // require-corp/credentialless would block cross-origin theme images.
      crossOriginEmbedderPolicy: false,
      // The browser default.
      referrerPolicy: 'strict-origin-when-cross-origin'
    },
    // Only Nuxt's own server routes would be affected.
    rateLimiter: false,
    requestSizeLimiter: false,
    xssValidator: false,
    // Keep console output in production builds, as before.
    removeLoggers: false
  }
})
