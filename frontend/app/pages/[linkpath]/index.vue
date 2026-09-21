<script setup lang="ts">
definePageMeta({
  layout: 'links'
})

const route = useRoute()
</script>

<template>
  <ShelfPublicView :path="route.params.linkpath as string" />
</template>
