<script setup lang="ts">
import { useShelfStore } from '~/stores/shelf'

definePageMeta({
  layout: 'app'
})

const { t } = useI18n()

const shelfStore = useShelfStore()
const loading = ref(true)

onMounted(async () => {
  await callOnce(shelfStore.fetch)
  loading.value = false
})

const recentShelves = computed(() => shelfStore.shelves.slice(0, 5))

function getShelfUrl(id: string): string {
  return `/app/shelf/${id}`
}
</script>

<template>
  <h1 class="text-2xl text-highlighted pb-4">{{ t('app.dashboard.title') }}</h1>

  <div v-if="loading" class="grid gap-4 sm:grid-cols-2">
    <USkeleton class="h-24 w-full" />
    <USkeleton class="h-24 w-full" />
  </div>

  <div v-else class="flex flex-col gap-6">
    <UCard>
      <div class="flex items-center gap-4">
        <UIcon name="i-lucide-books" class="size-8 text-primary" />
        <div>
          <p class="text-2xl font-semibold">{{ shelfStore.shelves.length }}</p>
          <p class="text-muted text-sm">{{ t('app.dashboard.shelfCount') }}</p>
        </div>
      </div>
    </UCard>

    <div>
      <h2 class="text-lg font-medium pb-2">{{ t('app.dashboard.recentShelves') }}</h2>

      <p v-if="recentShelves.length === 0" class="text-muted">
        {{ t('app.dashboard.noShelves') }}
      </p>

      <ul v-else class="flex flex-col divide-y divide-default">
        <li v-for="shelf in recentShelves" :key="shelf.id" class="py-2">
          <ULink :to="getShelfUrl(shelf.id)" class="font-medium">{{ shelf.title }}</ULink>
          <p v-if="shelf.description" class="text-sm text-muted">{{ shelf.description }}</p>
        </li>
      </ul>
    </div>
  </div>
</template>
