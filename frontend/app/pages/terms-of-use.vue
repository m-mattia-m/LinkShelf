<script setup lang="ts">
import type {SettingPageBody} from "~~/api";

const websiteSettings = useState('settings') as unknown as SettingPageBody

definePageMeta({
  layout: 'landingpage',
})
</script>

<template>
  <div class="my-8 mx-4 sm:mx-6 lg:mx-8">
    <MDC :value="websiteSettings.termsOfUse" tag="article" />
  </div>
</template>

<style scoped>
@reference "tailwindcss";

h1, h2, h3, h4, h5, h6, p, span, div {
  @apply text-inherit;
}

h1, h2, h3, h4, h5, h6 {
  @apply font-bold;
}
</style>
