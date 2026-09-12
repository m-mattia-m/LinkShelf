<script setup lang="ts">
import * as v from 'valibot'
import type { FormSubmitEvent } from '@nuxt/ui'

definePageMeta({
  layout: false
})

const { t } = useI18n()
const route = useRoute()

const token = computed(() => (typeof route.query.token === 'string' ? route.query.token : null))

const schema = v.pipe(
  v.object({
    password: v.pipe(
      v.string('Password is required'),
      v.nonEmpty('Password is required'),
      v.minLength(8, 'Must be at least 8 characters')
    ),
    confirmPassword: v.pipe(v.string('Please confirm your password'), v.nonEmpty('Please confirm your password'))
  }),
  v.forward(
    v.partialCheck(
      [['password'], ['confirmPassword']],
      input => input.password === input.confirmPassword,
      'Passwords do not match'
    ),
    ['confirmPassword']
  )
)

type Schema = v.InferOutput<typeof schema>

const fields = [
  { name: 'password', type: 'password' as const, label: t('auth.setPassword.password'), required: true },
  { name: 'confirmPassword', type: 'password' as const, label: t('auth.setPassword.confirmPassword'), required: true }
]

const loading = ref(false)
const errorMessage = ref<string | null>(null)
const success = ref(false)

async function onSubmit(payload: FormSubmitEvent<Schema>) {
  if (!token.value) return
  loading.value = true
  errorMessage.value = null
  try {
    const api = useApi()
    await api.auth.postSetPassword({ setPasswordRequest: { token: token.value, newPassword: payload.data.password } })
    success.value = true
  } catch (err) {
    const result = await parseApiError(err)
    errorMessage.value = result.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4">
    <div class="w-full max-w-sm flex flex-col gap-4">
      <template v-if="!token">
        <div class="flex flex-col items-center gap-4 text-center py-8">
          <UIcon
            name="i-lucide-circle-alert"
            class="size-12 text-error"
          />
          <p class="text-muted">
            {{ t('auth.setPassword.missingToken') }}
          </p>
          <ULink
            to="/auth/sign-in"
            class="text-primary font-medium"
          >{{ t('auth.setPassword.backToSignIn') }}</ULink>
        </div>
      </template>

      <template v-else-if="success">
        <div class="flex flex-col items-center gap-4 text-center py-8">
          <UIcon
            name="i-lucide-circle-check"
            class="size-12 text-success"
          />
          <h1 class="text-xl text-highlighted">
            {{ t('auth.setPassword.success.title') }}
          </h1>
          <ULink
            to="/auth/sign-in"
            class="text-primary font-medium"
          >{{ t('auth.setPassword.success.signInLink') }}</ULink>
        </div>
      </template>

      <template v-else>
        <UAlert
          v-if="errorMessage"
          color="error"
          variant="subtle"
          icon="i-lucide-circle-alert"
          :title="t('auth.setPassword.failedTitle')"
          :description="errorMessage"
          :close="{ onClick: () => (errorMessage = null) }"
        />

        <UAuthForm
          :schema="schema"
          :fields="fields"
          :title="t('auth.setPassword.title')"
          :description="t('auth.setPassword.description')"
          :submit="{ label: t('auth.setPassword.submit'), loading, block: true }"
          @submit="onSubmit"
        />
      </template>
    </div>
  </div>
</template>
