<script setup lang="ts">
import type { TableColumn } from '#ui/components/Table.vue'
import type { SettingPageBody, User } from '~~/api'
import { useUserStore } from '~/stores/user'
import UserFormDialog from '~/components/user/UserFormDialog.vue'
import UserPasswordDialog from '~/components/user/UserPasswordDialog.vue'

definePageMeta({
  layout: 'app',
  middleware: 'admin'
})

const { t } = useI18n()
const { userId: currentUserId } = useCurrentUser()
const websiteSettings = useState('settings') as unknown as Ref<SettingPageBody | null>

const userStore = useUserStore()
const loading = ref(true)
const actionLoading = ref<string | null>(null)

function userStatus(user: User): 'active' | 'pendingVerification' | 'invited' {
  if (user.emailVerified) return 'active'
  return user.hasPassword ? 'pendingVerification' : 'invited'
}

const statusBadge = {
  active: { label: t('app.settings.users.status.active'), color: 'success' as const },
  pendingVerification: { label: t('app.settings.users.status.pendingVerification'), color: 'warning' as const },
  invited: { label: t('app.settings.users.status.invited'), color: 'neutral' as const }
}

async function resendVerification(user: User) {
  actionLoading.value = user.id
  try {
    await userStore.resendVerification(user.email)
    const toast = useToast()
    toast.add({ title: t('app.settings.users.resendSuccess'), color: 'success' })
  } catch (err) {
    await handleApiError(err)
  } finally {
    actionLoading.value = null
  }
}

async function markVerified(user: User) {
  actionLoading.value = user.id
  try {
    await userStore.markVerified(user.id)
    const toast = useToast()
    toast.add({ title: t('app.settings.users.markVerifiedSuccess'), color: 'success' })
  } catch (err) {
    await handleApiError(err)
  } finally {
    actionLoading.value = null
  }
}

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

  if (websiteSettings.value?.emailVerificationEnabled && !user.emailVerified) {
    items.push([
      { label: t('app.settings.users.actions.resendVerification'), icon: 'i-lucide-mail', onSelect: () => resendVerification(user) },
      { label: t('app.settings.users.actions.markVerified'), icon: 'i-lucide-badge-check', onSelect: () => markVerified(user) }
    ])
  }

  // Only the account owner may change their own password - the backend
  // rejects this for anyone else, admins included.
  if (user.id === currentUserId.value) {
    items.push([{ label: t('app.settings.users.actions.changePassword'), icon: 'i-lucide-key-round', onSelect: () => openPassword(user) }])
  } else {
    items.push([{ label: t('app.settings.users.actions.delete'), icon: 'i-lucide-trash-2', color: 'error' as const, onSelect: () => openDelete(user) }])
  }

  return items
}

const columns = computed<TableColumn<User>[]>(() => {
  const cols: TableColumn<User>[] = [
    { accessorKey: 'firstName', header: t('app.settings.users.columns.firstName') },
    { accessorKey: 'lastName', header: t('app.settings.users.columns.lastName') },
    { accessorKey: 'email', header: t('app.settings.users.columns.email') }
  ]

  if (websiteSettings.value?.emailVerificationEnabled) {
    cols.push({ id: 'status', header: t('app.settings.users.columns.status') })
  }

  cols.push({ id: 'action' })
  return cols
})
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
    <template #status-cell="{ row }">
      <UBadge :label="statusBadge[userStatus(row.original)].label" :color="statusBadge[userStatus(row.original)].color" variant="subtle" />
    </template>

    <template #action-cell="{ row }">
      <UDropdownMenu :items="actionItems(row.original)">
        <UButton
          icon="i-lucide-ellipsis-vertical"
          color="neutral"
          variant="ghost"
          aria-label="Actions"
          :loading="actionLoading === row.original.id"
        />
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
