<script setup lang="ts">
import type { NavigationMenuItem, DropdownMenuItem } from '@nuxt/ui'

const route = useRoute()
const { t, locale, locales, setLocale } = useI18n()
const { user, ensureUser } = useCurrentUser()
const authStore = useAuthStore()
const router = useRouter()

// ULocaleSelect's `locales` prop is typed for @nuxt/ui's own Locale<M> (with
// `dir`/`messages` for its internal component strings), not @nuxtjs/i18n's
// app-content locale list this app actually configures - there's no de-CH
// @nuxt/ui locale pack to wire up here, so this intentionally only supplies
// code/name and casts past the mismatch. Kept in sync with the same switcher
// in AppLayout.vue's footer - both change the one global app locale.
const availableLocales = computed(() => {
  const mapped = locales.value.map(l => ({
    code: l.code,
    name: l.name ?? l.code
  }))
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return mapped as any
})

const items = computed<NavigationMenuItem[][]>(() => [
  [
    {
      label: 'Dashboard',
      to: '/app',
      exact: true,
      icon: 'uil-home-alt'
    },
    {
      label: 'Shelf',
      to: '/app/shelf',
      icon: 'uil-books',
      active: route.path.startsWith('/app/shelf')
    },
    {
      label: 'Themes',
      to: '/app/themes',
      icon: 'uil-palette',
      active: route.path.startsWith('/app/themes')
    },
    // Settings (general site config, user management) are admin-only - a
    // regular user can't reach these pages either (see middleware/admin.ts),
    // so there's no point showing the link.
    ...(authStore.isAdmin
      ? [{
          label: 'Settings',
          defaultOpen: true,
          icon: 'uil-cog',
          children: [
            {
              label: 'General',
              to: '/app/settings',
              exact: true
            },
            {
              label: 'Users',
              to: '/app/settings/users',
              exact: true
            },
            {
              label: 'Themes',
              to: '/app/settings/themes',
              exact: true
            }
          ]
        }]
      : [])
  ],
  [
    {
      label: 'Discord',
      to: 'https://discord.com/linkshelf',
      target: '_blank',
      icon: 'uil-discord'
    }
  ]
])

onMounted(() => {
  ensureUser()
})

const userLabel = computed(() => {
  if (!user.value) return ''
  const name = `${user.value.firstName} ${user.value.lastName}`.trim()
  return name || user.value.email
})

async function signOut() {
  await authStore.logout()
  await router.push('/auth/sign-in')
}

const userMenuItems = computed<DropdownMenuItem[][]>(() => [
  [
    {
      label: 'Account settings',
      icon: 'i-lucide-user-cog',
      to: '/app/profile'
    }
  ],
  [
    {
      label: 'Language',
      slot: 'language',
      onSelect: (e: Event) => e.preventDefault()
    },
    {
      label: 'Dark mode',
      slot: 'color-mode',
      onSelect: (e: Event) => e.preventDefault()
    }
  ],
  [
    {
      label: t('auth.userMenu.signOut'),
      icon: 'i-lucide-log-out',
      onSelect: signOut
    }
  ]
])
</script>

<template>
  <UDashboardGroup class="flex flex-col lg:flex-row">
    <UDashboardNavbar class="w-full lg:hidden">
      <UDashboardSidebarToggle />
    </UDashboardNavbar>

    <UDashboardSidebar
      collapsible
      resizable
      :ui="{ footer: 'border-t border-default' }"
    >
      <template #header="{ collapsed }">
        <ULink
          href="/"
          :class="collapsed ? 'mx-auto' : ''"
        >
          <AppLogo
            :class="collapsed ? 'size-8' : 'h-9 w-auto'"
            class="shrink-0"
          />
        </ULink>
      </template>

      <template #default="{ collapsed }">
        <UNavigationMenu
          :collapsed="collapsed"
          :items="items[0]"
          orientation="vertical"
        />

        <UNavigationMenu
          :collapsed="collapsed"
          :items="items[1]"
          orientation="vertical"
          class="mt-auto"
        />
      </template>

      <template #footer="{ collapsed }">
        <UDropdownMenu
          :items="userMenuItems"
          class="w-full"
        >
          <template #language>
            <div
              class="flex w-full items-center justify-between gap-2"
              @click.stop
            >
              <span class="flex items-center gap-2">
                <UIcon
                  name="i-lucide-languages"
                  class="size-4 shrink-0"
                />
                <span>Language</span>
              </span>
              <ULocaleSelect
                :model-value="locale"
                :locales="availableLocales"
                class="w-32"
                @update:model-value="setLocale($event as 'en' | 'de' | 'de-CH')"
              />
            </div>
          </template>

          <template #color-mode>
            <div
              class="flex w-full items-center justify-between gap-2"
              @click.stop
            >
              <span class="flex items-center gap-2">
                <UIcon
                  name="i-lucide-sun-moon"
                  class="size-4 shrink-0"
                />
                <span>Dark mode</span>
              </span>
              <UColorModeSwitch />
            </div>
          </template>

          <UButton
            :avatar="{ icon: 'i-lucide-user' }"
            :label="collapsed ? undefined : userLabel"
            color="neutral"
            variant="ghost"
            class="w-full"
            :block="collapsed"
          />
        </UDropdownMenu>
      </template>
    </UDashboardSidebar>

    <UDashboardPanel :ui="{ root: 'min-h-0 lg:min-h-svh', body: 'sm:py-8 sm:px-6 lg:px-8' }">
      <template #body>
        <slot />
      </template>
    </UDashboardPanel>
  </UDashboardGroup>
</template>

<style scoped>

</style>
