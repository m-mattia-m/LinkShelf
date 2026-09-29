<script setup lang="ts">
definePageMeta({
  layout: 'landingpage'
})

const { t } = useI18n()
</script>

<template>
  <div class="my-8 mx-4 sm:mx-6 lg:mx-8">
    <h1 class="text-4xl text-highlighted pb-4">
      {{ t('nav.cloud') }}
    </h1>
    <p>
      {{ t('cloud.notAvailable') }}
      <NuxtLink
        class="text-primary"
        to="https://github.com/sponsors/m-mattia-m"
      >{{ t('cloud.sponsor') }}</NuxtLink>
    </p>
  </div>
</template>

<style scoped>
</style>
