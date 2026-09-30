<script setup lang="ts">
import { ResponseError } from '~~/api'

definePageMeta({
  layout: false
})

const { t } = useI18n()
const route = useRoute()
const authStore = useAuthStore()

// 'taken': another account registered the address after the change was requested.
const status = ref<'confirming' | 'success' | 'taken' | 'failed'>('confirming')

onMounted(async () => {
  const token = typeof route.query.token === 'string' ? route.query.token : null
  if (!token) {
    status.value = 'failed'
    return
  }

  try {
    const api = useApi()
    await api.auth.postConfirmEmailChange({ verifyEmailRequest: { token } })
    status.value = 'success'
    // The link is usually opened while signed in: show the new email right away.
    if (authStore.isAuthenticated) await authStore.fetchUser().catch(() => {})
  } catch (err) {
    status.value = err instanceof ResponseError && err.response.status === 409 ? 'taken' : 'failed'
  }
})

const continueTo = computed(() => authStore.isAuthenticated ? '/app/profile' : '/auth/sign-in')
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4">
    <div class="flex flex-col items-center gap-4 text-center">
      <template v-if="status === 'confirming'">
        <UIcon
          name="i-lucide-loader-circle"
          class="size-8 animate-spin text-primary"
        />
        <p class="text-muted">
          {{ t('auth.confirmEmail.confirming') }}
        </p>
      </template>
      <template v-else-if="status === 'success'">
        <UIcon
          name="i-lucide-circle-check"
          class="size-12 text-success"
        />
        <h1 class="text-xl text-highlighted">
          {{ t('auth.confirmEmail.success.title') }}
        </h1>
        <p class="text-muted">
          {{ t('auth.confirmEmail.success.description') }}
        </p>
        <ULink
          :to="continueTo"
          class="text-primary font-medium"
        >{{ t('auth.confirmEmail.continue') }}</ULink>
      </template>
      <template v-else>
        <UIcon
          name="i-lucide-circle-alert"
          class="size-12 text-error"
        />
        <h1 class="text-xl text-highlighted">
          {{ status === 'taken' ? t('auth.confirmEmail.taken.title') : t('auth.confirmEmail.failed.title') }}
        </h1>
        <p class="text-muted">
          {{ status === 'taken' ? t('auth.confirmEmail.taken.description') : t('auth.confirmEmail.failed.description') }}
        </p>
        <ULink
          :to="continueTo"
          class="text-primary font-medium"
        >{{ t('auth.confirmEmail.continue') }}</ULink>
      </template>
    </div>
  </div>
</template>
