<script setup lang="ts">
import type { TableColumn } from '#ui/components/Table.vue'
import { useShelfStore } from '~/stores/shelf'
import type { Shelf } from '~~/api'
import ShelfFormDialog from '~/components/shelf/ShelfFormDialog.vue'

definePageMeta({
  app: 'Shelf',
  layout: 'app'
})

const { t } = useI18n()

const shelfStore = useShelfStore()
const loading = ref(true)

onMounted(async () => {
  await callOnce(shelfStore.fetch)
  loading.value = false
})

function getShelfUrl(id: string): string {
  return `/app/shelf/${id}`
}

const editOpen = ref(false)
const editingShelf = ref<Shelf>()

function openEdit(shelf: Shelf) {
  editingShelf.value = shelf
  editOpen.value = true
}

const deleteOpen = ref(false)
const deletingShelf = ref<Shelf | null>(null)
const deleting = ref(false)

function openDelete(shelf: Shelf) {
  deletingShelf.value = shelf
  deleteOpen.value = true
}

async function confirmDelete() {
  if (!deletingShelf.value) return
  deleting.value = true
  try {
    await shelfStore.remove(deletingShelf.value.id)
    deleteOpen.value = false
  } catch (err) {
    await handleApiError(err)
  } finally {
    deleting.value = false
  }
}

function actionItems(shelf: Shelf) {
  return [
    [{ label: t('app.shelf.actions.open'), icon: 'i-lucide-external-link', to: getShelfUrl(shelf.id) }],
    [{ label: t('app.shelf.actions.edit'), icon: 'i-lucide-pencil', onSelect: () => openEdit(shelf) }],
    [{ label: t('app.shelf.actions.delete'), icon: 'i-lucide-trash-2', color: 'error' as const, onSelect: () => openDelete(shelf) }]
  ]
}

const columns: TableColumn<Shelf>[] = [
  {
    accessorKey: 'title',
    header: t('app.shelf.columns.title')
  },
  {
    accessorKey: 'description',
    header: t('app.shelf.columns.description')
  },
  {
    accessorKey: 'path',
    header: t('app.shelf.columns.path')
  },
  {
    accessorKey: 'domain',
    header: t('app.shelf.columns.domain')
  },
  {
    id: 'action'
  }
] satisfies TableColumn<Shelf>[]
</script>

<template>
  <div class="flex justify-between items-center">
    <h1 class="text-2xl text-highlighted pb-4">{{ t('app.shelf.title') }}</h1>

    <ShelfFormDialog mode="create" />
  </div>

  <div v-if="loading" class="space-y-2">
    <USkeleton v-for="i in 3" :key="i" class="h-10 w-full" />
  </div>

  <div v-else-if="shelfStore.shelves.length === 0" class="flex flex-col items-center gap-4 py-16 text-center">
    <p class="text-muted">{{ t('app.shelf.empty') }}</p>
    <ShelfFormDialog mode="create" />
  </div>

  <UTable v-else :columns="columns" :data="shelfStore.shelves" class="flex-1">
    <template #title-cell="{ row }">
      <ULink :to="getShelfUrl(row.original.id)" class="font-medium">
        {{ row.original.title }}
      </ULink>
    </template>

    <template #action-cell="{ row }">
      <UDropdownMenu :items="actionItems(row.original)">
        <UButton icon="i-lucide-ellipsis-vertical" color="neutral" variant="ghost" aria-label="Actions" />
      </UDropdownMenu>
    </template>
  </UTable>

  <ShelfFormDialog
    v-model:open="editOpen"
    mode="edit"
    :shelf="editingShelf"
  />

  <ConfirmDialog
    v-model:open="deleteOpen"
    :title="t('app.shelf.deleteConfirm.title')"
    :description="t('app.shelf.deleteConfirm.description', { title: deletingShelf?.title })"
    :loading="deleting"
    @confirm="confirmDelete"
  />
</template>
