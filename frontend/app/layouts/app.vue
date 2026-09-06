<script setup lang="ts">
import type { NavigationMenuItem, DropdownMenuItem } from '@nuxt/ui'

const route = useRoute()
const { t } = useI18n()
const { user, ensureUser } = useCurrentUser()
const authStore = useAuthStore()
const router = useRouter()

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
      label: t('auth.userMenu.signOut'),
      icon: 'i-lucide-log-out',
      onSelect: signOut
    }
  ]
])

</script>

<template>
  <UApp>
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
          <ULink href="/" :class="collapsed ? 'mx-auto' : ''">
            <AppLogo :class="collapsed ? 'size-8' : 'h-9 w-auto'" class="shrink-0" />
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
          <UDropdownMenu :items="userMenuItems" class="w-full">
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

      <UDashboardPanel :ui="{ body: 'sm:py-8 sm:px-6 lg:px-8' }">
        <template #body>
          <slot />
        </template>
      </UDashboardPanel>
    </UDashboardGroup>
  </UApp>
</template>



<style scoped>

</style>
