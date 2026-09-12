<script setup lang="ts">
definePageMeta({
  layout: false
})

const { t } = useI18n()
const route = useRoute()

const status = ref<'verifying' | 'success' | 'failed'>('verifying')

onMounted(async () => {
  const token = typeof route.query.token === 'string' ? route.query.token : null
  if (!token) {
    status.value = 'failed'
    return
  }

  try {
    const api = useApi()
    await api.auth.postVerifyEmail({ verifyEmailRequest: { token } })
    status.value = 'success'
  } catch {
    status.value = 'failed'
  }
})
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4">
    <div class="flex flex-col items-center gap-4 text-center">
      <template v-if="status === 'verifying'">
        <UIcon
          name="i-lucide-loader-circle"
          class="size-8 animate-spin text-primary"
        />
        <p class="text-muted">
          {{ t('auth.verifyEmail.verifying') }}
        </p>
      </template>
      <template v-else-if="status === 'success'">
        <UIcon
          name="i-lucide-circle-check"
          class="size-12 text-success"
        />
        <h1 class="text-xl text-highlighted">
          {{ t('auth.verifyEmail.success.title') }}
        </h1>
        <ULink
          to="/auth/sign-in"
          class="text-primary font-medium"
        >{{ t('auth.verifyEmail.success.signInLink') }}</ULink>
      </template>
      <template v-else>
        <UIcon
          name="i-lucide-circle-alert"
          class="size-12 text-error"
        />
        <h1 class="text-xl text-highlighted">
          {{ t('auth.verifyEmail.failed.title') }}
        </h1>
        <p class="text-muted">
          {{ t('auth.verifyEmail.failed.description') }}
        </p>
        <ULink
          to="/auth/sign-in"
          class="text-primary font-medium"
        >{{ t('auth.verifyEmail.failed.backToSignIn') }}</ULink>
      </template>
    </div>
  </div>
</template>
