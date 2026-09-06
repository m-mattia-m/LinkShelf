<script setup lang="ts">
import type { TableColumn } from '#ui/components/Table.vue'
import type { User } from '~~/api'
import { useUserStore } from '~/stores/user'
import UserFormDialog from '~/components/user/UserFormDialog.vue'
import UserPasswordDialog from '~/components/user/UserPasswordDialog.vue'

definePageMeta({
  layout: 'app',
  middleware: 'admin'
})

const { t } = useI18n()
const { userId: currentUserId } = useCurrentUser()

const userStore = useUserStore()
const loading = ref(true)

onMounted(async () => {
  try {
    await callOnce(userStore.fetch)
  } catch (err) {
    await handleApiError(err)
  } finally {
    loading.value = false
  }
})

const editOpen = ref(false)
const editingUser = ref<User>()

function openEdit(user: User) {
  editingUser.value = user
  editOpen.value = true
}

const passwordOpen = ref(false)
const passwordTarget = ref<User>()

function openPassword(user: User) {
  passwordTarget.value = user
  passwordOpen.value = true
}

const deleteOpen = ref(false)
const deletingUser = ref<User | null>(null)
const deleting = ref(false)

function openDelete(user: User) {
  deletingUser.value = user
  deleteOpen.value = true
}

async function confirmDelete() {
  if (!deletingUser.value) return
  deleting.value = true
  try {
    await userStore.remove(deletingUser.value.id)
    deleteOpen.value = false
  } catch (err) {
    await handleApiError(err)
  } finally {
    deleting.value = false
  }
}

function actionItems(user: User) {
  const items = [
    [{ label: t('app.settings.users.actions.edit'), icon: 'i-lucide-pencil', onSelect: () => openEdit(user) }]
  ]

  // Only the account owner may change their own password - the backend
  // rejects this for anyone else, admins included.
  if (user.id === currentUserId.value) {
    items.push([{ label: t('app.settings.users.actions.changePassword'), icon: 'i-lucide-key-round', onSelect: () => openPassword(user) }])
  } else {
    items.push([{ label: t('app.settings.users.actions.delete'), icon: 'i-lucide-trash-2', color: 'error' as const, onSelect: () => openDelete(user) }])
  }

  return items
}

const columns: TableColumn<User>[] = [
  { accessorKey: 'firstName', header: t('app.settings.users.columns.firstName') },
  { accessorKey: 'lastName', header: t('app.settings.users.columns.lastName') },
  { accessorKey: 'email', header: t('app.settings.users.columns.email') },
  { id: 'action' }
]
</script>

<template>
  <div class="flex justify-between items-center pb-4">
    <h1 class="text-2xl text-highlighted flex items-center gap-1">
      {{ t('app.settings.title') }}
      <UIcon name="i-lucide-chevron-right" class="size-5" />
      {{ t('app.settings.users.title') }}
    </h1>

    <UserFormDialog mode="create" />
  </div>

  <div v-if="loading" class="space-y-2">
    <USkeleton v-for="i in 3" :key="i" class="h-10 w-full" />
  </div>

  <div v-else-if="userStore.users.length === 0" class="flex flex-col items-center gap-4 py-16 text-center">
    <p class="text-muted">{{ t('app.settings.users.empty') }}</p>
    <UserFormDialog mode="create" />
  </div>

  <UTable v-else :columns="columns" :data="userStore.users" class="flex-1">
    <template #action-cell="{ row }">
      <UDropdownMenu :items="actionItems(row.original)">
        <UButton icon="i-lucide-ellipsis-vertical" color="neutral" variant="ghost" aria-label="Actions" />
      </UDropdownMenu>
    </template>
  </UTable>

  <UserFormDialog
    v-model:open="editOpen"
    mode="edit"
    :user="editingUser"
  />

  <UserPasswordDialog
    v-model:open="passwordOpen"
    :user="passwordTarget"
  />

  <ConfirmDialog
    v-model:open="deleteOpen"
    :title="t('app.settings.users.deleteConfirm.title')"
    :description="t('app.settings.users.deleteConfirm.description', { email: deletingUser?.email })"
    :loading="deleting"
    @confirm="confirmDelete"
  />
</template>
