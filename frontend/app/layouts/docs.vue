<script setup lang="ts">
import type { AccordionItem } from '@nuxt/ui'

const route = useRoute()

const {data: page} = await useAsyncData(route.path, () => queryCollection('docs').path(route.path).first())
if (!page.value) {
  throw createError({statusCode: 404, statusMessage: 'Page not found', fatal: true})
}

const items = [
  {
    label: 'Navigate in the docs',
    slot: 'menu',
  }
] satisfies AccordionItem[]

</script>

<template>
  <AppLayout>

    <div class="block lg:hidden">
    <UAccordion :items="items" class="border-b border-gray-200 px-6">
      <template #menu>
        <DocsSidebar/>
      </template>
    </UAccordion>
    </div>

    <div class="min-h-screen flex">
      <div class="hidden lg:block w-56 min-w-56 max-w-56 border-r border-gray-200 bg-white">
        <DocsSidebar/>
      </div>

      <slot/>

      <div
        id="toc"
        v-if="page?.body?.toc?.links?.length"
        class="hidden lg:block w-56 min-w-56 max-w-56 border-l border-gray-200 bg-white -pl-4"
      >
        <UContentToc
          :links="page.body.toc.links"
          class="backdrop-blur-none mx-0 sm:mx-0 pt-0"
        />
      </div>

    </div>
  </AppLayout>
</template>

<style scoped lang="css">
</style>
