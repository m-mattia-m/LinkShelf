<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import * as v from 'valibot'
import type { User } from '~~/api'
import { useUserStore } from '~/stores/user'

const props = defineProps<{
  user?: User
}>()

const open = defineModel<boolean>('open', { default: false })

const { t } = useI18n()
const userStore = useUserStore()
const saving = ref(false)

const form = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

watch(open, (isOpen) => {
  if (!isOpen) return
  form.oldPassword = ''
  form.newPassword = ''
  form.confirmPassword = ''
})

const schema = v.pipe(
  v.object({
    oldPassword: v.pipe(v.string(), v.nonEmpty(() => t('validation.required'))),
    newPassword: v.pipe(v.string(), v.nonEmpty(() => t('validation.required'))),
    confirmPassword: v.pipe(v.string(), v.nonEmpty(() => t('validation.required')))
  }),
  v.forward(
    v.check(data => data.newPassword === data.confirmPassword, () => t('validation.passwordsMismatch')),
    ['confirmPassword']
  )
)

const formRef = ref<{
  validate: () => Promise<unknown>
  setErrors: (errs: { name?: string, message: string }[]) => void
} | null>(null)

async function save(close: () => void) {
  if (!props.user) return

  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  saving.value = true
  try {
    await userStore.patchPassword(props.user.id, {
      oldPassword: form.oldPassword,
      newPassword: form.newPassword
    })
    const toast = useToast()
    toast.add({ title: t('app.password.updated'), color: 'success' })
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
    :title="t('app.profile.changePassword')"
    :ui="{ footer: 'justify-end' }"
  >
    <template #body>
      <UForm
        ref="formRef"
        :schema="schema"
        :state="form"
        class="flex flex-col gap-4"
      >
        <UFormField
          :label="t('app.password.current')"
          name="oldPassword"
          required
        >
          <UInput
            v-model="form.oldPassword"
            type="password"
            class="w-full"
          />
        </UFormField>

        <UFormField
          :label="t('auth.setPassword.password')"
          name="newPassword"
          required
        >
          <UInput
            v-model="form.newPassword"
            type="password"
            class="w-full"
          />
        </UFormField>

        <UFormField
          :label="t('auth.resetPassword.confirmPassword')"
          name="confirmPassword"
          required
        >
          <UInput
            v-model="form.confirmPassword"
            type="password"
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
