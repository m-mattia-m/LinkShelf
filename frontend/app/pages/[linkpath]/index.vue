<script setup lang="ts">
import type { Link, PublicShelf, Section } from '~~/api'

definePageMeta({
  layout: 'links'
})

const route = useRoute()
const { t } = useI18n()
const linkpath = route.params.linkpath as string

const loading = ref(true)
const notFound = ref(false)
const shelf = ref<PublicShelf>()
const sections = ref<Section[]>([])
const links = ref<Link[]>([])

const linksBySectionId = computed(() => {
  const map = new Map<string, Link[]>()
  for (const link of links.value) {
    const list = map.get(link.sectionId) ?? []
    list.push(link)
    map.set(link.sectionId, list)
  }
  return map
})

const visibleSections = computed(() =>
  sections.value.filter((section) => (linksBySectionId.value.get(section.id)?.length ?? 0) > 0)
)

const hasAnyLinks = computed(() => links.value.length > 0)

async function load() {
  loading.value = true
  notFound.value = false

  try {
    const api = useApi()
    const resolvedShelf = await api.shelf.getPublicShelfByPath({ path: linkpath })
    shelf.value = resolvedShelf

    const [sectionList, linkList] = await Promise.all([
      api.section.getSections({ shelfId: resolvedShelf.id }),
      api.link.getLinks({ shelfId: resolvedShelf.id })
    ])
    sections.value = sectionList ?? []
    links.value = linkList ?? []
  } catch {
    notFound.value = true
  } finally {
    loading.value = false
  }
}

onMounted(load)

useSeoMeta({
  title: () => shelf.value?.title,
  description: () => shelf.value?.description
})
</script>

<template>
  <div class="w-full max-w-md flex flex-col items-center gap-6">
    <template v-if="loading">
      <USkeleton class="size-16 rounded-full" />
      <USkeleton class="h-5 w-40 rounded" />
      <USkeleton class="h-4 w-56 rounded" />
      <USkeleton class="h-12 w-full rounded-xl" />
      <USkeleton class="h-12 w-full rounded-xl" />
    </template>

    <template v-else-if="notFound">
      <p class="text-lg font-medium">{{ t('linkpage.notFound.title') }}</p>
      <p class="text-sm opacity-70 text-center">{{ t('linkpage.notFound.description') }}</p>
      <NuxtLink to="/" class="text-sm underline">{{ t('linkpage.notFound.backHome') }}</NuxtLink>
    </template>

    <template v-else-if="shelf">
      <UIcon v-if="shelf.icon" :name="shelf.icon" class="size-12" />
      <h1 class="text-xl font-semibold text-center">{{ shelf.title }}</h1>
      <p v-if="shelf.description" class="text-sm text-center opacity-70">{{ shelf.description }}</p>

      <template v-if="hasAnyLinks">
        <div v-for="section in visibleSections" :key="section.id" class="w-full flex flex-col gap-3">
          <h2 class="text-xs font-semibold uppercase tracking-wide opacity-60">{{ section.title }}</h2>

          <a
            v-for="link in linksBySectionId.get(section.id)"
            :key="link.id"
            :href="link.link"
            target="_blank"
            rel="noopener noreferrer"
            class="flex items-center gap-3 rounded-xl px-4 py-3 font-medium text-white shadow-sm transition-transform hover:scale-[1.02]"
            :style="{ backgroundColor: link.color }"
          >
            <UIcon v-if="link.icon" :name="link.icon" class="size-5 shrink-0" />
            <span class="truncate">{{ link.title }}</span>
          </a>
        </div>
      </template>

      <p v-else class="text-sm opacity-60">{{ t('linkpage.empty') }}</p>
    </template>
  </div>
</template>
