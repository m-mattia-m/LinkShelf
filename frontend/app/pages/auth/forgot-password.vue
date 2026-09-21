<script setup lang="ts">
import * as v from 'valibot'
import type { FormSubmitEvent } from '@nuxt/ui'
import type { SettingPageBody } from '~~/api'

definePageMeta({
  layout: false
})

const { t } = useI18n()
const websiteSettings = useState('settings') as unknown as Ref<SettingPageBody | null>

// Only an explicit "off" hides the form: with no settings loaded (backend
// briefly unreachable) the request itself will report what is wrong.
const disabled = computed(() => websiteSettings.value?.passwordResetEnabled === false)

const schema = v.object({
  email: v.pipe(
    v.string('Email is required'),
    v.nonEmpty('Email is required'),
    v.email('Please enter a valid email')
  )
})

type Schema = v.InferOutput<typeof schema>

const fields = [
  { name: 'email', type: 'text' as const, label: t('auth.forgotPassword.email'), required: true }
]

const loading = ref(false)
const errorMessage = ref<string | null>(null)
const sentTo = ref<string | null>(null)

async function onSubmit(payload: FormSubmitEvent<Schema>) {
  loading.value = true
  errorMessage.value = null
  try {
    const api = useApi()
    await api.auth.postForgotPassword({ forgotPasswordRequest: { email: payload.data.email } })
    // The backend answers the same way for every address, so this says
    // nothing about whether an account exists - and neither may the page.
    sentTo.value = payload.data.email
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
      <div
        v-if="disabled"
        class="flex flex-col items-center gap-4 text-center py-8"
      >
        <UIcon
          name="i-lucide-lock"
          class="size-12 text-muted"
        />
        <p class="text-muted">
          {{ t('auth.forgotPassword.disabled') }}
        </p>
        <ULink
          to="/auth/sign-in"
          class="text-primary font-medium"
        >{{ t('auth.forgotPassword.backToSignIn') }}</ULink>
      </div>

      <div
        v-else-if="sentTo"
        class="flex flex-col items-center gap-4 text-center py-8"
      >
        <UIcon
          name="i-lucide-mail-check"
          class="size-12 text-primary"
        />
        <h1 class="text-xl text-highlighted">
          {{ t('auth.forgotPassword.sent.title') }}
        </h1>
        <p class="text-muted">
          {{ t('auth.forgotPassword.sent.description', { email: sentTo }) }}
        </p>
        <ULink
          to="/auth/sign-in"
          class="text-primary font-medium"
        >{{ t('auth.forgotPassword.backToSignIn') }}</ULink>
      </div>

      <template v-else>
        <UAlert
          v-if="errorMessage"
          color="error"
          variant="subtle"
          icon="i-lucide-circle-alert"
          :title="t('auth.forgotPassword.failedTitle')"
          :description="errorMessage"
          :close="{ onClick: () => (errorMessage = null) }"
        />

        <UAuthForm
          :schema="schema"
          :fields="fields"
          :title="t('auth.forgotPassword.title')"
          :description="t('auth.forgotPassword.description')"
          :submit="{ label: t('auth.forgotPassword.submit'), loading, block: true }"
          @submit="onSubmit"
        >
          <template #footer>
            <ULink
              to="/auth/sign-in"
              class="text-primary font-medium"
            >{{ t('auth.forgotPassword.backToSignIn') }}</ULink>
          </template>
        </UAuthForm>
      </template>
    </div>
  </div>
</template>
