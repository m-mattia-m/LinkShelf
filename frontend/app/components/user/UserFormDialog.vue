<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import * as v from 'valibot'
import type { SettingPageBody, User } from '~~/api'
import { useUserStore } from '~/stores/user'
import { useShelfStore } from '~/stores/shelf'

const { t } = useI18n()
const websiteSettings = useState('settings') as unknown as Ref<SettingPageBody | null>

const props = withDefaults(defineProps<{
  mode?: 'create' | 'edit'
  user?: User
}>(), {
  mode: 'create'
})

const emit = defineEmits<{
  (e: 'saved'): void
}>()

const open = defineModel<boolean>('open', { default: false })

const userStore = useUserStore()
const shelfStore = useShelfStore()
const saving = ref(false)
const userBasedPaths = computed(() => websiteSettings.value?.userBasedPaths ?? false)

const roleOptions = [
  { label: 'User', value: 'user' },
  { label: 'Admin', value: 'admin' }
]

const form = reactive({
  firstName: props.user?.firstName ?? '',
  lastName: props.user?.lastName ?? '',
  username: props.user?.username ?? '',
  email: props.user?.email ?? '',
  password: '',
  role: (props.user?.role ?? 'user') as 'user' | 'admin'
})

watch(open, async (isOpen) => {
  if (!isOpen) return
  form.firstName = props.user?.firstName ?? ''
  form.lastName = props.user?.lastName ?? ''
  form.username = props.user?.username ?? ''
  form.email = props.user?.email ?? ''
  form.password = ''
  form.role = (props.user?.role ?? 'user') as 'user' | 'admin'

  if (props.mode === 'edit' && userBasedPaths.value && !shelfStore.loaded) {
    try {
      await shelfStore.fetch()
    } catch {
      // Only the rename warning depends on this; saving works without it.
    }
  }
})

// Password can be left blank on create only when email verification is on -
// the account is then created "invited" and the emailed link is the only
// way to set one (matches the backend's own requirement exactly).
const canInviteWithoutPassword = computed(() => websiteSettings.value?.emailVerificationEnabled ?? false)

const createSchema = computed(() => v.object({
  firstName: v.pipe(v.string(), v.nonEmpty('Required')),
  lastName: v.pipe(v.string(), v.nonEmpty('Required')),
  username: usernameSchema(),
  email: v.pipe(v.string(), v.nonEmpty('Required'), v.email('Must be a valid email address')),
  password: canInviteWithoutPassword.value ? v.string() : v.pipe(v.string(), v.nonEmpty('Required')),
  role: v.picklist(['user', 'admin'])
}))

const editSchema = v.object({
  firstName: v.pipe(v.string(), v.nonEmpty('Required')),
  lastName: v.pipe(v.string(), v.nonEmpty('Required')),
  username: usernameSchema(),
  email: v.pipe(v.string(), v.nonEmpty('Required'), v.email('Must be a valid email address')),
  password: v.string(),
  role: v.picklist(['user', 'admin'])
})

// Renaming a user changes the URL of every shelf of theirs that has a path.
// An admin sees all shelves, so they can be counted here.
const affectedShelves = computed(() =>
  props.mode === 'edit' && props.user
    ? shelfStore.shelves.filter(shelf => shelf.userId === props.user!.id && shelf.path).length
    : 0
)
const usernameChangeWarning = computed(() => {
  const renamed = props.mode === 'edit' && props.user && form.username !== props.user.username
  if (!renamed || !userBasedPaths.value || affectedShelves.value === 0) return undefined
  return affectedShelves.value === 1
    ? t('app.settings.users.form.usernameChangeWarningOne')
    : t('app.settings.users.form.usernameChangeWarningMany', { count: affectedShelves.value })
})

const schema = computed(() => (props.mode === 'create' ? createSchema.value : editSchema))

const formRef = ref<{
  validate: () => Promise<unknown>
  setErrors: (errs: { name?: string, message: string }[]) => void
} | null>(null)

async function save(close: () => void) {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  saving.value = true
  try {
    if (props.mode === 'edit' && props.user) {
      await userStore.update(props.user.id, {
        firstName: form.firstName,
        lastName: form.lastName,
        username: form.username,
        email: form.email,
        role: form.role
      })
    } else {
      await userStore.create({
        firstName: form.firstName,
        lastName: form.lastName,
        username: form.username,
        email: form.email,
        password: form.password,
        role: form.role
      })
    }

    emit('saved')
    close()
  } catch (err) {
    await handleApiError(err, formRef.value)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <UModal
    v-model:open="open"
    :title="mode === 'edit' ? 'Edit user' : 'New user'"
    :ui="{ footer: 'justify-end' }"
  >
    <UButton
      v-if="mode === 'create'"
      icon="i-lucide-plus"
      label="New"
    />

    <template #body>
      <UForm
        ref="formRef"
        :schema="schema"
        :state="form"
        class="flex flex-col gap-4"
      >
        <UFormField
          label="First name"
          name="firstName"
          required
        >
          <UInput
            v-model="form.firstName"
            class="w-full"
          />
        </UFormField>

        <UFormField
          label="Last name"
          name="lastName"
          required
        >
          <UInput
            v-model="form.lastName"
            class="w-full"
          />
        </UFormField>

        <UFormField
          :label="t('app.settings.users.form.username')"
          name="username"
          required
        >
          <UInput
            v-model="form.username"
            class="w-full"
          />
        </UFormField>

        <UAlert
          v-if="usernameChangeWarning"
          color="warning"
          variant="subtle"
          icon="i-lucide-triangle-alert"
          :description="usernameChangeWarning"
        />

        <UFormField
          label="Email"
          name="email"
          required
        >
          <UInput
            v-model="form.email"
            type="email"
            class="w-full"
          />
        </UFormField>

        <UFormField
          v-if="mode === 'create'"
          :label="t('app.settings.users.form.password')"
          name="password"
          :required="!canInviteWithoutPassword"
          :hint="canInviteWithoutPassword ? t('app.settings.users.form.passwordInviteHint') : undefined"
        >
          <UInput
            v-model="form.password"
            type="password"
            class="w-full"
          />
        </UFormField>

        <UFormField
          label="Role"
          name="role"
          required
        >
          <USelect
            v-model="form.role"
            :items="roleOptions"
            value-key="value"
            class="w-full"
          />
        </UFormField>
      </UForm>
    </template>

    <template #footer="{ close }">
      <UButton
        label="Cancel"
        color="neutral"
        variant="outline"
        @click="close"
      />
      <UButton
        label="Submit"
        color="neutral"
        :loading="saving"
        @click="save(close)"
      />
    </template>
  </UModal>
</template>
