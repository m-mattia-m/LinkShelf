<script setup lang="ts">
const route = useRoute()

const docPath = computed(() => {
  if (Array.isArray(route.params.slug)) {
    return `/docs/${route.params.slug.join('/')}`
  }
  return '/docs'
})

const { data: page } = await useAsyncData(
  () => `docs:${route.path}`,
  () => queryCollection('docs').where('path', '=', docPath.value).path(route.path).first()
)

if (!page.value) {
  throw createError({ statusCode: 404, statusMessage: 'Page not found' })
}

definePageMeta({
  layout: 'docs',
})
</script>

<template>
  <UContainer class="flex gap-10 py-8">
    <UPage v-if="page">
      <UPageBody class="mt-0">
        <article v-if="page" class="flex-1 prose max-w-none">
          <UPageHeader :title="page.title" class="pt-0" />
          <ContentRenderer v-if="page.body" :value="page"/>
        </article>
      </UPageBody>
    </UPage>
  </UContainer>
</template>
