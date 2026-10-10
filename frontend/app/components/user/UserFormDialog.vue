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

const roleOptions = computed(() => [
  { label: t('app.settings.users.roles.user'), value: 'user' },
  { label: t('app.settings.users.roles.admin'), value: 'admin' }
])

const form = reactive({
  firstName: props.user?.firstName ?? '',
  lastName: props.user?.lastName ?? '',
  username: props.user?.username ?? '',
  email: props.user?.email ?? '',
  password: '',
  role: (props.user?.role ?? 'user') as 'user' | 'admin',
  maxShelves: props.user?.maxShelves ?? null as number | null
})

watch(open, async (isOpen) => {
  if (!isOpen) return
  form.firstName = props.user?.firstName ?? ''
  form.lastName = props.user?.lastName ?? ''
  form.username = props.user?.username ?? ''
  form.email = props.user?.email ?? ''
  form.password = ''
  form.role = (props.user?.role ?? 'user') as 'user' | 'admin'
  form.maxShelves = props.user?.maxShelves ?? null

  if (props.mode === 'edit' && userBasedPaths.value && !shelfStore.loaded) {
    try {
      await shelfStore.fetch()
    } catch {
      // Only the rename warning depends on this; saving works without it.
    }
  }
})

// The password is optional on create when email verification sends an invite.
const canInviteWithoutPassword = computed(() => websiteSettings.value?.emailVerificationEnabled ?? false)

const createSchema = computed(() => v.object({
  firstName: v.pipe(v.string(), v.nonEmpty(t('validation.required'))),
  lastName: v.pipe(v.string(), v.nonEmpty(t('validation.required'))),
  username: usernameSchema(),
  email: emailSchema(),
  password: canInviteWithoutPassword.value ? v.string() : v.pipe(v.string(), v.nonEmpty(t('validation.required'))),
  role: v.picklist(['user', 'admin'])
}))

const editSchema = computed(() => v.object({
  firstName: v.pipe(v.string(), v.nonEmpty(t('validation.required'))),
  lastName: v.pipe(v.string(), v.nonEmpty(t('validation.required'))),
  username: usernameSchema(),
  email: emailSchema(),
  password: v.string(),
  role: v.picklist(['user', 'admin']),
  // Clearing the input yields undefined.
  maxShelves: v.nullish(v.pipe(v.number(), v.integer(), v.minValue(0)))
}))

// Renaming changes the URLs of the user's path shelves.
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

const schema = computed(() => (props.mode === 'create' ? createSchema.value : editSchema.value))

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
      const maxShelves = form.maxShelves ?? null
      if (maxShelves !== (props.user.maxShelves ?? null)) {
        await userStore.setMaxShelves(props.user.id, maxShelves)
      }
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
    :title="mode === 'edit' ? t('app.settings.users.edit') : t('app.settings.users.new')"
    :ui="{ footer: 'justify-end' }"
  >
    <UButton
      v-if="mode === 'create'"
      icon="i-lucide-plus"
      :label="t('common.new')"
    />

    <template #body>
      <UForm
        ref="formRef"
        :schema="schema"
        :state="form"
        class="flex flex-col gap-4"
      >
        <UFormField
          :label="t('app.settings.users.columns.firstName')"
          name="firstName"
          required
        >
          <UInput
            v-model="form.firstName"
            class="w-full"
          />
        </UFormField>

        <UFormField
          :label="t('app.settings.users.columns.lastName')"
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
          :label="t('app.settings.users.columns.email')"
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
          :label="t('app.profile.role')"
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

        <UFormField
          v-if="mode === 'edit'"
          :label="t('app.settings.users.form.maxShelves')"
          :hint="t('app.settings.users.form.maxShelvesHint')"
          name="maxShelves"
        >
          <UInputNumber
            v-model="form.maxShelves"
            :min="0"
            :increment="false"
            :decrement="false"
            class="w-full"
          />
        </UFormField>
      </UForm>
    </template>

    <template #footer="{ close }">
      <UButton
        :label="t('common.cancel')"
        color="neutral"
        variant="outline"
        @click="close"
      />
      <UButton
        :label="t('common.submit')"
        color="neutral"
        :loading="saving"
        @click="save(close)"
      />
    </template>
  </UModal>
</template>
