<script setup lang="ts">
// /<username>/<path>, used while app.userBasedPaths is on. With the setting
// off the backend answers 404 here, which shows the usual "not found" page.
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
