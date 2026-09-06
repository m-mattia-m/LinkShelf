<script setup>
import {SettingApi} from "~~/api/index.ts";

const route = useRoute();
const {locale} = useI18n()
const settingApi = new SettingApi();

const websiteSettings = useState('settings')

await callOnce(async () => {
  try {
    websiteSettings.value = await settingApi.getPageSettings({languageCode: locale.value})
  } catch (error) {
    // The backend is not reachable while prerendering during the build, and may
    // be down at runtime; render without settings instead of failing the page.
    console.warn('[app] could not load page settings:', error)
    websiteSettings.value = null
  }
})

const title = 'LinkShelf'
const description = 'A production-ready starter template powered by Nuxt UI. Build beautiful, accessible, and performant applications in minutes, not hours.'

useHead({
  meta: [
    {name: 'viewport', content: 'width=device-width, initial-scale=1'},
    {name: 'theme-color', content: '#1c274c'}
  ],
  link: [
    {rel: 'icon', type: 'image/png', href: '/favicon-96x96.png', sizes: '96x96'},
    {rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg'},
    {rel: 'shortcut icon', href: '/favicon.ico'},
    {rel: 'apple-touch-icon', sizes: '180x180', href: '/apple-touch-icon.png'},
    {rel: 'manifest', href: '/site.webmanifest'}
  ],
  htmlAttrs: {
    lang: 'en'
  }
})

useSeoMeta({
  title,
  description,
  ogTitle: title,
  ogDescription: description,
  ogImage: '/web-app-manifest-512x512.png',
  twitterImage: '/web-app-manifest-512x512.png',
  twitterCard: 'summary_large_image'
})
</script>

<template>
  <ClientOnly>
    <NuxtLayout>
      <NuxtPage/>
    </NuxtLayout>
  </ClientOnly>
</template>
