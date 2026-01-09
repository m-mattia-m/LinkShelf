<script setup lang="ts">
import type {NavigationMenuItem} from '@nuxt/ui'

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
      exact: true,
      icon: 'uil-books'
    },
    {
      label: 'Settings',
      defaultOpen: true,
      icon: 'uil-cog',
      children: [
        {
          label: 'General',
          to: '/app/settings',
          exact: true,
        },
        {
          label: 'Users',
          to: '/app/settings/users',
          exact: true,
        },
      ]
    }
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
        <AppLogo v-if="!collapsed" class="h-5 w-auto shrink-0" />
        <UIcon
          v-else
          name="i-simple-icons-nuxtdotjs"
          class="size-5 text-primary mx-auto"
        />
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
        <UButton
          :avatar="{ src: 'https://github.com/benjamincanac.png' }"
          :label="collapsed ? undefined : 'Benjamin'"
          color="neutral"
          variant="ghost"
          class="w-full"
          :block="collapsed"
        />
      </template>
    </UDashboardSidebar>

    <UDashboardPanel class="my-8 mx-4 sm:mx-6 lg:mx-8">
      <slot />
    </UDashboardPanel>
  </UDashboardGroup>
</template>



<style scoped>

</style>
