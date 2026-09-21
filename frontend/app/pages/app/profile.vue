<script setup lang="ts">
import * as v from 'valibot'
import type { FormSubmitEvent } from '@nuxt/ui'
import UserPasswordDialog from '~/components/user/UserPasswordDialog.vue'
import ConfirmDialog from '~/components/ConfirmDialog.vue'
import type { SettingPageBody } from '~~/api'
import { useShelfStore } from '~/stores/shelf'

definePageMeta({
  layout: 'app'
})

const { t } = useI18n()
const { user, ensureUser } = useCurrentUser()
const authStore = useAuthStore()
const shelfStore = useShelfStore()
const websiteSettings = useState('settings') as unknown as Ref<SettingPageBody | null>
const userBasedPaths = computed(() => websiteSettings.value?.userBasedPaths ?? false)

const loading = ref(true)
const saving = ref(false)
const passwordOpen = ref(false)

const form = reactive({
  firstName: '',
  lastName: '',
  username: '',
  email: ''
})

// Renaming changes the URL of every shelf that has a path, so it is
// confirmed first - but only when user-based paths are on and there is
// something to break.
const confirmOpen = ref(false)
const pendingProfile = ref<Schema | null>(null)
const affectedShelves = computed(() =>
  shelfStore.shelves.filter(shelf => shelf.userId === user.value?.id && shelf.path).length
)

onMounted(async () => {
  try {
    await ensureUser()
    if (user.value) {
      form.firstName = user.value.firstName
      form.lastName = user.value.lastName
      form.username = user.value.username
      form.email = user.value.email
    }
    if (userBasedPaths.value && !shelfStore.loaded) await shelfStore.fetch()
  } catch (err) {
    await handleApiError(err)
  } finally {
    loading.value = false
  }
})

const schema = v.object({
  firstName: v.pipe(v.string('First name is required'), v.nonEmpty('First name is required')),
  lastName: v.pipe(v.string('Last name is required'), v.nonEmpty('Last name is required')),
  username: usernameSchema(),
  email: v.pipe(
    v.string('Email is required'),
    v.nonEmpty('Email is required'),
    v.email('Please enter a valid email')
  )
})

type Schema = v.InferOutput<typeof schema>

function onSubmit(payload: FormSubmitEvent<Schema>) {
  if (!user.value) return

  const renamed = payload.data.username !== user.value.username
  if (renamed && userBasedPaths.value && affectedShelves.value > 0) {
    pendingProfile.value = payload.data
    confirmOpen.value = true
    return
  }
  return save(payload.data)
}

async function confirmRename() {
  const data = pendingProfile.value
  confirmOpen.value = false
  pendingProfile.value = null
  if (data) await save(data)
}

async function save(data: Schema) {
  if (!user.value) return
  saving.value = true
  try {
    const api = useApi()
    const updated = await api.user.putUpdateUser({
      userId: user.value.id,
      userBase: {
        firstName: data.firstName,
        lastName: data.lastName,
        username: data.username,
        email: data.email
      }
    })
    authStore.user = updated
    // The shelves' URLs follow the username, so their cached copies are stale.
    if (userBasedPaths.value) await shelfStore.fetch()
    const toast = useToast()
    toast.add({ title: t('app.profile.saveSuccess'), color: 'success' })
  } catch (err) {
    await handleApiError(err)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <h1 class="text-2xl text-highlighted pb-4">
      {{ t('app.profile.title') }}
    </h1>

    <div
      v-if="loading"
      class="flex flex-col gap-4 max-w-md"
    >
      <USkeleton
        v-for="i in 3"
        :key="i"
        class="h-10 w-full"
      />
    </div>

    <UForm
      v-else
      :schema="schema"
      :state="form"
      class="flex flex-col gap-4 max-w-md"
      @submit="onSubmit"
    >
      <UFormField
        :label="t('app.profile.firstName')"
        name="firstName"
        required
      >
        <UInput
          v-model="form.firstName"
          class="w-full"
        />
      </UFormField>

      <UFormField
        :label="t('app.profile.lastName')"
        name="lastName"
        required
      >
        <UInput
          v-model="form.lastName"
          class="w-full"
        />
      </UFormField>

      <UFormField
        :label="t('app.profile.username')"
        name="username"
        required
        :help="userBasedPaths ? t('app.profile.usernameHelpUserBased', { username: form.username || '…' }) : t('app.profile.usernameHelp')"
      >
        <UInput
          v-model="form.username"
          class="w-full"
        />
      </UFormField>

      <UFormField
        :label="t('app.profile.email')"
        name="email"
        required
      >
        <UInput
          v-model="form.email"
          type="email"
          class="w-full"
        />
      </UFormField>

      <UFormField :label="t('app.profile.role')">
        <UBadge
          :label="user?.role"
          color="neutral"
          variant="subtle"
        />
      </UFormField>

      <div class="flex gap-2 pt-2">
        <UButton
          type="submit"
          :label="t('app.profile.save')"
          color="neutral"
          :loading="saving"
        />
        <UButton
          :label="t('app.profile.changePassword')"
          color="neutral"
          variant="outline"
          @click="passwordOpen = true"
        />
      </div>
    </UForm>

    <ConfirmDialog
      v-model:open="confirmOpen"
      :title="t('app.profile.usernameConfirm.title')"
      :description="affectedShelves === 1
        ? t('app.profile.usernameConfirm.descriptionOne', { username: pendingProfile?.username ?? '' })
        : t('app.profile.usernameConfirm.descriptionMany', { count: affectedShelves, username: pendingProfile?.username ?? '' })"
      :confirm-label="t('app.profile.usernameConfirm.confirm')"
      color="primary"
      :loading="saving"
      @confirm="confirmRename"
    />

    <UserPasswordDialog
      v-model:open="passwordOpen"
      :user="user ?? undefined"
    />
  </div>
</template>
