<script setup lang="ts">
const { t } = useI18n()
</script>

<template>
  <div class="min-h-screen flex flex-col bg-[#f5f5f4] text-[#1c274c]">
    <main class="flex-1 flex items-start justify-center px-4 py-12">
      <slot />
    </main>

    <footer class="py-6 flex justify-center">
      <a href="/" class="text-xs opacity-60 hover:opacity-100 transition-opacity">
        {{ t('linkpage.poweredBy') }}
      </a>
    </footer>
  </div>
</template>
