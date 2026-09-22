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
    '@pinia/nuxt'
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
      { code: 'de-CH', name: 'Schwiizerdütsch', file: 'de-CH.json' }
    ]
  }
})
