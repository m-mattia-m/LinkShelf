<script setup lang="ts">
import type {TableColumn} from "#ui/components/Table.vue";
import {useShelfStore} from "~/stores/shelf";
import type {ShelfBase} from "~~/api";
import ShelfCreationDialog from "~/components/shelf/ShelfCreationDialog.vue";

const shelfStore = useShelfStore()

const data = ref<ShelfBase[]>([])
const columns: TableColumn<ShelfBase>[] = [
  {
    accessorKey: 'name',
    header: 'Name'
  },
  {
    accessorKey: 'description',
    header: 'Description'
  },
  {
    accessorKey: 'path',
    header: 'Path'
  },
  {
    accessorKey: 'domain',
    header: 'Domain'
  },
  {
    id: 'action',
  }
] satisfies TableColumn<ShelfBase>[]

function getShelfUrl(id: string): string {
  return `/app/shelf/${id}`
}

onMounted(async () => {
  await callOnce(shelfStore.fetch)
  console.log(shelfStore.shelves)
  data.value = shelfStore.shelves
})

definePageMeta({
  app: 'Shelf',
  layout: 'app',
})
</script>

<template>
  <div class="flex justify-between items-center">
    <h1 class="text-2xl text-highlighted pb-4">Shelf</h1>

    <ShelfCreationDialog/>
  </div>
    <UTable :columns="columns" :data="data" class="flex-1">
      <template #action-cell="{ row }">
        <ULink
          :href="getShelfUrl(row.original.id)"
          color="neutral"
          variant="ghost"
          aria-label="Actions"
        >
          <UIcon name="i-lucide-pencil" class="size-5"/>
        </ULink>
      </template>
    </UTable>
</template>

<style scoped>

</style>
