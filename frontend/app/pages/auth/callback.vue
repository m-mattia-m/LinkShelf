<script setup lang="ts">
definePageMeta({
  layout: false
})

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const failed = ref(false)

onMounted(async () => {
  const code = typeof route.query.code === 'string' ? route.query.code : null
  const state = typeof route.query.state === 'string' ? route.query.state : null
  const expectedState = authStore.consumeStoredOidcState()

  if (!code || !state || !expectedState || state !== expectedState) {
    failed.value = true
    return
  }

  try {
    await authStore.completeOidcLogin(code, state)
    await router.push('/app')
  } catch (err) {
    await handleApiError(err)
    failed.value = true
  }
})
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4">
    <div class="flex flex-col items-center gap-4 text-center">
      <template v-if="!failed">
        <UIcon name="i-lucide-loader-circle" class="size-8 animate-spin text-primary" />
        <p class="text-muted">{{ t('auth.callback.completing') }}</p>
      </template>
      <template v-else>
        <p class="text-muted">{{ t('auth.callback.failed') }}</p>
        <ULink to="/auth/sign-in" class="text-primary font-medium">{{ t('auth.callback.backToSignIn') }}</ULink>
      </template>
    </div>
  </div>
</template>
