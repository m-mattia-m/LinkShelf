<script setup lang="ts">
import type { Link, PublicShelf, Section } from '~~/api'

const props = defineProps<{
  // The shelf is found by its path, or - when it is served on a domain of its
  // own - by that domain, in which case there is no path.
  path?: string
  // Set for /<username>/<path> URLs (app.userBasedPaths), absent for /<path>.
  username?: string
  // Set when the page is being viewed on the shelf's own domain.
  domain?: string
}>()

const { t } = useI18n()

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
  sections.value.filter(section => (linksBySectionId.value.get(section.id)?.length ?? 0) > 0)
)

const hasAnyLinks = computed(() => links.value.length > 0)

// See layouts/links.vue - shared with it via the same useState key since a
// layout can't receive props from its page and these need to live on the
// layout's own root element for its background/text color to react to them.
const themeVars = useState<Record<string, string>>('public-shelf-theme-vars', () => ({}))

// A link's own color always has a DB-level default of "#000000" rather than
// being genuinely unset, so there's no way to tell "user picked black" apart
// from "user never touched this." Treating that default as "no override"
// lets a theme's --shelf-link-bg show through for links nobody has
// customized, rather than every untouched link staying hardcoded black
// regardless of the shelf's theme.
function linkBackgroundStyle(link: Link): Record<string, string> {
  if (link.color && link.color.toLowerCase() !== '#000000') {
    return { backgroundColor: link.color }
  }
  return {}
}

async function load() {
  loading.value = true
  notFound.value = false

  try {
    const api = useApi()
    const resolvedShelf = props.domain
      ? await api.shelf.getPublicShelfByDomain({ domain: props.domain })
      : props.username
        ? await api.shelf.getPublicShelfByUsernameAndPath({ username: props.username, path: props.path ?? '' })
        : await api.shelf.getPublicShelfByPath({ path: props.path ?? '' })
    shelf.value = resolvedShelf
    themeVars.value = resolvedShelf.theme ?? {}

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

// Nuxt reuses this component when only the route params change.
watch(() => [props.username, props.path, props.domain], load)

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
      <p class="text-lg font-medium">
        {{ t('linkpage.notFound.title') }}
      </p>
      <p class="text-sm opacity-70 text-center">
        {{ t('linkpage.notFound.description') }}
      </p>
      <NuxtLink
        to="/"
        class="text-sm underline"
      >{{ t('linkpage.notFound.backHome') }}</NuxtLink>
    </template>

    <template v-else-if="shelf">
      <UIcon
        v-if="shelf.icon"
        :name="shelf.icon"
        class="size-12"
      />
      <h1 class="text-xl font-semibold text-center">
        {{ shelf.title }}
      </h1>
      <p
        v-if="shelf.description"
        class="text-sm text-center opacity-70"
      >
        {{ shelf.description }}
      </p>

      <template v-if="hasAnyLinks">
        <div
          v-for="section in visibleSections"
          :key="section.id"
          class="w-full flex flex-col gap-3"
        >
          <h2 class="text-xs font-semibold uppercase tracking-wide opacity-60">
            {{ section.title }}
          </h2>

          <a
            v-for="link in linksBySectionId.get(section.id)"
            :key="link.id"
            :href="link.link"
            target="_blank"
            rel="noopener noreferrer"
            class="flex items-center gap-3 rounded-[var(--shelf-link-radius,0.75rem)] px-4 py-3 font-medium text-[var(--shelf-link-text,white)] shadow-sm transition-transform hover:scale-[1.02] bg-[var(--shelf-link-bg,#000)]"
            :style="linkBackgroundStyle(link)"
          >
            <UIcon
              v-if="link.icon"
              :name="link.icon"
              class="size-5 shrink-0"
            />
            <span class="truncate">{{ link.title }}</span>
          </a>
        </div>
      </template>

      <p
        v-else
        class="text-sm opacity-60"
      >
        {{ t('linkpage.empty') }}
      </p>
    </template>
  </div>
</template>
