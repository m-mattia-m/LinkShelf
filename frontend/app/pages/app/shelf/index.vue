<script setup lang="ts">
import type { TableColumn } from '#ui/components/Table.vue'
import { useShelfStore } from '~/stores/shelf'
import type { SettingPageBody, Shelf } from '~~/api'
import ShelfFormDialog from '~/components/shelf/ShelfFormDialog.vue'

definePageMeta({
  app: 'Shelf',
  layout: 'app'
})

const { t } = useI18n()

const shelfStore = useShelfStore()
const authStore = useAuthStore()
const { maxShelves, count, reached, upgradeUrl } = useShelfLimit()
const loading = ref(true)

onMounted(async () => {
  try {
    await Promise.all([callOnce(shelfStore.fetch), authStore.fetchUser()])
  } catch (err) {
    await handleApiError(err)
  } finally {
    loading.value = false
  }
})

function getShelfUrl(id: string): string {
  return `/app/shelf/${id}`
}

// Where the shelf is served: its domain or its path.
const websiteSettings = useState('settings') as unknown as Ref<SettingPageBody | null>
const origin = useRequestURL().origin

function getPublicUrl(shelf: Shelf): string {
  return publicShelfUrl(shelf, { userBasedPaths: websiteSettings.value?.userBasedPaths ?? false, origin })
}

// A domain which still needs something from the user is flagged in the list, so it is not found by accident.
function domainBadge(shelf: Shelf) {
  const state = shelf.domainStatus?.state
  if (!shelf.domain || (state !== 'pending' && state !== 'error')) return null
  return state === 'error'
    ? { color: 'error' as const, icon: 'i-lucide-circle-alert', label: t('app.shelf.domainStatus.state.error') }
    : { color: 'warning' as const, icon: 'i-lucide-clock', label: t('app.shelf.domainAlert.pending.badge') }
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
    [{ label: t('app.shelf.actions.open'), icon: 'i-lucide-external-link', to: getPublicUrl(shelf), target: '_blank', disabled: !getPublicUrl(shelf) }],
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
  <div>
    <div class="flex justify-between items-center">
      <h1 class="text-2xl text-highlighted pb-4">
        {{ t('app.shelf.title') }}
      </h1>

      <ShelfFormDialog mode="create" />
    </div>

    <UAlert
      v-if="!loading && maxShelves !== null"
      :color="reached ? 'warning' : 'neutral'"
      variant="subtle"
      icon="i-lucide-gauge"
      class="mb-4"
      :title="t('app.shelf.limit.usage', { count, max: maxShelves })"
      :description="reached ? t('app.shelf.limit.reached') : undefined"
      :actions="reached && upgradeUrl ? [{ label: t('app.shelf.limit.upgrade'), to: upgradeUrl, target: '_blank', color: 'neutral', variant: 'outline' }] : undefined"
    />

    <div
      v-if="loading"
      class="space-y-2"
    >
      <USkeleton
        v-for="i in 3"
        :key="i"
        class="h-10 w-full"
      />
    </div>

    <div
      v-else-if="shelfStore.shelves.length === 0"
      class="flex flex-col items-center gap-4 py-16 text-center"
    >
      <p class="text-muted">
        {{ t('app.shelf.empty') }}
      </p>
      <ShelfFormDialog mode="create" />
    </div>

    <UTable
      v-else
      :columns="columns"
      :data="shelfStore.shelves"
      class="flex-1"
    >
      <template #title-cell="{ row }">
        <ULink
          :to="getShelfUrl(row.original.id)"
          class="font-medium"
        >
          {{ row.original.title }}
        </ULink>
      </template>

      <template #domain-cell="{ row }">
        <div
          v-if="row.original.domain"
          class="flex flex-wrap items-center gap-2"
        >
          <span>{{ row.original.domain }}</span>
          <UBadge
            v-if="domainBadge(row.original)"
            :color="domainBadge(row.original)!.color"
            :icon="domainBadge(row.original)!.icon"
            variant="subtle"
            size="sm"
            data-testid="domain-badge"
          >
            {{ domainBadge(row.original)!.label }}
          </UBadge>
        </div>
      </template>

      <template #action-cell="{ row }">
        <UDropdownMenu :items="actionItems(row.original)">
          <UButton
            icon="i-lucide-ellipsis-vertical"
            color="neutral"
            variant="ghost"
            :aria-label="t('common.actions')"
          />
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
  </div>
</template>
