<script setup lang="ts">
const route = useRoute()

onMounted(() => {
  console.log(route.params["linkpath"])
})

definePageMeta({
  layout: 'links',
})

</script>

<template>
  <p>{{ route.params.slug }}</p>
</template>

<style scoped>

</style>
