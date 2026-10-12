<script setup lang="ts">
definePageMeta({
  layout: false
})

const { t } = useI18n()
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4">
    <div class="flex flex-col items-center gap-4 text-center max-w-md">
      <UIcon
        name="i-lucide-mail-check"
        class="size-12 text-primary"
      />
      <h1 class="text-xl text-highlighted">
        {{ t('auth.emailPending.title') }}
      </h1>
      <p class="text-muted">
        {{ t('auth.emailPending.description') }}
      </p>
      <UButton
        to="/auth/sign-in"
        size="lg"
      >
        {{ t('auth.emailPending.signIn') }}
      </UButton>
      <p class="text-sm text-muted">
        {{ t('auth.emailPending.hint') }}
      </p>
    </div>
  </div>
</template>
