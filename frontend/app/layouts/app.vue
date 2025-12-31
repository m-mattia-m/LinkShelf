<script setup lang="ts">
import type {NavigationMenuItem} from '@nuxt/ui'

const items: NavigationMenuItem[][] = [
  [
    {
      label: 'Dashbaord',
      icon: 'i-lucide-house',
      active: true
    },
    {
      label: 'Shelfs',
      icon: 'i-lucide-square-library'
    },
    {
      label: 'Settings',
      icon: 'i-lucide-settings',
      defaultOpen: true,
      children: [
        {label: 'General'},
        {label: 'Users'},
      ]
    }
  ],
  [
    {
      label: 'Discord',
      icon: 'uil-discord',
      to: 'https://discord.com/linkshelf',
      target: '_blank'
    }
  ]
]
</script>

<template>
  <UDashboardSidebar collapsible resizable :ui="{ footer: 'border-t border-default' }">
    <template #header="{ collapsed }">
      <Logo v-if="!collapsed" class="h-5 w-auto shrink-0"/>
      <UIcon v-else name="i-simple-icons-nuxtdotjs" class="size-5 text-primary mx-auto"/>
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
        :avatar="{
          src: 'https://github.com/benjamincanac.png'
        }"
        :label="collapsed ? undefined : 'Benjamin'"
        color="neutral"
        variant="ghost"
        class="w-full"
        :block="collapsed"
      />
    </template>
  </UDashboardSidebar>
</template>


<style scoped>

</style>
