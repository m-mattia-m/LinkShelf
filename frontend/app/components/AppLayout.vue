<script setup lang="ts">
import type { FooterColumn, NavigationMenuItem } from '@nuxt/ui'
import type { SettingPageBody } from '~~/api'

const route = useRoute()
const router = useRouter()
const websiteSettings = useState<SettingPageBody | null>('settings')
const { t, locale, locales, setLocale } = useI18n()
const columns = computed<FooterColumn[]>(() => [
  {
    label: t('footer.general'),
    children: [
      {
        label: t('app.settings.about.label'),
        to: '/about',
        class: websiteSettings.value?.aboutShow ? '' : 'hidden'
      },
      {
        label: t('app.settings.contact.label'),
        to: '/contact',
        class: websiteSettings.value?.contactShow ? '' : 'hidden'
      }
    ]
  },
  {
    label: t('footer.legal'),
    children: [
      {
        label: t('app.settings.imprint.label'),
        to: '/imprint',
        class: websiteSettings.value?.imprintShow ? '' : 'hidden'
      },
      {
        label: t('app.settings.termsOfUse.label'),
        to: '/terms-of-use',
        class: websiteSettings.value?.termsOfUseShow ? '' : 'hidden'
      },
      {
        label: t('app.settings.privacyPolicy.label'),
        to: '/privacy-policy',
        class: websiteSettings.value?.privacyPolicyShow ? '' : 'hidden'
      }
    ]
  },
  {
    label: t('footer.community'),
    children: [
      {
        label: 'Github',
        to: 'https://github.com/m-mattia-m/LinkShelf',
        icon: 'uil-github',
        target: '_blank'
      },
      {
        label: 'Discord',
        to: 'https://discord.com/linkshelf',
        icon: 'uil-discord',
        target: '_blank'
      }
    ]
  }
])
// ULocaleSelect expects @nuxt/ui's Locale type, which has no de-CH pack, so
// only code/name are supplied and the type is cast.
const availableLocales = computed(() => {
  const mapped = locales.value.map(l => ({
    code: l.code,
    name: l.name ?? l.code
  }))
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return mapped as any
})
const items = computed<NavigationMenuItem[]>(() => [
  {
    label: t('nav.home'),
    to: '/',
    active: isActive('/')
  },
  {
    label: t('nav.docs'),
    to: '/docs',
    active: isActive('/docs')
  },
  {
    label: t('nav.cloud'),
    to: '/cloud',
    active: isActive('/cloud')
  },
  {
    label: t('app.dashboard.title'),
    to: '/app',
    active: isActive('/app')
  }
])

onMounted(() => {
  if (websiteSettings.value?.redirectToDashboard) router.push('/app')
})

const isActive = (base: string) =>
  route.path === base || route.path.startsWith(`${base}/`)
</script>

<template>
  <UApp>
    <UHeader>
      <template #left>
        <NuxtLink to="/">
          <AppLogo class="w-auto h-10 shrink-0" />
        </NuxtLink>
      </template>

      <UNavigationMenu
        color="neutral"
        :items="items"
        class="w-full"
      />

      <template #right>
        <UColorModeButton />

        <UButton
          to="https://github.com/m-mattia-m/LinkShelf"
          target="_blank"
          icon="i-simple-icons-github"
          aria-label="GitHub"
          color="neutral"
          variant="ghost"
        />
      </template>

      <template #body>
        <UNavigationMenu
          orientation="vertical"
          :items="items"
        />
      </template>
    </UHeader>

    <UMain>
      <slot />
    </UMain>

    <USeparator />

    <div class="p-4 sm:p-6 lg:p-8 text-dimmed">
      <UFooterColumns :columns="columns" />
      <ULocaleSelect
        class="mt-4 lg:mt-0"
        :model-value="locale"
        :locales="availableLocales"
        @update:model-value="setLocale($event as 'en' | 'de' | 'de-CH' | 'es')"
      />
      <p class="flex items-center justify-center mt-8">
        {{ t('footer.madeWith') }}
        <UIcon
          name="i-lucide-heart"
          class="mx-1"
        />
        {{ t('footer.byAll') }}
        <ULink
          href="https://github.com/m-mattia-m/LinkShelf/graphs/contributors"
          class="ml-1 text-dimmed"
        >
          {{ t('footer.contributors') }}
        </ULink>
      </p>
    </div>
  </UApp>
</template>

<style scoped>

</style>
