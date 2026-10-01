<script setup lang="ts">
// /<username>/<path>, used while app.userBasedPaths is on.
definePageMeta({
  layout: 'links'
})

const route = useRoute()
</script>

<template>
  <ShelfPublicView
    :path="route.params.linkpath as string"
    :username="route.params.username as string"
  />
</template>
