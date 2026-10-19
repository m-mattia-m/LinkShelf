<script setup lang="ts">
const { locale, t } = useI18n()

const websiteSettings = useState('settings')

await callOnce(async () => {
  try {
    websiteSettings.value = await useApi().setting.getPageSettings({ languageCode: locale.value })
  } catch (error) {
    // The backend may be unreachable (prerender or downtime); render without settings.
    console.warn('[app] could not load page settings:', error)
    websiteSettings.value = null
  }
})

const route = useRoute()
const { origin, isMainHost, absolute } = useSiteSeo()
const shelfHost = useState<string | null | undefined>('shelf-host')

const title = computed(() => t('seo.title'))
const description = computed(() => t('seo.description'))
// public/og-image.png is 1200x630; without a site URL there is no absolute URL, so use the icon.
const ogImage = computed(() => origin ? absolute('/og-image.png') : '/web-app-manifest-512x512.png')

useHead(() => ({
  meta: [
    { name: 'viewport', content: 'width=device-width, initial-scale=1' },
    { name: 'theme-color', content: '#1D3557' }
  ],
  link: [
    { rel: 'icon', type: 'image/png', href: '/favicon-96x96.png', sizes: '96x96' },
    { rel: 'icon', type: 'image/svg+xml', href: '/logo.svg' },
    { rel: 'shortcut icon', href: '/favicon.ico' },
    { rel: 'apple-touch-icon', sizes: '180x180', href: '/apple-touch-icon.png' },
    { rel: 'manifest', href: '/site.webmanifest' },
    ...(isMainHost ? [{ rel: 'canonical' as const, href: absolute(route.path) }] : [])
  ],
  htmlAttrs: {
    lang: locale.value
  }
}))

useSeoMeta({
  title,
  description,
  ogTitle: title,
  ogDescription: description,
  ogImage,
  ogImageWidth: () => origin ? 1200 : undefined,
  ogImageHeight: () => origin ? 630 : undefined,
  ogType: 'website',
  ogSiteName: 'LinkShelf',
  ogUrl: () => isMainHost ? absolute(route.path) : undefined,
  twitterCard: 'summary_large_image',
  twitterTitle: title,
  twitterDescription: description,
  twitterImage: ogImage
})

if (isMainHost && stripLocalePrefix(route.path) === '/') {
  useHead({
    script: [{
      type: 'application/ld+json',
      innerHTML: JSON.stringify({
        '@context': 'https://schema.org',
        '@type': 'SoftwareApplication',
        'name': 'LinkShelf',
        'applicationCategory': 'WebApplication',
        'operatingSystem': 'Linux, Docker, Kubernetes',
        'license': 'https://www.gnu.org/licenses/agpl-3.0.html',
        'description': description.value,
        'url': absolute('/'),
        'codeRepository': 'https://github.com/m-mattia-m/LinkShelf'
      })
    }]
  })
}

// Only the instance's own marketing pages are rendered on the server. A shelf on this host, or any
// other host, stays client-only: its content depends on the host and the API.
const renderOnServer = computed(() => isMainHost && !shelfHost.value && isMarketingPath(route.path))
</script>

<template>
  <UApp v-if="renderOnServer">
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
  </UApp>
  <ClientOnly v-else>
    <UApp>
      <NuxtLayout>
        <NuxtPage />
      </NuxtLayout>
    </UApp>
  </ClientOnly>
</template>
