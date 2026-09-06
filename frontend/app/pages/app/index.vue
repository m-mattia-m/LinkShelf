<script setup lang="ts">
import type { Statistic } from '~~/api'
import { useShelfStore } from '~/stores/shelf'

definePageMeta({
  layout: 'app'
})

const { t } = useI18n()

const shelfStore = useShelfStore()
const statistic = ref<Statistic>()
const loading = ref(true)

onMounted(async () => {
  const api = useApi()
  await Promise.all([
    callOnce(shelfStore.fetch),
    api.statistic.getStatistic().then((result) => { statistic.value = result })
  ])
  loading.value = false
})

const tiles = computed(() => [
  { icon: 'i-lucide-books', label: t('app.dashboard.shelfCount'), value: statistic.value?.shelfNumber ?? 0 },
  { icon: 'i-lucide-list', label: t('app.dashboard.sectionCount'), value: statistic.value?.sectionNumber ?? 0 },
  { icon: 'i-lucide-link', label: t('app.dashboard.linkCount'), value: statistic.value?.linkNumber ?? 0 }
])

const recentShelves = computed(() => shelfStore.shelves.slice(0, 5))

function getShelfUrl(id: string): string {
  return `/app/shelf/${id}`
}
</script>

<template>
  <h1 class="text-2xl text-highlighted pb-4">{{ t('app.dashboard.title') }}</h1>

  <div v-if="loading" class="grid gap-4 sm:grid-cols-3">
    <USkeleton class="h-24 w-full" />
    <USkeleton class="h-24 w-full" />
    <USkeleton class="h-24 w-full" />
  </div>

  <div v-else class="flex flex-col gap-6">
    <div class="grid gap-4 sm:grid-cols-3">
      <UCard v-for="tile in tiles" :key="tile.label">
        <div class="flex items-center gap-4">
          <UIcon :name="tile.icon" class="size-8 text-primary" />
          <div>
            <p class="text-2xl font-semibold">{{ tile.value }}</p>
            <p class="text-muted text-sm">{{ tile.label }}</p>
          </div>
        </div>
      </UCard>
    </div>

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
