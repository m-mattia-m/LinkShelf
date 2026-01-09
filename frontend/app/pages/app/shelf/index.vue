<script setup lang="ts">
import type {TableColumn} from "#ui/components/Table.vue";

interface Shelf {
  id: string,
  name: string,
  description: string,
  path: string,
  domain: string,
}

const data = ref<Shelf[]>([
  {
    id: '019ba50b-c797-70f1-85ed-33b556a1901c',
    name: 'My links',
    description: 'A collection of all my links I want to share.',
    path: '/my-links',
    domain: 'mattias-links.ch',
  },
])

const columns: TableColumn<Shelf>[] = [
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
] satisfies TableColumn<Shelf>[]

function getShelfUrl(id: string): string {
  return `/app/shelf/${id}`
}

definePageMeta({
  app: 'Shelf',
  layout: 'app',
})
</script>

<template>
  <h1 class="text-2xl text-highlighted pb-4">Shelf</h1>
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
