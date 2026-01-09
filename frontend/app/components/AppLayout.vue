<script setup lang="ts">
import type {FooterColumn, NavigationMenuItem} from "@nuxt/ui";
import type {SettingPageBody} from "~~/api";

const route = useRoute()
const router = useRouter()
const websiteSettings = useState('settings') as unknown as SettingPageBody
const columns: FooterColumn[] = [
  {
    label: 'General',
    children: [
      {
        label: 'About',
        to: '/about',
        class: websiteSettings.aboutShow ? 'hidden' : '',
      },
      {
        label: 'Contact',
        to: '/contact',
        class: websiteSettings.contactShow ? 'hidden' : '',
      }
    ]
  },
  {
    label: 'Legal',
    children: [
      {
        label: 'Imprint',
        to: '/imprint',
        class: websiteSettings.imprintShow ? 'hidden' : '',
      },
      {
        label: 'Terms of use',
        to: '/terms-of-use',
        class: websiteSettings.termsOfUseShow ? 'hidden' : '',
      },
      {
        label: 'Privacy policy',
        to: '/privacy-policy',
        class: websiteSettings.privacyPolicyShow ? 'hidden' : '',
      }
    ]
  },
  {
    label: 'Community',
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
]
const {locale, locales, setLocale} = useI18n()
const availableLocales = computed(() => {
  return locales.value.map(l => ({
    code: l.code,
    name: l.name ?? l.code
  }))
})
const items = computed<NavigationMenuItem[]>(() => [
  {
    label: 'Home',
    to: '/',
    active: isActive('/'),
  },
  {
    label: 'Docs',
    to: '/docs',
    active: isActive('/docs'),
  },
  {
    label: 'Cloud',
    to: '/cloud',
    active: isActive('/cloud'),
  },
  {
    label: 'Dashboard',
    to: '/app',
    active: isActive('/app'),
  }
])

onMounted(() => {
  if (websiteSettings.redirectToDashboard) router.push("/dashboard")
})

const isActive = (base: string) =>
  route.path === base || route.path.startsWith(`${base}/`)

</script>

<template>
  <UApp>
    <UHeader>
      <template #left>
        <NuxtLink to="/">
          <AppLogo class="w-auto h-6 shrink-0"/>
        </NuxtLink>
      </template>

      <UNavigationMenu color="neutral" :items="items" class="w-full"/>

      <template #right>
        <UColorModeButton/>

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
        <UNavigationMenu orientation="vertical" :items="items" />
      </template>

    </UHeader>

    <UMain>
      <slot/>
    </UMain>

    <USeparator/>

    <div class="p-4 sm:p-6 lg:p-8 text-dimmed">
      <UFooterColumns :columns="columns"/>
      <ULocaleSelect
        class="mt-4 lg:mt-0"
        :model-value="locale"
        :locales="availableLocales"
        @update:model-value="setLocale($event)"
      />
      <p class="flex items-center justify-center mt-8">
        Made with
        <UIcon name="i-lucide-heart" class="mx-1"/>
        by all
        <ULink href="https://github.com/m-mattia-m/LinkShelf/graphs/contributors" class="ml-1 text-dimmed">
          contributers
        </ULink>
      </p>
    </div>
  </UApp>
</template>

<style scoped>

</style>
