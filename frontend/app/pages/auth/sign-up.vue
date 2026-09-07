<script setup lang="ts">
import * as v from 'valibot'
import type { FormSubmitEvent } from '@nuxt/ui'

definePageMeta({
  layout: false
})

const { t } = useI18n()
const router = useRouter()
const authStore = useAuthStore()

const schema = v.object({
  firstName: v.pipe(v.string('First name is required'), v.nonEmpty('First name is required')),
  lastName: v.pipe(v.string('Last name is required'), v.nonEmpty('Last name is required')),
  email: v.pipe(
    v.string('Email is required'),
    v.nonEmpty('Email is required'),
    v.email('Please enter a valid email')
  ),
  password: v.pipe(
    v.string('Password is required'),
    v.nonEmpty('Password is required'),
    v.minLength(8, 'Must be at least 8 characters')
  )
})

type Schema = v.InferOutput<typeof schema>

const fields = [
  { name: 'firstName', type: 'text' as const, label: t('auth.signUp.firstName'), required: true },
  { name: 'lastName', type: 'text' as const, label: t('auth.signUp.lastName'), required: true },
  { name: 'email', type: 'text' as const, label: t('auth.signUp.email'), required: true },
  { name: 'password', type: 'password' as const, label: t('auth.signUp.password'), required: true }
]

const loading = ref(false)
const errorMessage = ref<string | null>(null)

async function onSubmit(payload: FormSubmitEvent<Schema>) {
  loading.value = true
  errorMessage.value = null
  try {
    await authStore.register({
      email: payload.data.email,
      firstName: payload.data.firstName,
      lastName: payload.data.lastName,
      password: payload.data.password
    })
    const toast = useToast()
    toast.add({ title: t('auth.signUp.success'), color: 'success' })
    await router.push('/app')
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
      <UAlert
        v-if="errorMessage"
        color="error"
        variant="subtle"
        icon="i-lucide-circle-alert"
        :title="t('auth.signUp.failedTitle')"
        :description="errorMessage"
        :close="{ onClick: () => (errorMessage = null) }"
      />

      <UAuthForm
        :schema="schema"
        :fields="fields"
        :title="t('auth.signUp.title')"
        :submit="{ label: t('auth.signUp.submit'), loading, block: true }"
        @submit="onSubmit"
      >
        <template #footer>
          {{ t('auth.signUp.haveAccount') }}
          <ULink to="/auth/sign-in" class="text-primary font-medium">{{ t('auth.signUp.signInLink') }}</ULink>
        </template>
      </UAuthForm>
    </div>
  </div>
</template>
