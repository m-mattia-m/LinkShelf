<script setup lang="ts">
import * as v from 'valibot'
import type { FormSubmitEvent } from '@nuxt/ui'

definePageMeta({
  layout: false
})

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const schema = v.object({
  email: v.pipe(
    v.string('Email is required'),
    v.nonEmpty('Email is required'),
    v.email('Please enter a valid email')
  ),
  password: v.pipe(v.string('Password is required'), v.nonEmpty('Password is required'))
})

type Schema = v.InferOutput<typeof schema>

const fields = [
  { name: 'email', type: 'text' as const, label: t('auth.signIn.email'), required: true },
  { name: 'password', type: 'password' as const, label: t('auth.signIn.password'), required: true }
]

const loading = ref(false)
const errorMessage = ref<string | null>(null)

async function startOidc() {
  try {
    await authStore.startOidcLogin()
  } catch (err) {
    await handleApiError(err)
  }
}

const providers = [
  {
    label: t('auth.signIn.sso'),
    icon: 'i-lucide-key-round',
    color: 'neutral' as const,
    variant: 'subtle' as const,
    block: true,
    onClick: startOidc
  }
]

async function onSubmit(payload: FormSubmitEvent<Schema>) {
  loading.value = true
  errorMessage.value = null
  try {
    await authStore.login(payload.data.email, payload.data.password)
    const toast = useToast()
    toast.add({ title: t('auth.signIn.success'), color: 'success' })
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/app'
    await router.push(redirect)
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
        :title="t('auth.signIn.failedTitle')"
        :description="errorMessage"
        :close="{ onClick: () => (errorMessage = null) }"
      />

      <UAuthForm
        :schema="schema"
        :fields="fields"
        :providers="providers"
        :title="t('auth.signIn.title')"
        :submit="{ label: t('auth.signIn.submit'), loading, block: true }"
        :separator="t('auth.signIn.or')"
        @submit="onSubmit"
      >
        <template #footer>
          {{ t('auth.signIn.noAccount') }}
          <ULink to="/auth/sign-up" class="text-primary font-medium">{{ t('auth.signIn.signUpLink') }}</ULink>
        </template>
      </UAuthForm>
    </div>
  </div>
</template>
